package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sodiqit/metricpulse.git/internal/server/controllers"
	"github.com/sodiqit/metricpulse.git/internal/server/services"
)

func RegisterHandlers(mux *http.ServeMux, controllers []*controllers.Controller) {
	for _, controller := range controllers {
		for path, handler := range controller.Routes {
			mux.HandleFunc(path, handler)
		}
	}
}

func RegisterDeps() []*controllers.Controller {
	appService := services.NewAppService()
	appController := controllers.NewAppController(appService)

	return []*controllers.Controller{&appController.Controller}
}

func NewServeMux() *http.ServeMux {
	controllers := RegisterDeps()

	mux := http.NewServeMux()
	
	RegisterHandlers(mux, controllers)

	return mux
}

func RunServer(port int, mux *http.ServeMux) error {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server is starting on %s", addr)
	return http.ListenAndServe(addr, mux)
}
