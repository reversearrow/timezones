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

const (
	basePath = "/api"
	timePath = basePath + "/time"
	port     = "8080"
)

type timeZones struct {
	Data   map[string]string `json:"timezones"`
	logger *log.Logger
}

func New(logger *log.Logger) *timeZones {
	return &timeZones{
		Data:   make(map[string]string),
		logger: logger,
	}
}

func (tz *timeZones) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tz.logger.Printf("method:%v, path:%v", r.Method, r.URL.Path)
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
		tz.Data["UTC"] = now.Format(time.RFC822)
	default:
		for k := range parsedQueryValues {
			loc, err := time.LoadLocation(k)
			if err != nil {
				msg := fmt.Sprintf("timezone: %q is invalid", k)
				tz.logger.Printf("msg: %v, err: %v", msg, err)
				http.Error(w, msg, http.StatusNotFound)
				return
			}
			tz.Data[k] = now.In(loc).String()
		}
	}

	if err := encoder.Encode(tz.Data); err != nil {
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
	logger := log.New(os.Stdout, "timezone-service ", log.LstdFlags)
	mux := http.NewServeMux()
	tz := New(logger)
	mux.Handle(timePath, tz)

	svr := http.Server{
		Addr:         fmt.Sprintf(":%s", port),
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

	logger.Printf("listening on port: %v", port)
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
