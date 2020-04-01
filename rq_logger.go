package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
)

type RequestLogger struct {
	w *log.Logger
}

// TODO: rand.Seed(...) to ensure request IDs are unique

const RequestIDHeader = "X-Homework-Request-ID"

func NewRequestLogger(filename string) (*RequestLogger, error) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("creating request logger: %w", err)
	}

	w := log.New(f, "", log.Ldate|log.Ltime)
	return &RequestLogger{w}, nil
}

func (l *RequestLogger) WrapHTTPHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rqID := rand.Uint64()
		r.Header.Set(RequestIDHeader, fmt.Sprint(rqID))

		// TODO: also log r.Method ?
		l.w.Printf("%v %s", rqID, r.URL)

		h.ServeHTTP(w, r)
	})
}
