// internal/controller/http/router.go
package http

import (
	"net/http"

	"github.com/PaulLocust/Avito-review/config"
	"github.com/PaulLocust/Avito-review/internal/controller/http/v1"
	"github.com/PaulLocust/Avito-review/internal/usecase"
	"github.com/PaulLocust/Avito-review/pkg/logger"

	_ "github.com/PaulLocust/Avito-review/docs" // импорт сгенерированной документации swagger
	httpSwagger "github.com/swaggo/http-swagger"
)

// NewRouter создает http.Handler
// @title PR Reviewer Assignment Service
// @version 1.0.0
// @description Сервис назначения ревьюверов для Pull Request'ов
// @host localhost:8080
// @BasePath /api/v1
func NewRouter(cfg *config.Config, l logger.Interface, useCases *usecase.UseCases) http.Handler {
	mux := http.NewServeMux()
	
	// Health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Swagger UI - используем встроенную документацию из docs.go
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	
	// API v1 routes
	v1.SetupRoutes(mux, useCases, l)
	
	return mux
}