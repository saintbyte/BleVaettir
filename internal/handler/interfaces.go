package handler

type HandlerConfig struct {
	DB       *DBHandlerConfig
	HTTP     *HTTPHandlerConfig
	Narodmon *NarodmonHandlerConfig
	Log      *LogHandlerConfig
	DataCake *DataCakeHandlerConfig
}

type DBHandlerConfig struct {
	Enabled bool
}

type HTTPHandlerConfig struct {
	Enabled  bool
	Endpoint string
	APIKey   string
}

type NarodmonHandlerConfig struct {
	Enabled  bool
	Endpoint string
	Owner    string
	Lat      string
	Lon      string
	Alt      string
}

type DataCakeHandlerConfig struct {
	Enabled  bool
	Endpoint string
}

type LogHandlerConfig struct {
	Enabled bool
}

type Handler interface {
	Handle(reading *Reading, cfg *HandlerConfig) error
	Name() string
}
