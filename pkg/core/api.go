package core

import (
	"github.com/gorilla/mux"
)

// SetupRouter loads all API
func SetupRouter(apis ...*API) *mux.Router {
	router := mux.NewRouter()

	// Global middlewares
	router.Use(LoggerInOutRequest)

	for _, api := range apis {
		coreLogger.Debug("[API] Loading API of root \"%s\"", api.root)
		(*api).LoadInRouter(router)
	}

	return router
}
