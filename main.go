package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const basePath = "/api"
const timePath = basePath + "/time"

type timeZones struct {
	data map[string]string `json:"timezones"`
}

func New() *timeZones {
	return &timeZones{
		data: make(map[string]string),
	}
}

func (tz *timeZones) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tz.getTimeZone(w, r)
		return
	}
}

func (tz *timeZones) getTimeZone(w http.ResponseWriter, r *http.Request) {
	qv := r.URL.Query()
	parsedQueryValues := parseTimeZonesFromQuery(qv)
	now := time.Now().UTC()
	encoder := json.NewEncoder(w)
	switch len(parsedQueryValues) {
	case 0:
		tz.data["UTC"] = now.Format(time.RFC822)
	default:
		for k, _ := range parsedQueryValues {
			loc, err := time.LoadLocation(k)
			if err != nil {
				http.Error(w, fmt.Sprint("invalid timezone"), http.StatusNotFound)
				return
			}
			tz.data[k] = now.In(loc).String()
		}
	}

	if err := encoder.Encode(tz.data); err != nil {
		http.Error(w, fmt.Sprintf("internal server error"), http.StatusInternalServerError)
	}
}

func parseTimeZonesFromQuery(queryValues url.Values) map[string]struct{} {
	output := make(map[string]struct{})
	values, ok := queryValues["tz"]
	if !ok {
		return output
	}

	for _, v := range strings.Split(values[0], ",") {
		output[v] = struct{}{}
	}

	return output
}

func main() {
	mux := http.NewServeMux()
	tz := New()
	mux.Handle(timePath, tz)

	logger := log.New(os.Stdout, "timezone-service ", log.LstdFlags)
	svr := http.Server{
		Addr:         ":8081",
		Handler:      mux,
		ReadTimeout:  time.Second * 1,
		WriteTimeout: time.Second * 10,
		IdleTimeout:  time.Second * 5,
	}

	go func() {
		if err := svr.ListenAndServe(); err != nil {
			switch err {
			case http.ErrServerClosed:
				logger.Printf("server shutdown successfully: %v", err)
				os.Exit(0)
			default:
				logger.Printf("error starting go server: %v", err)
				os.Exit(0)
			}
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	sig := <-shutdown
	logger.Printf("shutdown signal received: %v, attempting to shutdown the serer", sig)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	if err := svr.Shutdown(ctx); err != nil {
		logger.Printf("error shutting down the server: %v", err)
	}
}
