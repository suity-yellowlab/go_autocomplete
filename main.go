package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"time"
)

// storing id is not strictly needed at the moment. May be useful later.
type AutocompleteTreffer struct {
	Id      int    `json:"Id"`
	Anfrage string `json:"Anfrage"`
	Treffer int    `json:"Treffer"`
}

type Settings struct {
	DBString           string `json:"DBString"`
	ServerAddress      string `json:"ServerAddress"`
	ResultLimit        int    `json:"ResultLimit"`
	TableName          string `json:"TableName"`
	IdColName          string `json:"IdColName"`
	QueryColName       string `json:"QueryColName"`
	ResultCountColName string `json:"ResultCountColName"`
}
type QueryResponse struct {
	Time        string                `json:"Time"`
	ResultCount int                   `json:"ResultCount"`
	Results     []AutocompleteTreffer `json:"Results"`
}

const settingsFile = "settings.json"

var resultLimit int = 50

func loadSettings() *Settings {
	f, err := os.Open(settingsFile)
	if err != nil {
		// If settings cannot be opened program should not run
		panic(err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	settings := Settings{}
	err = dec.Decode(&settings)
	if err != nil {
		// If settings cannot be opened program should not run
		log.Fatal(err)
	}
	return &settings
}

// Strictly readonly after initial build
var entries []AutocompleteTreffer

// For testing

func printResults(results []int) {
	for _, res := range results {
		fmt.Println(entries[res].Id, " ", entries[res].Anfrage, " ", entries[res].Treffer)
	}

}
func searchQuery(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	query := r.URL.Query()
	q := query.Get("q")
	// empty query
	if len(q) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	results := searchIndex(q)
	response := make([]AutocompleteTreffer, len(results))
	for i := range results {
		response[i] = entries[results[i]]
	}
	// sort by results in reverse
	sort.Slice(response, func(i, j int) bool {
		return response[i].Treffer > response[j].Treffer
	})
	enc := json.NewEncoder(w)
	duration := time.Since(start)
	response = response[:resultLimit]
	resp := QueryResponse{
		Time:        duration.String(),
		ResultCount: len(response),
		Results:     response,
	}
	err := enc.Encode(resp)
	if err != nil {
		log.Println(err)
	}

}

// Standard http server that can be shutdown via signals by e.g. systemctl
func startHttp(settings *Settings) {
	mux := http.NewServeMux()
	mux.HandleFunc("/query/", searchQuery)
	server := &http.Server{
		Addr:         settings.ServerAddress,
		Handler:      mux,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Server started")
	<-stop
	log.Println("Shutting down")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("Shutdown complete")

}

func main() {
	start := time.Now()
	settings := loadSettings()
	close := connect(settings)
	entries = loadAutoCompleteData(settings)
	// DB is no longer needed at this stage
	err := close()
	if err != nil {
		log.Println("Could not close DB ", err)
	}
	log.Println("Loaded DB in ", time.Since(start))
	buildIndex(entries)
	log.Println("Build index in ", time.Since(start))
	resultLimit = settings.ResultLimit
	startHttp(settings)
}
