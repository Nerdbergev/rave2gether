package queue

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	ytsearch "github.com/AnjanaMadu/YTSearch"
	"github.com/Nerdbergev/rave2gether/pkg/downloader"
	"github.com/Nerdbergev/rave2gether/pkg/user"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type Queue struct {
	MusicDir string
	Entries  []Entry
}

type Entry struct {
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

func (q *Queue) AddEntry(input string, user user.User) {
	var e Entry
	e.Addedby = user
	if isValidUrl(input) {
		e.URL = input
	} else {
		results, err := ytsearch.Search(input)
		if err != nil {
			log.Println("Error searching for song: " + err.Error())
			return
		}

		e.URL = "https://www.youtube.com/watch?v=" + results[1].VideoId
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
	q.Entries = q.Entries[1:]

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
