package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Nerdbergev/rave2gether/pkg/api"
	"github.com/Nerdbergev/rave2gether/pkg/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/spf13/pflag"
)

func main() {
	configPath := pflag.String("config", "config.toml", "config path")
	c, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Error loading config:", err)
	}
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	api.GetAPIRouter(c, r)

	go api.PrepareQueue()

	go api.DownloadQueue()

	go api.WorkQueue()

	log.Println("Listening on port", c.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(c.Port), r))

}
