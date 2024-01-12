package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// Fileserver handler
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	// Health check handler
	mux.HandleFunc("/healthz", healthCheck)

	corsMux := middlewareCors(mux)
	server := http.Server{Handler: corsMux, Addr: "localhost:8080"}
	server.ListenAndServe()
}

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
