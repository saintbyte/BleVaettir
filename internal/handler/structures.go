package handler

import "time"

type Reading struct {
	SensorMAC  string
	SensorName string
	Type       string
	Value      float64
	Unit       string
	Timestamp  time.Time
	Data       []byte
}
