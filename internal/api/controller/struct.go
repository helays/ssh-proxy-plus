package controller

import "github.com/helays/ssh-proxy-plus/configs"

type Controller struct {
	router configs.Router
}

func New(router configs.Router) *Controller {
	return &Controller{
		router: router,
	}
}
