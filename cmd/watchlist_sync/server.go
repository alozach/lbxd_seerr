package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/alozach/lbxd_seerr/internal/jellyseerr"
)

func getLastRequests(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /requests request\n")

	b, err := jellyseerr.GetLastRequets()
	if err != nil {
		log.Print(err)
		return
	}

	str := string(b)
	log.Println(str)
	io.WriteString(w, str)
}

func StartServer() {
	http.HandleFunc("/requests", getLastRequests)

	err := http.ListenAndServe(":3333", nil)

	if errors.Is(err, http.ErrServerClosed) {
		log.Printf("server closed\n")
	} else if err != nil {
		log.Fatalf("error starting server: %s\n", err)
	}
}
