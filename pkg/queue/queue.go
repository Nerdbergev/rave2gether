package queue

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Nerdbergev/rave2gether/pkg/downloader"
	"github.com/Nerdbergev/rave2gether/pkg/user"
	"github.com/google/uuid"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

const (
	HistoryFile                 = "history.json"
	sampleRate  beep.SampleRate = 44100
)

const baseURL = "https://www.googleapis.com/youtube/v3/search"

type YouTubeResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			Title string `json:"title"`
		} `json:"snippet"`
	} `json:"items"`
}

type SongInfo struct {
	Entry
	Position time.Duration `json:"position"`
	Length   time.Duration `json:"length"`
	Mutex    sync.Mutex    `json:"-"`
}

type Queue struct {
	MusicDir   string
	EntryMutex sync.Mutex
	Entries    []Entry
	SongInfo   SongInfo
}

type DownloadQueue struct {
	APIKey string
	Queue
}

type PlayQueue struct {
	Queue
	ctx        context.Context
	cancelFunc context.CancelFunc
}

type Entry struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	URL      string    `json:"url"`
	Hash     string    `json:"hash"`
	AddedBy  string    `json:"addedby"`
	AddedAt  time.Time `json:"addedat"`
	PlayedAt time.Time `json:"playedat"`
	Points   int       `json:"points"`
	votedFor map[string]int
}

func init() {
	var sampleRate beep.SampleRate = 44100
	speaker.Init(sampleRate, sampleRate.N(time.Second/10))
}

func isValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

func WriteHistory(e Entry, folder string) error {
	var history []Entry
	fp := filepath.Join(folder, HistoryFile)
	if _, err := os.Stat(fp); err == nil {
		historyFile, err := os.ReadFile(fp)
		if err != nil {
			return errors.New("Error reading history file: " + err.Error())
		}
		err = json.Unmarshal(historyFile, &history)
		if err != nil {
			return errors.New("Error unmarshalling history file: " + err.Error())
		}
	}
	history = append(history, e)
	historyJSON, err := json.MarshalIndent(history, "", "    ")
	if err != nil {
		return errors.New("Error marshalling history: " + err.Error())
	}
	err = os.WriteFile(fp, historyJSON, 0644)
	if err != nil {
		return errors.New("Error writing history file: " + err.Error())
	}

	return nil
}

func searchYouTube(query string, maxResults int, apiKey string) ([]map[string]string, error) {
	if apiKey == "" {
		return nil, errors.New("API key not set")
	}
	// Prepare the API request
	params := url.Values{}
	params.Add("part", "snippet")
	params.Add("q", query)
	params.Add("type", "video")
	params.Add("maxResults", fmt.Sprintf("%d", maxResults))
	params.Add("key", apiKey)

	// Create the request URL
	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Perform the HTTP request
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("error making API request: %v", err)
	}
	defer resp.Body.Close()

	// Check for a non-200 status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: received status code %d", resp.StatusCode)
	}

	// Decode the JSON response
	var response YouTubeResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	// Extract video titles and IDs
	results := []map[string]string{}
	for _, item := range response.Items {
		results = append(results, map[string]string{
			"title": item.Snippet.Title,
			"url":   fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.ID.VideoID),
		})
	}

	return results, nil
}

func (q *DownloadQueue) AddEntry(input string, user user.User) error {
	var e Entry
	e.votedFor = make(map[string]int)
	e.AddedBy = user.Username
	e.AddedAt = time.Now()
	if isValidUrl(input) {
		e.URL = input
		cmd := exec.Command("yt-dlp", "--print", "title", input)

		// Capture the output
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error running yt-dlp: %v\n", err)
			return errors.New("Error running yt-dlp: " + err.Error())
		}
		e.Name = strings.TrimSpace(string(out.String()))
	} else {
		result, err := searchYouTube(input, 1, q.APIKey)
		if err != nil {
			log.Println("Error searching for song: " + err.Error())
			return errors.New("Error searching for song: " + err.Error())
		}

		if len(result) == 0 {
			return errors.New("no results found")
		}

		e.Name = html.UnescapeString(result[0]["title"])
		e.URL = result[0]["url"]
	}
	log.Println("Adding song to queue: " + e.Name + " (" + e.URL + ")")
	h := sha1.New()
	h.Write([]byte(e.URL))
	e.Hash = hex.EncodeToString(h.Sum(nil))
	e.ID = uuid.New().String()
	log.Println(e)
	q.EntryMutex.Lock()
	q.Entries = append(q.Entries, e)
	q.EntryMutex.Unlock()
	return nil
}

func (q *Queue) PopEntry() Entry {
	q.EntryMutex.Lock()
	defer q.EntryMutex.Unlock()
	e := q.Entries[0]
	q.Entries = q.Entries[1:]
	return e
}

