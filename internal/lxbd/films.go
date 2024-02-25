package lxbd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ryanbradynd05/go-tmdb"
)

type Film struct {
	Lid          int         `json:"lid"`
	TmdbId       int         `json:"tmdbId"`
	LxbdEndpoint string      `json:"link"`
	VODAvailable bool        `json:"vod_available"`
	TmdbInfo     *tmdb.Movie `json:"tmdb_info"`
}

func SaveFilms(films []Film, filename string) error {
	jsonData, err := json.Marshal(films)
	if err != nil {
		return err
	}

	split := strings.Split(filename, "/")
	dirName := fmt.Sprint("/", filepath.Join(split[:len(split)-1]...))
	if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
		log.Println("Failed to save request data: ", err)
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Println("Failed to save request data: ", err)
		return err
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		log.Println("Failed to save request data: ", err)
		return err
	}

	log.Println("Films data successfully saved")
	return nil
}

func GetSavedFilms(filename string) ([]Film, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode the JSON data
	var films []Film
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&films); err != nil {
		return nil, err
	}
	return films, nil
}
