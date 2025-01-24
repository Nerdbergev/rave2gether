package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Nerdbergev/rave2gether/pkg/config"
	"github.com/Nerdbergev/rave2gether/pkg/queue"
	"github.com/Nerdbergev/rave2gether/pkg/user"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

var playlist queue.PlayQueue
var downloadlist queue.DownloadQueue

type errorResponse struct {
	Httpstatus   string `json:"httpstatus"`
	Errormessage string `json:"errormessage"`
}

type addSongRequest struct {
	Queries []string `json:"queries"`
}

type voteRequest struct {
	Upvote bool `json:"upvote"`
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
	for _, q := range req.Queries {
		if q == "" {
			continue
		}
		err = downloadlist.AddEntry(q, u)
		if err != nil {
			apierror(w, r, "Error adding song to queue: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func listQueueHandler(w http.ResponseWriter, r *http.Request) {
	j, err := json.MarshalIndent(playlist.Entries, "", "    ")
	if err != nil {
		apierror(w, r, "Error marshalling queue: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(j)
}

func getSongPositionHandler(w http.ResponseWriter, r *http.Request) {
	playlist.SongPosition.Mutex.Lock()
	posi := playlist.SongPosition
	playlist.SongPosition.Mutex.Unlock()
	j, err := json.MarshalIndent(posi, "", "    ")
	if err != nil {
		apierror(w, r, "Error marshalling queue: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(j)
}

func voteSongHandler(w http.ResponseWriter, r *http.Request) {
	var req voteRequest
	log.Println("Voting for Song")
	songid := chi.URLParam(r, "songid")
	if songid == "" {
		apierror(w, r, "No songid provided", http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		apierror(w, r, "Error decoding request: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = playlist.VoteSong(songid, req.Upvote)
	if err != nil {
		apierror(w, r, "Error voting for song: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func deleteSongHandler(w http.ResponseWriter, r *http.Request) {
	songid := chi.URLParam(r, "songid")
	if songid == "" {
		apierror(w, r, "No songid provided", http.StatusBadRequest)
		return
	}
	err := playlist.DeleteSong(songid)
	if err != nil {
		apierror(w, r, "Error deleting song: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func DownloadQueue() {
	for {
		e, err := downloadlist.DownloadNext()
		if err != nil {
			log.Printf("Error downloading Song: %v ID: %v Error: %v", e.Name, e.ID, err)
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

func getHistoryHandler(w http.ResponseWriter, r *http.Request, location string) {
	history, err := os.ReadFile(filepath.Join(location, queue.HistoryFile))
	if err != nil {
		apierror(w, r, "Error reading history file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(history)
}

func GetAPIRouter(location string, apiKey string, mode config.Operatingmode) *chi.Mux {
	playlist.Queue.MusicDir = location
	downloadlist.Queue.MusicDir = location
	downloadlist.APIKey = apiKey

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))
	r.Route("/api", func(r chi.Router) {
		r.Route("/queue", func(r chi.Router) {
			r.Get("/", listQueueHandler)
			r.Post("/", addtoQueueHandler)
			r.Get("/position", getSongPositionHandler)
			r.Route("/{songid}", func(r chi.Router) {
				if mode == config.Voting || mode == config.UserVoting || mode == config.UserCoins {
					r.Post("/vote", voteSongHandler)
				}
				r.Delete("/", deleteSongHandler)
			})
		})
		r.Route("/history", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				getHistoryHandler(w, r, location)
			})
		})
	})

	return r
}
