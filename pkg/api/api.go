package api

import (
	"log"
	"net/http"
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

func Authenticator(ja *jwtauth.JWTAuth, ur user.Userright) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, claims, err := jwtauth.FromContext(r.Context())
			if err != nil {
				tokenInvalid(w, r)
				return
			}
			isrefresh, ok := claims["refresh"].(bool)
			if !ok || isrefresh {
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
		r.Get("/mode", func(w http.ResponseWriter, r *http.Request) {
			apiModeHandler(w, r, cfg.Mode)
		})
		if cfg.Mode > config.Voting {
			r.Post("/token", apiGetTokenHandler)
			r.Post("/refreshtoken", apiRefreshTokenHandler)
		}
		r.Route("/queue", func(r chi.Router) {
			r.Get("/", listQueueHandler)
			r.Get("/download", listDownloadQueueHandler)
			r.Get("/current", getCurrentSongHandler)
			r.Group(func(r chi.Router) {
				if cfg.Mode > config.Voting {
					r.Use(jwtauth.Verifier(tokenAuth))
					r.Use(Authenticator(tokenAuth, user.Unprivileged))
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
			r.Route("/self", func(r chi.Router) {
				r.Use(jwtauth.Verifier(tokenAuth))
				r.Use(Authenticator(tokenAuth, user.Unprivileged))
				r.Get("/", selfHandler)
			})
			r.Route("/users", func(r chi.Router) {
				r.Use(jwtauth.Verifier(tokenAuth))
				r.Use(Authenticator(tokenAuth, user.Unprivileged))
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