func (q *Queue) GetAllEntries() []Entry {
	q.EntryMutex.Lock()
	defer q.EntryMutex.Unlock()
	return q.Entries
}

func (q *Queue) GetEntryCount() int {
	q.EntryMutex.Lock()
	defer q.EntryMutex.Unlock()
	return len(q.Entries)
}

func (q *PlayQueue) SortEntries() {
	q.EntryMutex.Lock()
	sort.Slice(q.Entries, func(i, j int) bool {
		if q.Entries[i].Points == q.Entries[j].Points {
			return q.Entries[i].AddedAt.Before(q.Entries[j].AddedAt)
		}
		return q.Entries[i].Points > q.Entries[j].Points
	})
	q.EntryMutex.Unlock()
}

func (q *PlayQueue) SkipSong() {
	q.cancelFunc()
}

func (q *PlayQueue) PlayNext() error {
	if len(q.Entries) == 0 {
		return nil
	}

	log.Println("Trying to play next Song")

	e := q.PopEntry()

	q.SongInfo.Mutex.Lock()
	q.SongInfo.Entry = e
	q.SongInfo.Mutex.Unlock()

	log.Println("Playing next Song " + e.Hash + " " + e.Name)

	fp := filepath.Join(q.MusicDir, e.Hash) + ".mp3"

	f, err := os.Open(fp)
	if err != nil {
		return errors.New("Error opening file: " + err.Error())
	}
	defer f.Close()

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return errors.New("Error decoding file: " + err.Error())
	}
	defer streamer.Close()

	resampled := beep.Resample(4, format.SampleRate, sampleRate, streamer)
	q.ctx, q.cancelFunc = context.WithCancel(context.Background())
	defer q.cancelFunc()
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		q.cancelFunc()
	})))

	// Start a ticker to display the current position
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				position := sampleRate.D(streamer.Position())
				length := sampleRate.D(streamer.Len())
				q.SongInfo.Mutex.Lock()
				q.SongInfo.Position = position
				q.SongInfo.Length = length
				q.SongInfo.Mutex.Unlock()
			case <-q.ctx.Done():
				return
			}
		}
	}()

	<-q.ctx.Done()

	q.SongInfo.Mutex.Lock()
	q.SongInfo.Entry = Entry{}
	q.SongInfo.Position = 0
	q.SongInfo.Length = 0
	q.SongInfo.Mutex.Unlock()

	e.PlayedAt = time.Now()
	WriteHistory(e, q.MusicDir)
	log.Println("Song played: " + e.Name)
	return nil
}

func (q *PlayQueue) VoteSong(id string, upvote bool, user user.User) error {
	amount := 1
	if !upvote {
		amount = -1
	}
	for i, e := range q.Entries {
		if e.ID == id {
			q.EntryMutex.Lock()
			lastvote, ok := q.Entries[i].votedFor[user.Username]
			if ok {
				if lastvote == amount {
					q.EntryMutex.Unlock()
					return errors.New("already voted")
				}
				q.Entries[i].Points -= lastvote
			}
			q.Entries[i].votedFor[user.Username] = amount
			q.Entries[i].Points += amount
			q.EntryMutex.Unlock()
			q.SortEntries()
			return nil
		}
	}
	return errors.New("song not found")
}

func (q *Queue) DeleteSong(id string) error {
	for i, e := range q.Entries {
		if e.ID == id {
			q.EntryMutex.Lock()
			q.Entries = append(q.Entries[:i], q.Entries[i+1:]...)
			q.EntryMutex.Unlock()
			return nil
		}
	}
	return errors.New("song not found")
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	// Return true only if it exists and is not a directory
	return !info.IsDir()
}

func (q *Queue) DownloadNext() (Entry, error) {
	if len(q.Entries) == 0 {
		return Entry{}, nil
	}

	log.Println("Trying to download next Song")

	e := q.PopEntry()

	q.SongInfo.Mutex.Lock()
	q.SongInfo.Entry = e
	q.SongInfo.Mutex.Unlock()

	log.Println("Downloading next Song " + e.Hash)

	fp := filepath.Join(q.MusicDir, e.Hash) + ".mp3"
	if !fileExists(fp) {
		log.Println("Downloading to " + fp)
		err := downloader.Download(e.URL, e.Hash, q.MusicDir)
		if err != nil {
			return e, errors.New("Error downloading file: " + err.Error())
		}
	} else {
		log.Println("File already exists")
	}

	q.SongInfo.Mutex.Lock()
	q.SongInfo.Entry = Entry{}
	q.SongInfo.Mutex.Unlock()

	return e, nil
}
