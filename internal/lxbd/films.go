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

const filmsFilename = "/app/data/films.txt"

func SaveFilms(films []Film) error {
	jsonData, err := json.Marshal(films)
	if err != nil {
		return err
	}

	split := strings.Split(filmsFilename, "/")
	dirName := fmt.Sprint("/", filepath.Join(split[:len(split)-1]...))
	if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
		log.Println("Failed to save request data: ", err)
		return err
	}

	file, err := os.Create(filmsFilename)
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

func GetSavedFilms() ([]Film, error) {
	file, err := os.Open(filmsFilename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var films []Film
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&films); err != nil {
		return nil, err
	}
	return films, nil
}
