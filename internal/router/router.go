package router

import (
	"isac-cran-system/internal/handler"
	"isac-cran-system/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(
	irsHandler *handler.IRSHandler,
	channelHandler *handler.ChannelHandler,
	algorithmHandler *handler.AlgorithmHandler,
	sensorHandler *handler.SensorHandler,
	systemHandler *handler.SystemHandler,
) *gin.Engine {
	router := gin.New()

	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	api := router.Group("/api/v1")
	{
		api.GET("/health", systemHandler.Health)
		api.GET("/info", systemHandler.Info)

		irs := api.Group("/irs")
		{
			irs.POST("/config", irsHandler.Configure)
			irs.GET("/status", irsHandler.GetStatus)
			irs.GET("/config", irsHandler.GetCurrentConfig)
			irs.POST("/optimal", irsHandler.ApplyOptimal)
		}

		channel := api.Group("/channel")
		{
			channel.POST("/collect", channelHandler.Collect)
			channel.GET("/data", channelHandler.Query)
			channel.GET("/realtime", channelHandler.GetRealtime)
		}

		algorithm := api.Group("/algorithm")
		{
			algorithm.POST("/beamforming", algorithmHandler.RunBeamforming)
			algorithm.POST("/doa", algorithmHandler.RunDOA)
			algorithm.GET("/result/:id", algorithmHandler.GetResult)
			algorithm.GET("/results", algorithmHandler.ListResults)
		}

		sensor := api.Group("/sensor")
		{
			sensor.GET("/list", sensorHandler.List)
			sensor.GET("/data", sensorHandler.GetData)
			sensor.GET("/read/:id", sensorHandler.ReadSensor)
			sensor.POST("/start", sensorHandler.StartCollection)
			sensor.POST("/stop", sensorHandler.StopCollection)
		}
	}

	return router
}
