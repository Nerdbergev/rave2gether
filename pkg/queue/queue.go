package queue

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/Nerdbergev/rave2gether/pkg/downloader"
	"github.com/Nerdbergev/rave2gether/pkg/user"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

const (
	HistoryFile = "history.log"
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

type Queue struct {
	APIKey   string
	MusicDir string
	Entries  []Entry
}

type Entry struct {
	Name    string
	URL     string
	Hash    string
	Addedby user.User
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

func WriteHistory(name string, folder string) error {
	f, err := os.Create(filepath.Join(folder, HistoryFile))
	if err != nil {
		return errors.New("Error creating history file: " + err.Error())
	}
	defer f.Close()

	stamp := time.Now().Format("2006-01-02 15:04:05")

	_, err = f.WriteString(stamp + ":" + name + "\n")
	if err != nil {
		return errors.New("Error writing history file: " + err.Error())
	}

	return nil
}

func searchYouTube(query string, maxResults int, apiKey string) ([]map[string]string, error) {
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

func (q *Queue) AddEntry(input string, user user.User) {
	var e Entry
	e.Addedby = user
	if isValidUrl(input) {
		e.URL = input
		e.Name = input
	} else {
		result, err := searchYouTube(input, 1, q.APIKey)
		if err != nil {
			log.Println("Error searching for song: " + err.Error())
			return
		}

		for i, r := range result {
			fmt.Println(i, r)
		}

		e.Name = result[0]["title"]
		e.URL = result[0]["url"]
	}
	h := sha1.New()
	h.Write([]byte(e.URL))
	e.Hash = hex.EncodeToString(h.Sum(nil))
	q.Entries = append(q.Entries, e)
}

func (q *Queue) RemoveEntry(index int) {
	q.Entries = append(q.Entries[:index], q.Entries[index+1:]...)
}

func (q *Queue) GetAllEntries() []Entry {
	return q.Entries
}

func (q *Queue) PlayNext() error {
	if len(q.Entries) == 0 {
		return nil
	}

	log.Println("Trying to play next Song")

	e := q.Entries[0]

	log.Println("Playing next Song " + e.Hash)

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

	resampled := beep.Resample(4, format.SampleRate, 44100, streamer)

	done := make(chan bool)
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		done <- true
	})))

	<-done
	WriteHistory(e.Name, q.MusicDir)
	q.Entries = q.Entries[1:]
	return nil
}

func (q *Queue) DownloadNext() (Entry, error) {
	if len(q.Entries) == 0 {
		return Entry{}, nil
	}

	log.Println("Trying to download next Song")

	e := q.Entries[0]

	log.Println("Downloading next Song " + e.Hash)
	q.Entries = q.Entries[1:]

	fp := filepath.Join(q.MusicDir, e.Hash) + ".mp3"
	log.Println("Downloading to " + fp)
	err := downloader.Download(e.URL, e.Hash, q.MusicDir)
	if err != nil {
		return Entry{}, errors.New("Error downloading file: " + err.Error())
	}
	return e, nil
}
