package main

import (
	"fmt"
	"log"
	"os"
	"time"

	c "github.com/alozach/lbxd_seerr/internal/config"
	"github.com/alozach/lbxd_seerr/internal/jellyseerr"
	"github.com/alozach/lbxd_seerr/internal/lxbd"
	"github.com/alozach/lbxd_seerr/internal/scrapping"
)

func initLogs() {
	currentTime := time.Now()

	logFilePath := fmt.Sprint("./logs/", currentTime.Format("2006_01_02_15_04_05"), ".txt")
	logFile, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
}

func main() {
	initLogs()

	config := c.GetConfig()

	previousDataFilename := "./data/last_requests.txt"

	previousData, err := lxbd.GetSavedFilms(previousDataFilename)
	if err != nil {
		log.Println("Failed to get previously saved data")
	}

	if err := jellyseerr.Init(config.Jellyseerr.ApiKey, config.Jellyseerr.BaseUrl); err != nil {
		os.Exit(1)
	}
	jellyseerr.AddFilters(config.Jellyseerr.Filters)

	s := scrapping.Init(config.TMDb.ApiKey)
	defer scrapping.Deinit(s)

	if err := s.LxbdAcceptCookies(); err != nil {
		os.Exit(1)
	}

	if err := s.LxbdLogIn(config.Lxbd.Username, config.Lxbd.Password); err != nil {
		os.Exit(1)
	}

	watchlist, err := getWatchlist(s, previousData)
	if err != nil {
		os.Exit(1)
	}

	log.Printf("Got %d films in watchlist", len(watchlist))

	nbRequests := 0
	for _, f := range watchlist {
		if config.Jellyseerr.RequestsLimit > 0 && nbRequests >= config.Jellyseerr.RequestsLimit {
			break
		}

		if req_ok := jellyseerr.CreateRequest(f); req_ok {
			nbRequests++
		}
	}

	log.Printf("%d requests done", nbRequests)

	lxbd.SaveFilms(watchlist, previousDataFilename)
}
