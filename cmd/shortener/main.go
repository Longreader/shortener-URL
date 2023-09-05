package main

import (
	"flag"
	"net/http"

	"github.com/Longreader/go-shortener-url.git/config"
	"github.com/Longreader/go-shortener-url.git/internal/app"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

func main() {

	flag.Parse()

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)

	logrus.StandardLogger().Level = logrus.DebugLevel

	r.Get("/{id:[0-9A-Za-z]+}", app.IDGetHandler)
	r.Post("/", app.ShortenerURLHandler)
	r.Post("/api/shorten", app.APIShortenerURLHandler)

	http.Handle("/", r)

	logrus.Fatal(http.ListenAndServe(config.GetAddress(), r))
}
