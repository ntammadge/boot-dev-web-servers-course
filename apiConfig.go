package main

import (
	"fmt"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}

func (config *apiConfig) middlewareIncrementMetrics(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		config.fileserverHits++
		handler.ServeHTTP(writer, request)
	})
}

func (config *apiConfig) apiMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if request.Method == http.MethodGet {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(fmt.Sprintf("Hits: %v", config.fileserverHits)))
		return
	}
}

func (config *apiConfig) resetMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	config.fileserverHits = 0
}
