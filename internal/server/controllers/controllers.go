package controllers

import (
	"net/http"

	"github.com/sodiqit/metricpulse.git/internal/server/services"
)

type Controller struct {
	Routes map[string]http.HandlerFunc
}

type AppController struct {
	Controller
}

func MainHandler(appService services.IAppService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(appService.SayHello()))
	}
}

func NewAppController(appService services.AppService) *AppController {
	controller := &AppController{}

	controller.Routes = map[string]http.HandlerFunc{
		"/": MainHandler(appService),
	}

	return controller
}
