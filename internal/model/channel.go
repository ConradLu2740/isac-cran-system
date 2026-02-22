package model

import (
	"time"
)

type ChannelMeasurement struct {
	MeasurementID string    `json:"measurement_id"`
	ExperimentID  string    `json:"experiment_id"`
	UserID        int       `json:"user_id"`
	FrequencyBand string    `json:"frequency_band"`
	Amplitude     []float64 `json:"amplitude"`
	Phase         []float64 `json:"phase"`
	SNR           float64   `json:"snr"`
	BER           float64   `json:"ber"`
	Timestamp     time.Time `json:"timestamp"`
}

func (ChannelMeasurement) MeasurementName() string {
	return "channel_measurement"
}

type ChannelDataQuery struct {
	ExperimentID  string    `form:"experiment_id"`
	UserID        int       `form:"user_id"`
	FrequencyBand string    `form:"frequency_band"`
	StartTime     time.Time `form:"start_time" time_format:"2006-01-02T15:04:05"`
	EndTime       time.Time `form:"end_time" time_format:"2006-01-02T15:04:05"`
	Page          int       `form:"page" binding:"min=1"`
	PageSize      int       `form:"page_size" binding:"min=1,max=100"`
}

type ChannelCollectRequest struct {
	ExperimentID  string  `json:"experiment_id" binding:"required"`
	UserID        int     `json:"user_id" binding:"required"`
	FrequencyBand string  `json:"frequency_band" binding:"required"`
	Duration      float64 `json:"duration" binding:"required,min=0.1,max=60"`
	SampleRate    float64 `json:"sample_rate" binding:"required,min=1000,max=100000000"`
}

type ChannelDataPoint struct {
	Index     int     `json:"index"`
	Amplitude float64 `json:"amplitude"`
	Phase     float64 `json:"phase"`
	I         float64 `json:"i"`
	Q         float64 `json:"q"`
}
