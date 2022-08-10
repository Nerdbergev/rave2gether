package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Nerdbergev/rave2gether/pkg/queue"
	"github.com/Nerdbergev/rave2gether/pkg/user"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

var playlist queue.Queue
var downloadlist queue.Queue

type errorResponse struct {
	Httpstatus   string `json:"httpstatus"`
	Errormessage string `json:"errormessage"`
}

type addSongRequest struct {
	Query string `json:"query"`
}

func apierror(w http.ResponseWriter, r *http.Request, err string, httpcode int) {
	log.Println(err)
	er := errorResponse{strconv.Itoa(httpcode), err}
	j, erro := json.Marshal(&er)
	if erro != nil {
		return
	}
	http.Error(w, string(j), httpcode)
}

func addtoQueueHandler(w http.ResponseWriter, r *http.Request) {
	var req addSongRequest
	log.Println("Adding song to queue")
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		apierror(w, r, "Error decoding request: "+err.Error(), http.StatusBadRequest)
		return
	}
	u := user.User{Username: "Fick Hans"}
	downloadlist.AddEntry(req.Query, u)
	w.WriteHeader(http.StatusOK)
}

func listQueueHandler(w http.ResponseWriter, r *http.Request) {
	j, err := json.MarshalIndent(playlist, "", "    ")
	if err != nil {
		apierror(w, r, "Error marshalling queue: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(j)
}

func DownloadQueue() {
	for {
		e, err := downloadlist.DownloadNext()
		if err != nil {
			log.Println("Error downloading next song:", err)
		} else {
			if e.Hash != "" {
				playlist.Entries = append(playlist.Entries, e)
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func WorkQueue() {
	for {
		err := playlist.PlayNext()
		if err != nil {
			log.Println("Error playing next:", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func GetAPIRouter(location string) *chi.Mux {
	playlist = queue.Queue{MusicDir: location}
	downloadlist = queue.Queue{MusicDir: location}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/queue", func(r chi.Router) {
		r.Get("/", listQueueHandler)
		r.Post("/", addtoQueueHandler)
	})
	return r
}
