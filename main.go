package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {

	router := chi.NewRouter()
	apiConfig := NewAPIConfig()

	// Fileserver handler
	fileServerHandler := apiConfig.middlewareIncrementMetrics(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	router.Handle("/app", fileServerHandler)
	router.Handle("/app/*", fileServerHandler)

	// API handlers
	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", healthCheck)
	apiRouter.Get("/metrics", apiConfig.apiMetrics)
	apiRouter.HandleFunc("/reset", apiConfig.resetMetrics)
	apiRouter.Post("/chirps", apiConfig.createChirp)
	apiRouter.Get("/chirps", apiConfig.getChirps)
	apiRouter.Get("/chirps/{chirpId}", apiConfig.getChirp)

	router.Mount("/api", apiRouter)

	// Admin handlers
	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiConfig.adminApiMetrics)

	router.Mount("/admin", adminRouter)

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
