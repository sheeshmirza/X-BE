package handlers

import "X-BE/internal/service"

type Handlers struct {
	services *service.Services
}

func NewHandlers(services *service.Services) *Handlers {
	return &Handlers{services: services}
}
