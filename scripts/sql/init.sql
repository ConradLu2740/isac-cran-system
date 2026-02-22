CREATE DATABASE IF NOT EXISTS isac_cran DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE isac_cran;

CREATE TABLE IF NOT EXISTS irs_config (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT 'Configuration name',
    element_count INT NOT NULL COMMENT 'Number of IRS elements',
    phase_shifts JSON COMMENT 'Phase shift values array',
    frequency_band VARCHAR(50) COMMENT 'Frequency band',
    status TINYINT DEFAULT 1 COMMENT 'Status: 0=inactive, 1=active, 2=applied',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='IRS configuration table';

CREATE TABLE IF NOT EXISTS experiment_result (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    experiment_id VARCHAR(50) NOT NULL UNIQUE COMMENT 'Experiment ID',
    algorithm_type VARCHAR(50) NOT NULL COMMENT 'Algorithm type: beamforming, doa, scheduling, rateless',
    parameters JSON COMMENT 'Experiment parameters',
    result_data JSON COMMENT 'Result data',
    matlab_file_path VARCHAR(255) COMMENT 'MATLAB file path for data exchange',
    status TINYINT DEFAULT 0 COMMENT 'Status: 0=pending, 1=running, 2=completed, 3=failed',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL COMMENT 'Completion time',
    INDEX idx_experiment_id (experiment_id),
    INDEX idx_algorithm_type (algorithm_type),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Experiment result table';

CREATE TABLE IF NOT EXISTS sensor_info (
    sensor_id VARCHAR(50) PRIMARY KEY COMMENT 'Sensor ID',
    sensor_type VARCHAR(50) NOT NULL COMMENT 'Sensor type: temperature, humidity, pressure, etc.',
    location VARCHAR(100) COMMENT 'Sensor location',
    unit VARCHAR(20) COMMENT 'Measurement unit',
    min_value DOUBLE COMMENT 'Minimum value range',
    max_value DOUBLE COMMENT 'Maximum value range',
    status TINYINT DEFAULT 1 COMMENT 'Status: 0=offline, 1=online',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_sensor_type (sensor_type),
    INDEX idx_location (location),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Sensor information table';

INSERT INTO sensor_info (sensor_id, sensor_type, location, unit, min_value, max_value, status) VALUES
('temp-001', 'temperature', 'Room-A', '°C', 15, 35, 1),
('temp-002', 'temperature', 'Room-B', '°C', 15, 35, 1),
('hum-001', 'humidity', 'Room-A', '%', 30, 90, 1),
('hum-002', 'humidity', 'Room-B', '%', 30, 90, 1),
('press-001', 'pressure', 'Room-A', 'kPa', 95, 105, 1),
('volt-001', 'voltage', 'Power-Unit-1', 'V', 200, 240, 1),
('curr-001', 'current', 'Power-Unit-1', 'A', 0, 50, 1),
('power-001', 'power', 'Power-Unit-1', 'kW', 0, 10, 1);
