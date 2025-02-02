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
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

var playlist queue.PlayQueue
var downloadlist queue.DownloadQueue
var tokenAuth *jwtauth.JWTAuth
var userdb user.UserDB
var idleSleep = 500

const maxSleep = 5000

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
	_, claims, _ := jwtauth.FromContext(r.Context())
	username, _ := claims["username"].(string)
	if username == "" {
		username = "Fick Hans"
	}
	u := user.User{Username: username}
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
	var ee []queue.Entry
	ee = append(ee, playlist.Entries...)
	j, err := json.MarshalIndent(ee, "", "    ")
	if err != nil {
		apierror(w, r, "Error marshalling queue: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(j)
}

func getCurrentSongHandler(w http.ResponseWriter, r *http.Request) {
	playlist.SongInfo.Mutex.Lock()
	info := currentSongResponse{playlist.SongInfo.Name, playlist.SongInfo.Position, playlist.SongInfo.Length, playlist.SongInfo.AddedBy, playlist.SongInfo.AddedAt, playlist.SongInfo.Points}
	playlist.SongInfo.Mutex.Unlock()
	j, err := json.MarshalIndent(info, "", "    ")
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
		if downloadlist.GetEntryCount() == 0 {
			time.Sleep(time.Millisecond * time.Duration(idleSleep))
			if idleSleep < maxSleep {
				idleSleep += 500
			}
			continue
		}
		e, err := downloadlist.DownloadNext()
		if err != nil {
			log.Printf("Error downloading Song: %v ID: %v Error: %v", e.Name, e.ID, err)
		} else {
			if e.Hash != "" {
				playlist.Entries = append(playlist.Entries, e)
			}
		}
		idleSleep = 500
	}
}

