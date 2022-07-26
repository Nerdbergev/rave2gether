package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/Nerdbergev/rave2gether/pkg/api"
	"github.com/Nerdbergev/rave2gether/pkg/config"
)

func main() {
	c, err := config.LoadConfig("config.toml")
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	r := api.GetAPIRouter(c.MusicDir)

	go api.DownloadQueue()

	go api.WorkQueue()

	log.Println("Listening on port", c.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(c.Port), r))

}
