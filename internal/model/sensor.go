package model

import (
	"time"
)

type SensorData struct {
	SensorID   string    `json:"sensor_id"`
	SensorType string    `json:"sensor_type"`
	Location   string    `json:"location"`
	Value      float64   `json:"value"`
	Unit       string    `json:"unit"`
	Quality    float64   `json:"quality"`
	Timestamp  time.Time `json:"timestamp"`
}

func (SensorData) MeasurementName() string {
	return "sensor_data"
}

type SensorType string

const (
	SensorTypeTemperature SensorType = "temperature"
	SensorTypeHumidity    SensorType = "humidity"
	SensorTypePressure    SensorType = "pressure"
	SensorTypeVoltage     SensorType = "voltage"
	SensorTypeCurrent     SensorType = "current"
	SensorTypePower       SensorType = "power"
)

type SensorInfo struct {
	SensorID   string     `json:"sensor_id" gorm:"primaryKey"`
	SensorType SensorType `json:"sensor_type" gorm:"type:varchar(50)"`
	Location   string     `json:"location" gorm:"type:varchar(100)"`
	Unit       string     `json:"unit" gorm:"type:varchar(20)"`
	MinValue   float64    `json:"min_value"`
	MaxValue   float64    `json:"max_value"`
	Status     int        `json:"status" gorm:"type:tinyint;default:1"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (SensorInfo) TableName() string {
	return "sensor_info"
}

type SensorDataQuery struct {
	SensorID   string    `form:"sensor_id"`
	SensorType string    `form:"sensor_type"`
	Location   string    `form:"location"`
	StartTime  time.Time `form:"start_time" time_format:"2006-01-02T15:04:05"`
	EndTime    time.Time `form:"end_time" time_format:"2006-01-02T15:04:05"`
	Page       int       `form:"page" binding:"min=1"`
	PageSize   int       `form:"page_size" binding:"min=1,max=100"`
}

type SensorSubscribeRequest struct {
	SensorIDs []string `json:"sensor_ids" binding:"required"`
	Topic     string   `json:"topic" binding:"required"`
}

type SensorAggregatedData struct {
	SensorID   string  `json:"sensor_id"`
	SensorType string  `json:"sensor_type"`
	AvgValue   float64 `json:"avg_value"`
	MinValue   float64 `json:"min_value"`
	MaxValue   float64 `json:"max_value"`
	Count      int     `json:"count"`
}
