package scanner

import "github.com/saintbyte/BleVaettir/internal/handler"

// DeviceScope map of device MAC addresses and their associated reading
type DeviceScope map[string][]handler.Reading

// Scope halders scope of device to send after process all devices
type Scope map[string]DeviceScope
