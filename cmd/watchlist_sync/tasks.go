package main

import (
	"log"
	"time"

	"github.com/alozach/lbxd_seerr/internal/jellyseerr"
	"github.com/alozach/lbxd_seerr/internal/lxbd"
	"github.com/go-co-op/gocron/v2"
)

func dlWatchlist() {
	log.Println("Starting dl_watchlist job")

	previousData, err := lxbd.GetSavedFilms()
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

	jellyseerr.ResetRequestsCounter()

	var requests []jellyseerr.Request
	nbRequestsOK := 0
	for i, f := range watchlist {
		req := jellyseerr.CreateRequest(f, (i == 0))
		if req.Status == jellyseerr.REQ_OK {
			nbRequestsOK++
		}

		requests = append(requests, req)
		log.Printf("%s (%d): %s - %s", f.TmdbInfo.Title, f.TmdbInfo.ID, req.Status, req.Details)
	}

	log.Printf("%d requests done", nbRequestsOK)

	lxbd.SaveFilms(watchlist)
	jellyseerr.SaveRequests(requests)
}

func StartScheduler() {
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