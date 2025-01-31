package api

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Nerdbergev/rave2gether/pkg/config"
	"github.com/Nerdbergev/rave2gether/pkg/queue"
	"github.com/Nerdbergev/rave2gether/pkg/user"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

var playlist queue.PlayQueue
var downloadlist queue.DownloadQueue
var tokenAuth *jwtauth.JWTAuth
var userdb user.UserDB

type errorResponse struct {
	Httpstatus   string `json:"httpstatus"`
	Errormessage string `json:"errormessage"`
	RequestURL   string `json:"requesturl"`
}

type addSongRequest struct {
	Queries []string `json:"queries"`
}

type voteRequest struct {
	Upvote bool `json:"upvote"`
}

type positionResponse struct {
	Position time.Duration `json:"position"`
	Length   time.Duration `json:"length"`
}

type authResponse struct {
	Token string `json:"token"`
}

func apierror(w http.ResponseWriter, r *http.Request, err string, httpcode int) {
	log.Println(err)
	er := errorResponse{strconv.Itoa(httpcode), err, r.URL.Path}
	j, erro := json.Marshal(&er)
	if erro != nil {
		return
	}
	http.Error(w, string(j), httpcode)
}

func loginUnsucessfull(w http.ResponseWriter, r *http.Request) {
	apierror(w, r, "Login unsucessfull", http.StatusUnauthorized)
}

func tokenInvalid(w http.ResponseWriter, r *http.Request) {
	apierror(w, r, "Token invalid", http.StatusUnauthorized)
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
	posi := positionResponse{playlist.SongPosition.Position, playlist.SongPosition.Length}
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

func skipSongHandler(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	username, _ := claims["username"].(string)
	log.Println("User", username, "skipped song")
	playlist.SkipSong()
	w.WriteHeader(http.StatusOK)
}

func apiGetTokenHandler(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	s := strings.Split(auth, " ")
	if len(s) != 2 {
		log.Println("Wrong auth format: ", auth)
		loginUnsucessfull(w, r)
		return
	}
	if strings.ToLower(s[0]) != "basic" {
		log.Println("Wrong auth type: ", s[0])
		loginUnsucessfull(w, r)
		return
	}
	data, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		log.Println("Error decoding base64:", err)
		loginUnsucessfull(w, r)
		return
	}
	s = strings.Split(string(data), ":")
	if len(s) != 2 {
		log.Println("Wrong user:pass format: ", string(data))
		loginUnsucessfull(w, r)
		return
	}
	u, err := userdb.GetUser(s[0])
	if err != nil {
		log.Println("Error getting user:", err)
		loginUnsucessfull(w, r)
		return
	}
	if !u.CheckPassword(s[1]) {
		log.Println("Wrong password")
		loginUnsucessfull(w, r)
		return
	}
	_, tokenString, err := tokenAuth.Encode(map[string]interface{}{"username": u.Username, "exp": time.Now().Add(time.Hour).Unix()})
	if err != nil {
		apierror(w, r, "Error encoding token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	j, err := json.MarshalIndent(authResponse{tokenString}, "", "    ")
	if err != nil {
		apierror(w, r, "Error marshalling token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(j)
}

func Authenticator(ja *jwtauth.JWTAuth, ur user.Userright) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, claims, err := jwtauth.FromContext(r.Context())
			if err != nil {
				tokenInvalid(w, r)
				return
			}

			username, ok := claims["username"].(string)
			if !ok {
				tokenInvalid(w, r)
				return
			}
			u, err := userdb.GetUser(username)
			if err != nil {
				tokenInvalid(w, r)
				return
			}
			if u.Right < ur {
				tokenInvalid(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetAPIRouter(cfg config.Config) *chi.Mux {
	playlist.Queue.MusicDir = cfg.FileDir
	downloadlist.Queue.MusicDir = cfg.FileDir
	downloadlist.APIKey = cfg.YTApiKey

	if cfg.Mode > config.Voting {
		err := userdb.LoadFromFile(cfg.FileDir + "/users.txt")
		if err != nil {
			log.Fatalln("Error loading userdb:", err)
		}
		if cfg.Mode == config.UserCoins {
			for _, u := range userdb.ListUsers() {
				userdb.SetUserCoins(u.Username, 10)
			}
		}
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))
	r.Route("/api", func(r chi.Router) {
		if cfg.Mode > config.Voting {
			r.Post("/token", apiGetTokenHandler)
		}
		r.Route("/queue", func(r chi.Router) {
			r.Get("/", listQueueHandler)
			r.Get("/position", getSongPositionHandler)
			r.Group(func(r chi.Router) {
				if cfg.Mode > config.Voting {
					tokenAuth = jwtauth.New("HS256", []byte(cfg.Secret), nil)
					r.Use(jwtauth.Verifier(tokenAuth))
					r.Use(jwtauth.Authenticator(tokenAuth))
				}
				r.Post("/", addtoQueueHandler)
				r.Group(func(r chi.Router) {
					if cfg.Mode > config.Voting {
						r.Use(Authenticator(tokenAuth, user.Moderator))
					}
					r.Post("/skip", skipSongHandler)
				})
				r.Route("/{songid}", func(r chi.Router) {
					if cfg.Mode > config.Simple {
						r.Post("/vote", voteSongHandler)
					}
					r.Group(func(r chi.Router) {
						if cfg.Mode > config.Voting {
							r.Use(Authenticator(tokenAuth, user.Moderator))
						}
						r.Delete("/", deleteSongHandler)
					})
				})

			})
		})
		r.Route("/history", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				getHistoryHandler(w, r, cfg.FileDir)
			})
		})
	})

	return r
}
