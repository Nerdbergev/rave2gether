package downloader

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func ytdlp(url string, path string) error {
	log.Println("Downloading", url, "to", path)
	cmd := exec.Command("yt-dlp")
	cmd.Args = append(cmd.Args, "-x")
	cmd.Args = append(cmd.Args, "--audio-format=mp3")
	cmd.Args = append(cmd.Args, url)
	cmd.Args = append(cmd.Args, "-o"+path+"")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}
	fmt.Println("Result: " + out.String())
	fmt.Println("Done downloading")
	return nil
}

func Download(url string, hash string, location string) error {
	path := filepath.Join(location, hash) + ".mp3"
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if errors.Is(err, os.ErrNotExist) {
		err := ytdlp(url, path)
		if err != nil {
			return errors.New("Error downloading: " + err.Error())
		}
		return nil
	} else {
		return errors.New("Error checking file: " + err.Error())
	}
}
