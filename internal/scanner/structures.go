package scanner

import (
	"github.com/go-ble/ble"
	"github.com/saintbyte/BleVaettir/internal/handler"
)

type HandlerWithConfig struct {
	Handler handler.Handler
	Config  *handler.HandlerConfig
}

type scanResult struct {
	mac string
	adv ble.Advertisement
}
