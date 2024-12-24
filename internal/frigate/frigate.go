package frigate

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"log"

	"github.com/geffersonFerraz/frigate-s3-telegram/internal/config"
)

type (
	frigate struct {
		cfg    *config.Config
		apiUrl string
	}

	Frigate interface {
		Events() ([]EventStruct, error)
		GetEvent(eventID string) (*EventStruct, bool, error)
		SaveThumbnail(evt EventStruct) string
		SaveClip(evt EventStruct) string
	}
)

func NewFrigate() (Frigate, error) {
	cfg := config.New()
	apiUrl := cfg.FrigateURL + "/api/events"
	return &frigate{cfg: cfg, apiUrl: apiUrl}, nil
}

func (f *frigate) Events() ([]EventStruct, error) {
	FrigateURL := f.apiUrl + "?limit=" + strconv.Itoa(f.cfg.FrigateEventLimit)

	FrigateURL += "&in_progress=1"

	// Request to Frigate
	resp, err := http.Get(FrigateURL)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Body == nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Check response status code
	if resp.StatusCode != 200 {
		return nil, err
	}

	// Read data from response
	byteValue, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse data from JSON to struct

	var events []EventStruct
	err1 := json.Unmarshal(byteValue, &events)
	if err1 != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Println("syntax error at byte offset " + strconv.Itoa(int(e.Offset)) + " URL: " + FrigateURL)
		}
		log.Println("Exit. URL: " + FrigateURL)
		return nil, err
	}

	// Return Events
	return events, nil

}

func (f *frigate) GetEvent(eventID string) (*EventStruct, bool, error) {
	FrigateURL := f.apiUrl + "/" + eventID

	// Request to Frigate
	resp, err := http.Get(FrigateURL)
	if err != nil {
		return nil, true, err
	}
	if resp == nil || resp.Body == nil {
		return nil, true, err
	}
	defer resp.Body.Close()
	// Check response status code
	if resp.StatusCode != 200 {
		return nil, true, err
	}

	// Read data from response
	byteValue, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}

	// Parse data from JSON to struct
	var event EventStruct
	err1 := json.Unmarshal(byteValue, &event)
	if err1 != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Println("syntax error at byte offset " + strconv.Itoa(int(e.Offset)) + " URL: " + FrigateURL)
		}
		log.Println("Exit. URL: " + FrigateURL)
		return nil, true, err
	}

	// Return Events
	inProgress := event.EndTime == nil
	return &event, inProgress, nil
}

func (f *frigate) SaveThumbnail(evt EventStruct) string {
	// Decode string Thumbnail base64
	dec, err := base64.StdEncoding.DecodeString(evt.Thumbnail)
	if err != nil {
		log.Fatal("Error when decode base64: " + err.Error())
	}

	// Generate uniq filename
	filename := "/tmp/" + evt.ID + ".jpg"
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal("Error when create file: " + err.Error())
	}
	defer file.Close()
	if _, err := file.Write(dec); err != nil {
		log.Fatal("Error when write file: " + err.Error())
	}
	if err := file.Sync(); err != nil {
		log.Fatal("Error when sync file: " + err.Error())
	}
	return filename
}

func (f *frigate) SaveClip(evt EventStruct) string {
	// Get config
	conf := config.New()

	// Generate clip URL
	ClipURL := conf.FrigateURL + "/api/events/" + evt.ID + "/clip.mp4"

	// Generate uniq filename
	filename := "/tmp/" + evt.ID + ".mp4"

	// Create clip file
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal("Error when create file: " + err.Error())
	}
	defer file.Close()

	// Download clip file
	resp, err := http.Get(ClipURL)
	if err != nil {
		log.Fatal("Error when download clip: " + err.Error())
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		log.Fatal("Error when download clip: " + resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatal("Error when write file: " + err.Error())
	}
	return filename
}
