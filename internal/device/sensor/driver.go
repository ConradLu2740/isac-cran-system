package sensor

type DriverType string

const (
	DriverTypeSimulator DriverType = "simulator"
	DriverTypeMQTT      DriverType = "mqtt"
	DriverTypeSerial    DriverType = "serial"
)

type DriverFactory struct{}

func NewDriverFactory() *DriverFactory {
	return &DriverFactory{}
}

func (f *DriverFactory) Create(driverType DriverType, opts ...DriverOption) (Driver, error) {
	config := &DriverConfig{}

	for _, opt := range opts {
		opt(config)
	}

	switch driverType {
	case DriverTypeSimulator:
		return NewSimulator(), nil
	case DriverTypeMQTT:
		return nil, ErrMQTTDriverNotImplemented
	case DriverTypeSerial:
		return nil, ErrSerialDriverNotImplemented
	default:
		return nil, ErrUnknownDriverType
	}
}

type DriverConfig struct {
	BrokerURL  string
	ClientID   string
	SerialPort string
	BaudRate   int
}

type DriverOption func(*DriverConfig)

func WithBrokerURL(url string) DriverOption {
	return func(c *DriverConfig) {
		c.BrokerURL = url
	}
}

func WithClientID(id string) DriverOption {
	return func(c *DriverConfig) {
		c.ClientID = id
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
	ErrMQTTDriverNotImplemented   = &FactoryError{Message: "mqtt driver not implemented"}
	ErrSerialDriverNotImplemented = &FactoryError{Message: "serial driver not implemented"}
	ErrUnknownDriverType          = &FactoryError{Message: "unknown driver type"}
)

type FactoryError struct {
	Message string
}

func (e *FactoryError) Error() string {
	return e.Message
}
