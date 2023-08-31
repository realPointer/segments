package v1

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/realPointer/segments/internal/service"
	"github.com/realPointer/segments/pkg/logger"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewRouter(handler chi.Router, l logger.Interface, services *service.Services) {
	handler.Use(middleware.Logger)
	handler.Use(middleware.Recoverer)
	handler.Use(middleware.Timeout(60 * time.Second))

	handler.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong!"))
	})

	handler.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	handler.Route("/v1", func(r chi.Router) {
		r.Mount("/user/{user_id:[0-9]+}", NewUserRouter(services.User, l))
		r.Mount("/segment", NewSegmentRouter(services.Segment, l))
	})
}
