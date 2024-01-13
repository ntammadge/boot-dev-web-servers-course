package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	router := chi.NewRouter()
	apiConfig := apiConfig{fileserverHits: 0}

	// Fileserver handler
	fileServerHandler := apiConfig.middlewareIncrementMetrics(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	router.Handle("/app", fileServerHandler)
	router.Handle("/app/*", fileServerHandler)

	// Meta handlers
	router.Get("/healthz", healthCheck)
	router.Get("/metrics", apiConfig.apiMetrics)
	router.HandleFunc("/reset", apiConfig.resetMetrics)

	corsMux := middlewareCors(router)
	server := http.Server{Handler: corsMux, Addr: "localhost:8080"}
	server.ListenAndServe()
}

// Copied (as directed) from ch1.4
func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func healthCheck(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if request.Method == "GET" {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("OK"))
		return
	}
}
