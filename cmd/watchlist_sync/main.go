package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	c "github.com/alozach/lbxd_seerr/internal/config"
	"github.com/alozach/lbxd_seerr/internal/jellyseerr"
	"github.com/alozach/lbxd_seerr/internal/scrapping"
)

var scrap *scrapping.Scrapping
var config *c.Configuration

func initLogs() {
	logsDir := "/config/logs"
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

func main() {
	initLogs()

	config = c.GetConfig()

	jellyseerr.Init(config.Jellyseerr)
	jellyseerr.AddFilters(config.Jellyseerr.Filters)

	scrap = scrapping.Init(config.TMDb.ApiKey)
	defer scrapping.Deinit(scrap)

	go StartScheduler()
	go StartServer()
}
