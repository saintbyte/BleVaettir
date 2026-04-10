package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/saintbyte/BleVaettir/internal/config"
	"github.com/saintbyte/BleVaettir/internal/handler"
	"github.com/saintbyte/BleVaettir/internal/scanner"
	"github.com/saintbyte/BleVaettir/internal/storage"
)

var (
	configPath = flag.String("config", "blevaettir.yaml", "Path to config file")
	daemonMode = flag.Bool("d", false, "Run as daemon")
	logFile    = flag.String("log-file", "", "Log file path (for daemon mode)")
)

func main() {
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if *daemonMode {
		daemonize()
	}

	initLogger(cfg.Log.Level)

	store, err := storage.New(cfg.Storage.Path)
	if err != nil {
		slog.Error("failed to open storage", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	objectHandlers, unknownHandlers := buildObjectHandlers(cfg, store)

	sc, err := scanner.New(cfg, objectHandlers, unknownHandlers)
	if err != nil {
		slog.Error("failed to init scanner", "error", err)
		os.Exit(1)
	}
	defer sc.Close()

	stop := make(chan struct{})
	go sc.Run(stop)

	slog.Info("BleVaettir daemon started",
		"hci", cfg.BLE.HCI,
		"objects", len(cfg.BLEObjects),
		"handlers", countHandlers(objectHandlers),
		"unknown_objects", cfg.UnknownObjects.Enabled,
	)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	slog.Info("shutting down")
	close(stop)
}

func daemonize() {
	if os.Getenv("BLEVAETTIR_DAEMON") == "1" {
		return
	}

	exe, err := os.Executable()
	if err != nil {
		slog.Error("failed to get executable path", "error", err)
		os.Exit(1)
	}

	args := os.Args[1:]
	for i, arg := range args {
		if arg == "-d" {
			args = append(args[:i], args[i+1:]...)
			break
		}
	}

	env := append(os.Environ(), "BLEVAETTIR_DAEMON=1")

	procAttr := &os.ProcAttr{
		Dir:   "/",
		Env:   env,
		Files: []*os.File{nil, nil, nil},
	}

	_, err = os.StartProcess(exe, args, procAttr)
	if err != nil {
		slog.Error("failed to start daemon", "error", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func buildObjectHandlers(cfg *config.Config, store *storage.Storage) (map[string][]scanner.HandlerWithConfig, []scanner.HandlerWithConfig) {
	result := make(map[string][]scanner.HandlerWithConfig)
	globalHandlers := buildGlobalHandlers(cfg.Handlers, store)
	handlerFactoryMap := buildHandlerFactoryMap(store)
	unknownHandlers := buildUnknownHandlers(cfg.UnknownObjects, handlerFactoryMap)

	for _, obj := range cfg.BLEObjects {
		if len(obj.Handlers) == 0 {
			result[obj.MAC] = globalHandlers
		} else {
			var objHandlers []scanner.HandlerWithConfig
			for _, h := range obj.Handlers {
				cfg := getHandlerConfig(h, cfg.Handlers)
				if f, ok := handlerFactoryMap[h.Type]; ok {
					objHandlers = append(objHandlers, scanner.HandlerWithConfig{
						Handler: f(),
						Config:  cfg,
					})
				}
			}
			result[obj.MAC] = objHandlers
		}

		if len(result[obj.MAC]) == 0 {
			result[obj.MAC] = []scanner.HandlerWithConfig{
				{Handler: handler.NewDBHandler(store), Config: nil},
			}
		}
	}

	return result, unknownHandlers
}

type handlerFactory func() handler.Handler

func buildGlobalHandlers(handlerCfgs []config.HandlerConfig, store *storage.Storage) []scanner.HandlerWithConfig {
	var handlers []scanner.HandlerWithConfig

	for _, h := range handlerCfgs {
		cfg := getHandlerConfigFromGlobal(h)
		f := getHandlerFactory(h.Type, store)
		if f != nil {
			handlers = append(handlers, scanner.HandlerWithConfig{
				Handler: f(),
				Config:  cfg,
			})
		}
	}

	return handlers
}

func buildHandlerFactoryMap(store *storage.Storage) map[string]handlerFactory {
	return map[string]handlerFactory{
		"db":       func() handler.Handler { return handler.NewDBHandler(store) },
		"http":     func() handler.Handler { return handler.NewHTTPHandler() },
		"narodmon": func() handler.Handler { return handler.NewNarodmonHandler() },
		"log":      func() handler.Handler { return handler.NewLogHandler() },
		"datacake": func() handler.Handler { return handler.NewDataCakeHandler() },
	}
}

func getHandlerFactory(handlerType string, store *storage.Storage) handlerFactory {
	factories := buildHandlerFactoryMap(store)
	return factories[handlerType]
}

func getHandlerConfig(objH config.ObjectHandler, globalCfgs []config.HandlerConfig) *handler.HandlerConfig {
	var cfg handler.HandlerConfig

	switch objH.Type {
	case "db":
		if objH.DB != nil {
			cfg.DB = &handler.DBHandlerConfig{Enabled: objH.DB.Enabled}
		}
	case "http":
		if objH.HTTP != nil {
			cfg.HTTP = &handler.HTTPHandlerConfig{
				Enabled:    objH.HTTP.Enabled,
				Endpoint:   objH.HTTP.Endpoint,
				APIKey:     objH.HTTP.APIKey,
				CACert:     objH.HTTP.CACert,
				ClientCert: objH.HTTP.ClientCert,
				ClientKey:  objH.HTTP.ClientKey,
				SkipVerify: &objH.HTTP.SkipVerify,
			}
		}
	case "narodmon":
		if objH.Narodmon != nil {
			cfg.Narodmon = &handler.NarodmonHandlerConfig{
				Enabled:  objH.Narodmon.Enabled,
				Endpoint: objH.Narodmon.Endpoint,
				Owner:    objH.Narodmon.Owner,
				Lat:      objH.Narodmon.Lat,
				Lon:      objH.Narodmon.Lon,
				Alt:      objH.Narodmon.Alt,
			}
		}
	case "log":
		if objH.Log != nil {
			cfg.Log = &handler.LogHandlerConfig{Enabled: objH.Log.Enabled}
		}
	case "datacake":
		if objH.DataCake != nil {
			cfg.DataCake = &handler.DataCakeHandlerConfig{
				Enabled:    objH.DataCake.Enabled,
				Endpoint:   objH.DataCake.Endpoint,
				SkipVerify: &objH.DataCake.SkipVerify,
			}
		}
	}

	if !hasConfig(&cfg) {
		return nil
	}

	return &cfg
}

func getHandlerConfigFromGlobal(h config.HandlerConfig) *handler.HandlerConfig {
	var cfg handler.HandlerConfig

	switch h.Type {
	case "db":
		if h.DB != nil {
			cfg.DB = &handler.DBHandlerConfig{Enabled: h.DB.Enabled}
		}
	case "http":
		if h.HTTP != nil {
			cfg.HTTP = &handler.HTTPHandlerConfig{
				Enabled:  h.HTTP.Enabled,
				Endpoint: h.HTTP.Endpoint,
				APIKey:   h.HTTP.APIKey,
			}
		}
	case "narodmon":
		if h.Narodmon != nil {
			cfg.Narodmon = &handler.NarodmonHandlerConfig{
				Enabled:  h.Narodmon.Enabled,
				Endpoint: h.Narodmon.Endpoint,
				Owner:    h.Narodmon.Owner,
				Lat:      h.Narodmon.Lat,
				Lon:      h.Narodmon.Lon,
				Alt:      h.Narodmon.Alt,
			}
		}
	case "log":
		if h.Log != nil {
			cfg.Log = &handler.LogHandlerConfig{Enabled: h.Log.Enabled}
		}
	case "datacake":
		if h.DataCake != nil {
			cfg.DataCake = &handler.DataCakeHandlerConfig{
				Enabled:  h.DataCake.Enabled,
				Endpoint: h.DataCake.Endpoint,
			}
		}
	}

	if !hasConfig(&cfg) {
		return nil
	}

	return &cfg
}

func hasConfig(cfg *handler.HandlerConfig) bool {
	return cfg.DB != nil || cfg.HTTP != nil || cfg.Narodmon != nil || cfg.Log != nil || cfg.DataCake != nil
}

func buildUnknownHandlers(cfg config.UnknownObjectsConfig, handlerFactoryMap map[string]handlerFactory) []scanner.HandlerWithConfig {
	if !cfg.Enabled || len(cfg.Handlers) == 0 {
		return nil
	}

	var handlers []scanner.HandlerWithConfig
	for _, h := range cfg.Handlers {
		if f, ok := handlerFactoryMap[h.Type]; ok {
			handlers = append(handlers, scanner.HandlerWithConfig{
				Handler: f(),
				Config:  nil,
			})
		}
	}

	return handlers
}

func countHandlers(m map[string][]scanner.HandlerWithConfig) int {
	total := 0
	for _, hs := range m {
		total += len(hs)
	}
	return total
}

func initLogger(level string) {
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	var handler slog.Handler
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			slog.Warn("failed to open log file, using stdout", "error", err)
			handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel})
		} else {
			handler = slog.NewTextHandler(f, &slog.HandlerOptions{Level: slogLevel})
		}
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel})
	}

	l := slog.New(handler)
	slog.SetDefault(l)
}
