package usrp

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
		SampleRate: 10e6,
		CenterFreq: 2.4e9,
		Gain:       30.0,
	}

	for _, opt := range opts {
		opt(config)
	}

	switch driverType {
	case DriverTypeSimulator:
		return NewSimulator(config.SampleRate, config.CenterFreq), nil
	case DriverTypeHardware:
		return nil, ErrHardwareDriverNotImplemented
	default:
		return nil, ErrUnknownDriverType
	}
}

type DriverConfig struct {
	SampleRate float64
	CenterFreq float64
	Gain       float64
	IPAddress  string
	Port       int
}

type DriverOption func(*DriverConfig)

func WithSampleRate(rate float64) DriverOption {
	return func(c *DriverConfig) {
		c.SampleRate = rate
	}
}

func WithCenterFreq(freq float64) DriverOption {
	return func(c *DriverConfig) {
		c.CenterFreq = freq
	}
}

func WithGain(gain float64) DriverOption {
	return func(c *DriverConfig) {
		c.Gain = gain
	}
}

func WithIPAddress(ip string) DriverOption {
	return func(c *DriverConfig) {
		c.IPAddress = ip
	}
}

func WithPort(port int) DriverOption {
	return func(c *DriverConfig) {
		c.Port = port
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
