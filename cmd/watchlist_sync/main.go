package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	c "github.com/alozach/lbxd_seerr/internal/config"
	"github.com/alozach/lbxd_seerr/internal/jellyseerr"
	"github.com/alozach/lbxd_seerr/internal/lxbd"
	"github.com/alozach/lbxd_seerr/internal/scrapping"
	"github.com/go-co-op/gocron/v2"
)

var scrap *scrapping.Scrapping
var config *c.Configuration

func initLogs() {
	logsDir := "/logs"
	if err := os.MkdirAll(logsDir, os.ModePerm); err != nil {
		log.Fatalln("Failed to create logs folder: ", err)
	}

	currentTime := time.Now()
	logFilePath := filepath.Join(logsDir, currentTime.Format("2006_01_02_15_04_05")+".txt")
	logFile, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal("Failed to open logs file: ", err)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
}

func dlWatchlist() {
	log.Println("Starting dl_watchlist job")

	previousDataFilename := "/app/data/last_requests.txt"
	previousData, err := lxbd.GetSavedFilms(previousDataFilename)
	if err != nil {
		log.Println("Failed to get previously saved data")
	}

	if err := scrap.LxbdAcceptCookies(); err != nil {
		return
	}

	if err := scrap.LxbdLogIn(config.Lxbd.Username, config.Lxbd.Password); err != nil {
		return
	}

	watchlist, err := getWatchlist(scrap, previousData)
	if err != nil {
		return
	}

	log.Printf("Got %d films in watchlist", len(watchlist))

	nbRequests := 0
	for i, f := range watchlist {
		if config.Jellyseerr.RequestsLimit > 0 && nbRequests >= config.Jellyseerr.RequestsLimit {
			break
		}

		req := jellyseerr.CreateRequest(f, (i==0))
		if req.Status == jellyseerr.REQ_OK {
			nbRequests++
		}
		log.Printf("%s (%d): %s - %s", f.TmdbInfo.Title, f.TmdbInfo.ID, req.Status, req.Details)
	}

	log.Printf("%d requests done", nbRequests)

	lxbd.SaveFilms(watchlist, previousDataFilename)
}

func main() {
	initLogs()

	config = c.GetConfig()

	jellyseerr.Init(config.Jellyseerr.ApiKey, config.Jellyseerr.BaseUrl)
	jellyseerr.AddFilters(config.Jellyseerr.Filters)

	scrap = scrapping.Init(config.TMDb.ApiKey)
	defer scrapping.Deinit(scrap)

	location, _ := time.LoadLocation("Europe/Paris")
	sched, err := gocron.NewScheduler(gocron.WithLocation(location))
	if err != nil {
		log.Fatalln("Failed to create scheduler: ", err)
	}

	if config.Tasks.DLWatchlist != "disabled" {
		j, err := sched.NewJob(
			gocron.CronJob(config.Tasks.DLWatchlist, false),
			gocron.NewTask(
				dlWatchlist,
			),
			gocron.WithName("dl_watchlist"),
			gocron.WithSingletonMode(gocron.LimitModeReschedule),
		)
		if err != nil {
			log.Fatalln("Failed to create job: ", err)
		}

		log.Printf("Created job %s (%s)", j.Name(), j.ID())
	} else {
		log.Println("dl_watchlist task is disabled")
	}

	log.Println("Starting scheduler")
	sched.Start()
	select {}

	/* 	err = sched.Shutdown()
	   	if err != nil {
	   		// handle error
	   	} */
}
