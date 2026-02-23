package model

import (
	"time"
)

type IRSConfig struct {
	ID            int64        `json:"id" gorm:"primaryKey;autoIncrement"`
	Name          string       `json:"name" gorm:"type:varchar(100);not null"`
	ElementCount  int          `json:"element_count" gorm:"not null"`
	PhaseShifts   []float64    `json:"phase_shifts" gorm:"type:json"`
	FrequencyBand string       `json:"frequency_band" gorm:"type:varchar(50)"`
	Status        ConfigStatus `json:"status" gorm:"type:tinyint;default:1"`
	CreatedAt     time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}

type ConfigStatus int

const (
	ConfigStatusInactive ConfigStatus = 0
	ConfigStatusActive   ConfigStatus = 1
	ConfigStatusApplied  ConfigStatus = 2
)

func (IRSConfig) TableName() string {
	return "irs_config"
}

type IRSStatus struct {
	ElementCount  int       `json:"element_count"`
	PhaseShifts   []float64 `json:"phase_shifts"`
	FrequencyBand string    `json:"frequency_band"`
	Temperature   float64   `json:"temperature"`
	PowerStatus   bool      `json:"power_status"`
	LastUpdate    time.Time `json:"last_update"`
}

type IRSConfigRequest struct {
	Name          string    `json:"name" binding:"required"`
	ElementCount  int       `json:"element_count" binding:"required,min=1,max=256"`
	PhaseShifts   []float64 `json:"phase_shifts" binding:"required"`
	FrequencyBand string    `json:"frequency_band" binding:"required"`
}

func (r *IRSConfigRequest) Validate() error {
	if len(r.PhaseShifts) != r.ElementCount {
		return NewValidationError("phase_shifts length must equal element_count")
	}
	for i, ps := range r.PhaseShifts {
		if ps < 0 || ps > 2*3.14159265359 {
			return NewValidationErrorf("phase_shift[%d] must be in range [0, 2π]", i)
		}
	}
	return nil
}
