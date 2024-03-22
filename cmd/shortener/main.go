package main

import (
	"net/http"

	"github.com/Longreader/go-shortener-url.git/config"
	"github.com/Longreader/go-shortener-url.git/internal/app/auth"
	"github.com/Longreader/go-shortener-url.git/internal/app/handlers"
	"github.com/Longreader/go-shortener-url.git/internal/app/middlewares"
	"github.com/Longreader/go-shortener-url.git/internal/app/routers"
	"github.com/Longreader/go-shortener-url.git/internal/storage"
	"github.com/sirupsen/logrus"
)

func main() {

	logrus.StandardLogger().Level = logrus.DebugLevel
	logrus.SetFormatter(&logrus.JSONFormatter{})

	cfg := config.NewConfig()

	s, err := storage.NewStorager(cfg)
	if err != nil {
		logrus.Fatal("Error connection drop", err)
	}

	h := handlers.NewHandler(
		s,
		cfg.ServerBaseURL,
	)

	a := auth.NewAuth(
		cfg,
	)

	m := middlewares.NewMiddlewares(
		cfg,
		a,
	)

	r := routers.NewRouter(
		m,
		h,
	)

	//s.RunDelete()

	http.Handle("/", r)

	logrus.Info("Start service")
	logrus.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
