package model

import (
	"time"
)

type ExperimentResult struct {
	ID             int64            `json:"id" gorm:"primaryKey;autoIncrement"`
	ExperimentID   string           `json:"experiment_id" gorm:"type:varchar(50);uniqueIndex;not null"`
	AlgorithmType  AlgorithmType    `json:"algorithm_type" gorm:"type:varchar(50);not null"`
	Parameters     string           `json:"parameters" gorm:"type:json"`
	ResultData     *string          `json:"result_data" gorm:"type:json"`
	MATLABFilePath string           `json:"matlab_file_path" gorm:"type:varchar(255)"`
	Status         ExperimentStatus `json:"status" gorm:"type:tinyint;default:1"`
	CreatedAt      time.Time        `json:"created_at" gorm:"autoCreateTime"`
	CompletedAt    *time.Time       `json:"completed_at"`
}

type AlgorithmType string

const (
	AlgorithmTypeBeamforming AlgorithmType = "beamforming"
	AlgorithmTypeDOA         AlgorithmType = "doa"
	AlgorithmTypeScheduling  AlgorithmType = "scheduling"
	AlgorithmTypeRateless    AlgorithmType = "rateless"
)

type ExperimentStatus int

const (
	ExperimentStatusPending   ExperimentStatus = 0
	ExperimentStatusRunning   ExperimentStatus = 1
	ExperimentStatusCompleted ExperimentStatus = 2
	ExperimentStatusFailed    ExperimentStatus = 3
)

func (ExperimentResult) TableName() string {
	return "experiment_result"
}

type ExperimentCreateRequest struct {
	ExperimentID  string        `json:"experiment_id" binding:"required"`
	AlgorithmType AlgorithmType `json:"algorithm_type" binding:"required"`
	Parameters    interface{}   `json:"parameters" binding:"required"`
}

type BeamformingParams struct {
	ElementCount       int       `json:"element_count"`
	TargetDirection    float64   `json:"target_direction"`
	InterferenceAngles []float64 `json:"interference_angles"`
	SNRThreshold       float64   `json:"snr_threshold"`
	MaxIterations      int       `json:"max_iterations"`
}

type DOAParams struct {
	ElementCount   int     `json:"element_count"`
	NumSources     int     `json:"num_sources"`
	SnapshotLength int     `json:"snapshot_length"`
	Method         string  `json:"method"`
	SearchRangeMin float64 `json:"search_range_min"`
	SearchRangeMax float64 `json:"search_range_max"`
	SearchStep     float64 `json:"search_step"`
}

type BeamformingResult struct {
	Weights           [][]float64 `json:"weights"`
	BeamPattern       []float64   `json:"beam_pattern"`
	MainLobeDirection float64     `json:"main_lobe_direction"`
	MainLobeWidth     float64     `json:"main_lobe_width"`
	SLL               float64     `json:"side_lobe_level"`
	Iterations        int         `json:"iterations"`
	Converged         bool        `json:"converged"`
}

type DOAResult struct {
	EstimatedAngles []float64 `json:"estimated_angles"`
	Spectrum        []float64 `json:"spectrum"`
	TrueAngles      []float64 `json:"true_angles,omitempty"`
	RMSE            float64   `json:"rmse,omitempty"`
}
