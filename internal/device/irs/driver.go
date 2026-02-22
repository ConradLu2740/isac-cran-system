package irs

type DriverType string

const (
	DriverTypeSimulator DriverType = "simulator"
	DriverTypeHardware  DriverType = "hardware"
)

type DriverFactory struct{}

func NewDriverFactory() *DriverFactory {
	return &DriverFactory{}
}

func (f *DriverFactory) Create(driverType DriverType, opts ...DriverOption) (Driver, error) {
	config := &DriverConfig{
		ElementCount:  64,
		FrequencyBand: "2.4GHz",
	}

	for _, opt := range opts {
		opt(config)
	}

	switch driverType {
	case DriverTypeSimulator:
		return NewSimulator(config.ElementCount, config.FrequencyBand), nil
	case DriverTypeHardware:
		return nil, ErrHardwareDriverNotImplemented
	default:
		return nil, ErrUnknownDriverType
	}
}

type DriverConfig struct {
	ElementCount  int
	FrequencyBand string
	SerialPort    string
	BaudRate      int
}

type DriverOption func(*DriverConfig)

func WithElementCount(count int) DriverOption {
	return func(c *DriverConfig) {
		c.ElementCount = count
	}
}

func WithFrequencyBand(band string) DriverOption {
	return func(c *DriverConfig) {
		c.FrequencyBand = band
	}
}

func WithSerialPort(port string) DriverOption {
	return func(c *DriverConfig) {
		c.SerialPort = port
	}
}

func WithBaudRate(rate int) DriverOption {
	return func(c *DriverConfig) {
		c.BaudRate = rate
	}
}

var (
	ErrHardwareDriverNotImplemented = &FactoryError{Message: "hardware driver not implemented"}
	ErrUnknownDriverType            = &FactoryError{Message: "unknown driver type"}
)

type FactoryError struct {
	Message string
}

func (e *FactoryError) Error() string {
	return e.Message
}
