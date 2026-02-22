package influxdb

import (
	"context"
	"time"

	"isac-cran-system/internal/config"
	"isac-cran-system/internal/model"
	"isac-cran-system/pkg/errors"
	"isac-cran-system/pkg/logger"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"go.uber.org/zap"
)

type Client struct {
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
	queryAPI api.QueryAPI
	org      string
	bucket   string
}

func NewClient(cfg *config.InfluxDBConfig) (*Client, error) {
	client := influxdb2.NewClient(cfg.URL, cfg.Token)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health, err := client.Health(ctx)
	if err != nil {
		return nil, errors.Wrap(errors.CodeInfluxConnectError, "failed to connect influxdb", err)
	}

	if health.Status != "pass" {
		return nil, errors.NewWithDetail(errors.CodeInfluxConnectError, "influxdb health check failed", string(health.Status))
	}

	writeAPI := client.WriteAPIBlocking(cfg.Org, cfg.Bucket)
	queryAPI := client.QueryAPI(cfg.Org)

	logger.Info("InfluxDB connected",
		zap.String("url", cfg.URL),
		zap.String("org", cfg.Org),
		zap.String("bucket", cfg.Bucket),
	)

	return &Client{
		client:   client,
		writeAPI: writeAPI,
		queryAPI: queryAPI,
		org:      cfg.Org,
		bucket:   cfg.Bucket,
	}, nil
}

func (c *Client) Close() {
	c.client.Close()
}

type ChannelDataRepository struct {
	client *Client
}

func NewChannelDataRepository(client *Client) *ChannelDataRepository {
	return &ChannelDataRepository{client: client}
}

func (r *ChannelDataRepository) Write(ctx context.Context, data *model.ChannelMeasurement) error {
	p := influxdb2.NewPoint(
		data.MeasurementName(),
		map[string]string{
			"experiment_id":  data.ExperimentID,
			"user_id":        string(rune(data.UserID)),
			"frequency_band": data.FrequencyBand,
		},
		map[string]interface{}{
			"amplitude": data.Amplitude,
			"phase":     data.Phase,
			"snr":       data.SNR,
			"ber":       data.BER,
		},
		data.Timestamp,
	)

	if err := r.client.writeAPI.WritePoint(ctx, p); err != nil {
		return errors.Wrap(errors.CodeInfluxWriteError, "failed to write channel data", err)
	}

	return nil
}

func (r *ChannelDataRepository) WriteBatch(ctx context.Context, dataPoints []*model.ChannelMeasurement) error {
	for _, data := range dataPoints {
		if err := r.Write(ctx, data); err != nil {
			return err
		}
	}
	return nil
}

func (r *ChannelDataRepository) Query(ctx context.Context, q *model.ChannelDataQuery) ([]*model.ChannelMeasurement, error) {
	return []*model.ChannelMeasurement{}, nil
}

type SensorDataRepository struct {
	client *Client
}

func NewSensorDataRepository(client *Client) *SensorDataRepository {
	return &SensorDataRepository{client: client}
}

func (r *SensorDataRepository) Write(ctx context.Context, data *model.SensorData) error {
	p := influxdb2.NewPoint(
		data.MeasurementName(),
		map[string]string{
			"sensor_id":   data.SensorID,
			"sensor_type": data.SensorType,
			"location":    data.Location,
		},
		map[string]interface{}{
			"value":   data.Value,
			"quality": data.Quality,
		},
		data.Timestamp,
	)

	if err := r.client.writeAPI.WritePoint(ctx, p); err != nil {
		return errors.Wrap(errors.CodeInfluxWriteError, "failed to write sensor data", err)
	}

	return nil
}

func (r *SensorDataRepository) WriteBatch(ctx context.Context, dataPoints []*model.SensorData) error {
	for _, data := range dataPoints {
		if err := r.Write(ctx, data); err != nil {
			return err
		}
	}
	return nil
}

func (r *SensorDataRepository) Query(ctx context.Context, q *model.SensorDataQuery) ([]*model.SensorData, error) {
	return []*model.SensorData{}, nil
}