func WorkQueue() {
	for {
		if playlist.GetEntryCount() == 0 {
			time.Sleep(time.Millisecond * time.Duration(idleSleep))
			if idleSleep < maxSleep {
				idleSleep += 500
			}
			continue
		}
		err := playlist.PlayNext()
		if err != nil {
			log.Println("Error playing next:", err)
		}
		idleSleep = 500
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
	claims := map[string]interface{}{"username": u.Username}
	jwtauth.SetExpiryIn(claims, time.Hour)
	jwtauth.SetIssuedNow(claims)
	_, tokenString, err := tokenAuth.Encode(claims)
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

func apiRefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	tokenstr := jwtauth.TokenFromHeader(r)
	token, err := tokenAuth.Decode(tokenstr)
	if err != nil {
		tokenInvalid(w, r)
		return
	}
	cl := token.PrivateClaims()
	username, ok := cl["username"].(string)
	if !ok || !userdb.DoesUserExist(username) {
		tokenInvalid(w, r)
		return
	}

	claims := map[string]interface{}{"username": username}
	jwtauth.SetExpiryIn(claims, time.Hour)
	jwtauth.SetIssuedNow(claims)
	_, tokenString, err := tokenAuth.Encode(claims)
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

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	users := userdb.ListUsers()
	j, err := json.MarshalIndent(users, "", "    ")
	if err != nil {
		apierror(w, r, "Error marshalling users: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(j)
}

func addUserHandler(w http.ResponseWriter, r *http.Request) {
	var req addUserRequest
	log.Println("Adding user")
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		apierror(w, r, "Error decoding request: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = userdb.AddUser(req.Username, req.Password, user.Userright(req.Right))
	if err != nil {
		apierror(w, r, "Error adding user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req passwordChangeRequest
	log.Println("Changing password")
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		apierror(w, r, "Error decoding request: "+err.Error(), http.StatusBadRequest)
		return
	}
	_, claims, _ := jwtauth.FromContext(r.Context())
	username, _ := claims["username"].(string)
	u, err := userdb.GetUser(username)
	if err != nil {
		apierror(w, r, "Error getting user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = u.SetPassword(req.NewPassword)
	if err != nil {
		apierror(w, r, "Error setting password: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func getCoinsHandler(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	username, _ := claims["username"].(string)
	u, err := userdb.GetUser(username)
	if err != nil {
		apierror(w, r, "Error getting user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	j, err := json.MarshalIndent(coinsResponse{u.Coins}, "", "    ")
	if err != nil {
		apierror(w, r, "Error marshalling coins: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(j)
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	user := chi.URLParam(r, "username")
	if user == "" {
		apierror(w, r, "No user provided", http.StatusBadRequest)
		return
	}
	err := userdb.RemoveUser(user)
	if err != nil {
		apierror(w, r, "Error deleting user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req addUserRequest
	u := chi.URLParam(r, "username")
	if u == "" {
		apierror(w, r, "No user provided", http.StatusBadRequest)
		return
	}
	log.Println("Updating user")
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		apierror(w, r, "Error decoding request: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = userdb.UpdateUser(u, req.Password, user.Userright(req.Right))
	if err != nil {
		apierror(w, r, "Error updating user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func setCoinsHandler(w http.ResponseWriter, r *http.Request) {
	user := chi.URLParam(r, "username")
	if user == "" {
		apierror(w, r, "No user provided", http.StatusBadRequest)
		return
	}
	var req coinsResponse
	log.Println("Setting coins")
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		apierror(w, r, "Error decoding request: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = userdb.SetUserCoins(user, req.Coins)
	if err != nil {
		apierror(w, r, "Error setting coins: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func addCoinsHandler(w http.ResponseWriter, r *http.Request) {
	user := chi.URLParam(r, "username")
	if user == "" {
		apierror(w, r, "No user provided", http.StatusBadRequest)
		return
	}
	var req coinsResponse
	log.Println("Adding coins")
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		apierror(w, r, "Error decoding request: "+err.Error(), http.StatusBadRequest)
		return
	}
	u, err := userdb.GetUser(user)
	if err != nil {
		apierror(w, r, "Error getting user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	u.Coins += req.Coins
	err = userdb.SetUserCoins(user, u.Coins)
	if err != nil {
		apierror(w, r, "Error setting coins: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func GetAPIRouter(cfg config.Config, r *chi.Mux) {
	playlist.Queue.MusicDir = cfg.FileDir
	downloadlist.Queue.MusicDir = cfg.FileDir
	downloadlist.APIKey = cfg.YTApiKey

	if cfg.Mode > config.Voting {
		tokenAuth = jwtauth.New("HS256", []byte(cfg.Secret), nil)
	}

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

	r.Route("/api", func(r chi.Router) {
		if cfg.Mode > config.Voting {
			r.Post("/token", apiGetTokenHandler)
			r.Post("/refreshtoken", apiRefreshTokenHandler)
		}
		r.Route("/queue", func(r chi.Router) {
			r.Get("/", listQueueHandler)
			r.Get("/current", getCurrentSongHandler)
			r.Group(func(r chi.Router) {
				if cfg.Mode > config.Voting {
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
		if cfg.Mode > config.Voting {
			r.Route("/users", func(r chi.Router) {
				r.Get("/", getUsersHandler)
				r.Post("/{username}/password", changePasswordHandler)
				r.Get("/{username}/coins", getCoinsHandler)
				r.Group(func(r chi.Router) {
					r.Use(Authenticator(tokenAuth, user.Moderator))
					r.Post("/{username}/coins", setCoinsHandler)
					r.Post("/{username}/addcoins", addCoinsHandler)
				})
				r.Group(func(r chi.Router) {
					r.Use(Authenticator(tokenAuth, user.Admin))
					r.Post("/", addUserHandler)
					r.Route("/{username}", func(r chi.Router) {
						r.Delete("/", deleteUserHandler)
						r.Put("/", updateUserHandler)
					})
				})
			})
		}
		r.Route("/history", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				getHistoryHandler(w, r, cfg.FileDir)
			})
		})
	})

}
