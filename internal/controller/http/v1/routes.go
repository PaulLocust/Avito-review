// internal/controller/http/v1/routes.go
package v1

import (
	"net/http"

	"github.com/PaulLocust/Avito-review/internal/usecase"
	"github.com/PaulLocust/Avito-review/pkg/logger"
)

func SetupRoutes(mux *http.ServeMux, useCases *usecase.UseCases, l logger.Interface) {
	// Инициализируем хендлеры
	teamHandlers := newTeamHandlers(useCases.Team, l)
	userHandlers := newUserHandlers(useCases.User, l)
	prHandlers := newPRHandlers(useCases.PR, l)

	// Teams
	mux.HandleFunc("POST /api/v1/team/add", teamHandlers.addTeam)
	mux.HandleFunc("GET /api/v1/team/get", teamHandlers.getTeam)
	
	// Users
	mux.HandleFunc("POST /api/v1/users/setIsActive", userHandlers.setIsActive)
	mux.HandleFunc("GET /api/v1/users/getReview", userHandlers.getReviews)
	
	// Pull Requests
	mux.HandleFunc("POST /api/v1/pullRequest/create", prHandlers.createPR)
	mux.HandleFunc("POST /api/v1/pullRequest/merge", prHandlers.mergePR)
	mux.HandleFunc("POST /api/v1/pullRequest/reassign", prHandlers.reassignReviewer)
}