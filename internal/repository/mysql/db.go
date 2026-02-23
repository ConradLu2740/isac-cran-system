package mysql

import (
	"context"
	"time"

	"isac-cran-system/internal/config"
	"isac-cran-system/internal/model"
	"isac-cran-system/pkg/errors"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	*gorm.DB
}

func NewDB(cfg *config.MySQLConfig) (*DB, error) {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN()), gormConfig)
	if err != nil {
		return nil, errors.Wrap(errors.CodeDBConnectError, "failed to connect mysql", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(errors.CodeDBConnectError, "failed to get sql db", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, errors.Wrap(errors.CodeDBConnectError, "failed to ping mysql", err)
	}

	return &DB{DB: db}, nil
}

func (db *DB) AutoMigrate() error {
	return db.DB.AutoMigrate(
		&model.IRSConfig{},
		&model.ExperimentResult{},
		&model.SensorInfo{},
	)
}

func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

type IRSConfigRepository struct {
	db *DB
}

func NewIRSConfigRepository(db *DB) *IRSConfigRepository {
	return &IRSConfigRepository{db: db}
}

func (r *IRSConfigRepository) Create(ctx context.Context, config *model.IRSConfig) error {
	if err := r.db.WithContext(ctx).Create(config).Error; err != nil {
		return errors.Wrap(errors.CodeDBInsertError, "failed to create irs config", err)
	}
	return nil
}

func (r *IRSConfigRepository) GetByID(ctx context.Context, id int64) (*model.IRSConfig, error) {
	var config model.IRSConfig
	if err := r.db.WithContext(ctx).First(&config, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeNotFound, "irs config not found")
		}
		return nil, errors.Wrap(errors.CodeDBQueryError, "failed to get irs config", err)
	}
	return &config, nil
}

func (r *IRSConfigRepository) List(ctx context.Context, page, pageSize int) ([]model.IRSConfig, int64, error) {
	var configs []model.IRSConfig
	var total int64

	if err := r.db.WithContext(ctx).Model(&model.IRSConfig{}).Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDBQueryError, "failed to count irs configs", err)
	}

	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Find(&configs).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDBQueryError, "failed to list irs configs", err)
	}

	return configs, total, nil
}

func (r *IRSConfigRepository) Update(ctx context.Context, config *model.IRSConfig) error {
	result := r.db.WithContext(ctx).Save(config)
	if result.Error != nil {
		return errors.Wrap(errors.CodeDBUpdateError, "failed to update irs config", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New(errors.CodeNotFound, "irs config not found")
	}
	return nil
}

func (r *IRSConfigRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&model.IRSConfig{}, id)
	if result.Error != nil {
		return errors.Wrap(errors.CodeDBUpdateError, "failed to delete irs config", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New(errors.CodeNotFound, "irs config not found")
	}
	return nil
}

type ExperimentRepository struct {
	db *DB
}

func NewExperimentRepository(db *DB) *ExperimentRepository {
	return &ExperimentRepository{db: db}
}

func (r *ExperimentRepository) Create(ctx context.Context, result *model.ExperimentResult) error {
	if err := r.db.WithContext(ctx).Create(result).Error; err != nil {
		return errors.Wrap(errors.CodeDBInsertError, "failed to create experiment result", err)
	}
	return nil
}

func (r *ExperimentRepository) GetByID(ctx context.Context, id int64) (*model.ExperimentResult, error) {
	var result model.ExperimentResult
	if err := r.db.WithContext(ctx).First(&result, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeNotFound, "experiment result not found")
		}
		return nil, errors.Wrap(errors.CodeDBQueryError, "failed to get experiment result", err)
	}
	return &result, nil
}

func (r *ExperimentRepository) GetByExperimentID(ctx context.Context, experimentID string) (*model.ExperimentResult, error) {
	var result model.ExperimentResult
	if err := r.db.WithContext(ctx).Where("experiment_id = ?", experimentID).First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeNotFound, "experiment result not found")
		}
		return nil, errors.Wrap(errors.CodeDBQueryError, "failed to get experiment result", err)
	}
	return &result, nil
}

func (r *ExperimentRepository) List(ctx context.Context, algorithmType model.AlgorithmType, page, pageSize int) ([]model.ExperimentResult, int64, error) {
	var results []model.ExperimentResult
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ExperimentResult{})
	if algorithmType != "" {
		query = query.Where("algorithm_type = ?", algorithmType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDBQueryError, "failed to count experiment results", err)
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&results).Error; err != nil {
		return nil, 0, errors.Wrap(errors.CodeDBQueryError, "failed to list experiment results", err)
	}

	return results, total, nil
}

func (r *ExperimentRepository) UpdateStatus(ctx context.Context, id int64, status model.ExperimentStatus, resultData string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if resultData != "" {
		updates["result_data"] = resultData
	}
	if status == model.ExperimentStatusCompleted || status == model.ExperimentStatusFailed {
		now := time.Now()
		updates["completed_at"] = &now
	}

	result := r.db.WithContext(ctx).Model(&model.ExperimentResult{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return errors.Wrap(errors.CodeDBUpdateError, "failed to update experiment status", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New(errors.CodeNotFound, "experiment result not found")
	}
	return nil
}

func (r *ExperimentRepository) UpdateMATLABPath(ctx context.Context, id int64, path string) error {
	result := r.db.WithContext(ctx).Model(&model.ExperimentResult{}).Where("id = ?", id).Update("matlab_file_path", path)
	if result.Error != nil {
		return errors.Wrap(errors.CodeDBUpdateError, "failed to update matlab path", result.Error)
	}
	return nil
}

type SensorInfoRepository struct {
	db *DB
}

func NewSensorInfoRepository(db *DB) *SensorInfoRepository {
	return &SensorInfoRepository{db: db}
}

func (r *SensorInfoRepository) Create(ctx context.Context, info *model.SensorInfo) error {
	if err := r.db.WithContext(ctx).Create(info).Error; err != nil {
		return errors.Wrap(errors.CodeDBInsertError, "failed to create sensor info", err)
	}
	return nil
}

func (r *SensorInfoRepository) GetByID(ctx context.Context, sensorID string) (*model.SensorInfo, error) {
	var info model.SensorInfo
	if err := r.db.WithContext(ctx).Where("sensor_id = ?", sensorID).First(&info).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.CodeNotFound, "sensor info not found")
		}
		return nil, errors.Wrap(errors.CodeDBQueryError, "failed to get sensor info", err)
	}
	return &info, nil
}

func (r *SensorInfoRepository) List(ctx context.Context, sensorType model.SensorType) ([]model.SensorInfo, error) {
	var sensors []model.SensorInfo
	query := r.db.WithContext(ctx)
	if sensorType != "" {
		query = query.Where("sensor_type = ?", sensorType)
	}
	if err := query.Find(&sensors).Error; err != nil {
		return nil, errors.Wrap(errors.CodeDBQueryError, "failed to list sensor info", err)
	}
	return sensors, nil
}

func (r *SensorInfoRepository) Update(ctx context.Context, info *model.SensorInfo) error {
	result := r.db.WithContext(ctx).Save(info)
	if result.Error != nil {
		return errors.Wrap(errors.CodeDBUpdateError, "failed to update sensor info", result.Error)
	}
	return nil
}
