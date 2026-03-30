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

func buildObjectHandlers(cfg *config.Config, store *storage.Storage) (map[string][]handler.Handler, []handler.Handler) {
	result := make(map[string][]handler.Handler)
	globalHandlers := buildGlobalHandlers(cfg.Handlers, store)
	handlerMap := buildHandlerMap(cfg.Handlers, store)
	unknownHandlers := buildUnknownHandlers(cfg.UnknownObjects, handlerMap, store)

	for _, obj := range cfg.BLEObjects {
		if len(obj.Handlers) == 0 {
			result[obj.MAC] = globalHandlers
		} else {
			var objHandlers []handler.Handler
			for _, h := range obj.Handlers {
				if ih, ok := handlerMap[h.Type]; ok {
					objHandlers = append(objHandlers, ih)
				}
			}
			result[obj.MAC] = objHandlers
		}

		if len(result[obj.MAC]) == 0 {
			result[obj.MAC] = []handler.Handler{handler.NewDBHandler(store)}
		}
	}

	return result, unknownHandlers
}

func buildGlobalHandlers(handlerCfgs []config.HandlerConfig, store *storage.Storage) []handler.Handler {
	var handlers []handler.Handler

	for _, h := range handlerCfgs {
		switch h.Type {
		case "db":
			if h.DB != nil && h.DB.Enabled {
				handlers = append(handlers, handler.NewDBHandler(store))
			}
		case "http":
			if h.HTTP != nil && h.HTTP.Enabled {
				handlers = append(handlers, handler.NewHTTPHandler(h.HTTP.Endpoint, h.HTTP.APIKey))
			}
		case "narodmon":
			if h.Narodmon != nil && h.Narodmon.Enabled {
				handlers = append(handlers, handler.NewNarodmonHandler())
			}
		case "log":
			handlers = append(handlers, handler.NewLogHandler())
		case "datacake":
			handlers = append(handlers, handler.NewDataCakeHandler())
		}
	}

	return handlers
}

func buildHandlerMap(handlerCfgs []config.HandlerConfig, store *storage.Storage) map[string]handler.Handler {
	result := make(map[string]handler.Handler)

	for _, h := range handlerCfgs {
		switch h.Type {
		case "db":
			if h.DB != nil && h.DB.Enabled {
				result["db"] = handler.NewDBHandler(store)
			}
		case "http":
			if h.HTTP != nil && h.HTTP.Enabled {
				result["http"] = handler.NewHTTPHandler(h.HTTP.Endpoint, h.HTTP.APIKey)
			}
		case "narodmon":
			if h.Narodmon != nil && h.Narodmon.Enabled {
				result["narodmon"] = handler.NewNarodmonHandler()
			}
		case "log":
			result["log"] = handler.NewLogHandler()
		}
	}

	return result
}

func buildUnknownHandlers(cfg config.UnknownObjectsConfig, handlerMap map[string]handler.Handler, store *storage.Storage) []handler.Handler {
	if !cfg.Enabled || len(cfg.Handlers) == 0 {
		return nil
	}

	var handlers []handler.Handler
	for _, h := range cfg.Handlers {
		if ih, ok := handlerMap[h.Type]; ok {
			handlers = append(handlers, ih)
		}
	}

	return handlers
}

func countHandlers(m map[string][]handler.Handler) int {
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
