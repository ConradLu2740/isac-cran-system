package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"isac-cran-system/internal/algorithm/beamforming"
	"isac-cran-system/internal/algorithm/doa"
	"isac-cran-system/internal/config"
	"isac-cran-system/internal/device/irs"
	"isac-cran-system/internal/device/sensor"
	"isac-cran-system/internal/device/usrp"
	"isac-cran-system/internal/handler"
	"isac-cran-system/internal/middleware"
	"isac-cran-system/internal/repository/influxdb"
	"isac-cran-system/internal/repository/mysql"
	"isac-cran-system/internal/router"
	"isac-cran-system/internal/service"
	"isac-cran-system/pkg/logger"
	"isac-cran-system/pkg/pool"
	"isac-cran-system/pkg/queue"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/config.yaml", "config file path")
}

func main() {
	flag.Parse()

	if err := config.Init(configFile); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	cfg := config.Get()

	logCfg := &logger.Config{
		Level:      cfg.Log.Level,
		Format:     cfg.Log.Format,
		Output:     cfg.Log.Output,
		Filename:   cfg.Log.Filename,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
	}

	if err := logger.Init(logCfg); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting ISAC-CRAN System",
		zap.String("config", configFile),
	)

	gin.SetMode(cfg.Server.Mode)

	var db *mysql.DB
	var influxClient *influxdb.Client
	var err error

	db, err = mysql.NewDB(&cfg.MySQL)
	if err != nil {
		logger.Warn("MySQL connection failed, running without database", zap.Error(err))
	} else {
		defer db.Close()
		if err := db.AutoMigrate(); err != nil {
			logger.Warn("Auto migrate failed", zap.Error(err))
		}
		logger.Info("MySQL connected successfully")
	}

	influxClient, err = influxdb.NewClient(&cfg.InfluxDB)
	if err != nil {
		logger.Warn("InfluxDB connection failed, running without time-series database", zap.Error(err))
	} else {
		defer influxClient.Close()
		logger.Info("InfluxDB connected successfully")
	}

	irsDriverFactory := irs.NewDriverFactory()
	irsDriver, err := irsDriverFactory.Create(
		irs.DriverTypeSimulator,
		irs.WithElementCount(cfg.Device.IRS.ElementCount),
		irs.WithFrequencyBand(cfg.Device.IRS.FrequencyBand),
	)
	if err != nil {
		logger.Fatal("Failed to create IRS driver", zap.Error(err))
	}
	irsController := irs.NewController(irsDriver)

	usrpDriverFactory := usrp.NewDriverFactory()
	usrpDriver, err := usrpDriverFactory.Create(
		usrp.DriverTypeSimulator,
		usrp.WithSampleRate(cfg.Device.USRP.SampleRate),
		usrp.WithCenterFreq(cfg.Device.USRP.CenterFreq),
	)
	if err != nil {
		logger.Fatal("Failed to create USRP driver", zap.Error(err))
	}
	usrpReceiver := usrp.NewReceiver(usrpDriver, cfg.Device.USRP.SampleRate, cfg.Device.USRP.CenterFreq)

	sensorDriverFactory := sensor.NewDriverFactory()
	sensorDriver, err := sensorDriverFactory.Create(sensor.DriverTypeSimulator)
	if err != nil {
		logger.Fatal("Failed to create sensor driver", zap.Error(err))
	}
	sensorCollector := sensor.NewCollector(sensorDriver, cfg.Device.Sensor.CollectionInterval)

	ctx := context.Background()

	if err := irsController.Connect(ctx); err != nil {
		logger.Warn("Failed to connect IRS controller", zap.Error(err))
	}

	if err := usrpReceiver.Connect(ctx); err != nil {
		logger.Warn("Failed to connect USRP receiver", zap.Error(err))
	}

	if err := sensorCollector.Connect(ctx); err != nil {
		logger.Warn("Failed to connect sensor collector", zap.Error(err))
	}

	var channelDataRepo *influxdb.ChannelDataRepository
	var sensorDataRepo *influxdb.SensorDataRepository
	var experimentRepo *mysql.ExperimentRepository

	if influxClient != nil {
		channelDataRepo = influxdb.NewChannelDataRepository(influxClient)
		sensorDataRepo = influxdb.NewSensorDataRepository(influxClient)
	}

	if db != nil {
		experimentRepo = mysql.NewExperimentRepository(db)
	}

	irsSvc := service.NewIRSService(irsController)
	channelSvc := service.NewChannelService(usrpReceiver, channelDataRepo)
	algorithmSvc := service.NewAlgorithmService(experimentRepo)
	sensorSvc := service.NewSensorService(sensorCollector, sensorDataRepo)

	beamformingOptimizer := beamforming.NewOptimizer(
		cfg.Algorithm.Beamforming.MaxIterations,
		cfg.Algorithm.Beamforming.MaxIterations,
		cfg.Algorithm.Beamforming.ConvergenceThreshold,
	)
	_ = beamformingOptimizer

	doaEstimator := doa.NewEstimator(
		64,
		cfg.Algorithm.DOA.NumSources,
		cfg.Algorithm.DOA.SnapshotLength,
		cfg.Algorithm.DOA.Method,
	)
	_ = doaEstimator

	irsHandler := handler.NewIRSHandler(irsSvc)
	channelHandler := handler.NewChannelHandler(channelSvc)
	algorithmHandler := handler.NewAlgorithmHandler(algorithmSvc)
	sensorHandler := handler.NewSensorHandler(sensorSvc)
	systemHandler := handler.NewSystemHandler()

	engine := router.Setup(irsHandler, channelHandler, algorithmHandler, sensorHandler, systemHandler)

	engine.Use(func(c *gin.Context) {
		if len(c.Request.URL.Path) >= 7 && c.Request.URL.Path[:7] == "/debug/" {
			c.Next()
			return
		}
		middleware.RateLimit(100, time.Minute)(c)
	})

	middleware.RegisterMetrics(engine)

	workerPool := pool.NewWorkerPool(10, 100)
	workerPool.Start()
	defer workerPool.Stop()

	taskQueue := queue.NewTaskQueue(5, 100)
	taskQueue.Start()
	defer taskQueue.Stop()

	logger.Info("Worker pool and task queue started")

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Server starting",
			zap.Int("port", cfg.Server.Port),
			zap.String("mode", cfg.Server.Mode),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	irsController.Disconnect()
	usrpReceiver.Disconnect()
	sensorCollector.Disconnect()

	logger.Info("Server exited properly")
}
