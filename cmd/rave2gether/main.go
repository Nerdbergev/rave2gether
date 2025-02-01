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
)

func main() {
	c, err := config.LoadConfig("/home/philmacfly/Coding/go/src/github.com/Nerdbergev/rave2gether/config.toml")
	if err != nil {
		log.Fatal("Error loading config:", err)
	}
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))
	api.GetAPIRouter(c, r)

	go api.DownloadQueue()

	go api.WorkQueue()

	log.Println("Listening on port", c.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(c.Port), r))

}
