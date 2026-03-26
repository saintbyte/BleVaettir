package handler

type Handler interface {
	Handle(reading *Reading) error
	Name() string
}
