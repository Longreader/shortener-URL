package routers

import (
	"github.com/Longreader/go-shortener-url.git/internal/app/handlers"
	"github.com/Longreader/go-shortener-url.git/internal/app/middlewares"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func NewRouter(m middlewares.Middlewares, h *handlers.Handler) chi.Router {

	r := chi.NewRouter()

	r.Use(middleware.Compress(5))
	r.Use(m.DecompresMiddleware)
	r.Use(m.UserCookie)
	r.Use(middleware.Recoverer)

	r.Get("/{id:[0-9A-Za-z]+}", h.IDGetHandler)
	r.Post("/", h.ShortenerURLHandler)
	r.Post("/api/shorten", h.APIShortenerURLHandler)

	return r
}