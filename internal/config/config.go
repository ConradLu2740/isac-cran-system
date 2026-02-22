package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	MySQL     MySQLConfig     `mapstructure:"mysql"`
	InfluxDB  InfluxDBConfig  `mapstructure:"influxdb"`
	Redis     RedisConfig     `mapstructure:"redis"`
	MQTT      MQTTConfig      `mapstructure:"mqtt"`
	Log       LogConfig       `mapstructure:"log"`
	Device    DeviceConfig    `mapstructure:"device"`
	Algorithm AlgorithmConfig `mapstructure:"algorithm"`
	MATLAB    MATLABConfig    `mapstructure:"matlab"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type MySQLConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Database        string `mapstructure:"database"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Charset         string `mapstructure:"charset"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

func (c *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.Charset)
}

type InfluxDBConfig struct {
	URL    string `mapstructure:"url"`
	Token  string `mapstructure:"token"`
	Org    string `mapstructure:"org"`
	Bucket string `mapstructure:"bucket"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type MQTTConfig struct {
	Broker   string `mapstructure:"broker"`
	ClientID string `mapstructure:"client_id"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	QoS      byte   `mapstructure:"qos"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

type DeviceConfig struct {
	IRS    IRSDeviceConfig    `mapstructure:"irs"`
	USRP   USRPDeviceConfig   `mapstructure:"usrp"`
	Sensor SensorDeviceConfig `mapstructure:"sensor"`
}

type IRSDeviceConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	Simulator     bool   `mapstructure:"simulator"`
	ElementCount  int    `mapstructure:"element_count"`
	FrequencyBand string `mapstructure:"frequency_band"`
}

type USRPDeviceConfig struct {
	Enabled    bool    `mapstructure:"enabled"`
	Simulator  bool    `mapstructure:"simulator"`
	SampleRate float64 `mapstructure:"sample_rate"`
	CenterFreq float64 `mapstructure:"center_freq"`
}

type SensorDeviceConfig struct {
	Enabled            bool          `mapstructure:"enabled"`
	Simulator          bool          `mapstructure:"simulator"`
	CollectionInterval time.Duration `mapstructure:"collection_interval"`
}

type AlgorithmConfig struct {
	Beamforming BeamformingConfig `mapstructure:"beamforming"`
	DOA         DOAConfig         `mapstructure:"doa"`
}

type BeamformingConfig struct {
	MaxIterations        int     `mapstructure:"max_iterations"`
	ConvergenceThreshold float64 `mapstructure:"convergence_threshold"`
}

type DOAConfig struct {
	Method         string `mapstructure:"method"`
	NumSources     int    `mapstructure:"num_sources"`
	SnapshotLength int    `mapstructure:"snapshot_length"`
}

type MATLABConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	DataDir      string `mapstructure:"data_dir"`
	ExportFormat string `mapstructure:"export_format"`
}

var globalConfig *Config

func Init(configPath string) error {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	globalConfig = &cfg
	return nil
}

func Get() *Config {
	if globalConfig == nil {
		panic("config not initialized, please call Init() first")
	}
	return globalConfig
}

func GetServer() *ServerConfig {
	return &Get().Server
}

func GetMySQL() *MySQLConfig {
	return &Get().MySQL
}

func GetInfluxDB() *InfluxDBConfig {
	return &Get().InfluxDB
}

func GetRedis() *RedisConfig {
	return &Get().Redis
}

func GetMQTT() *MQTTConfig {
	return &Get().MQTT
}

func GetLog() *LogConfig {
	return &Get().Log
}

func GetDevice() *DeviceConfig {
	return &Get().Device
}

func GetAlgorithm() *AlgorithmConfig {
	return &Get().Algorithm
}

func GetMATLAB() *MATLABConfig {
	return &Get().MATLAB
}
