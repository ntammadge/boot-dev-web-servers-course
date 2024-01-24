package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/trolfu/boot-dev-web-servers-course/apiConfig"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load environment variables")
	}

	debug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	databasePath := "./database.json"

	if _, err := os.Stat(databasePath); err == nil && *debug {
		os.Remove(databasePath)
	}

	router := chi.NewRouter()
	apiConfig := apiConfig.NewAPIConfig(databasePath, os.Getenv("JWT_SECRET"), os.Getenv("POLKA_API_KEY"))

	// Fileserver handler
	fileServerHandler := apiConfig.MiddlewareIncrementMetrics(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	router.Handle("/app", fileServerHandler)
	router.Handle("/app/*", fileServerHandler)

	// API handlers
	apiRouter := chi.NewRouter()
	apiRouter.Get("/healthz", healthCheck)
	apiRouter.Get("/metrics", apiConfig.ApiMetrics)
	apiRouter.HandleFunc("/reset", apiConfig.ResetMetrics)
	apiRouter.Post("/chirps", apiConfig.CreateChirp)
	apiRouter.Get("/chirps", apiConfig.GetChirps)
	apiRouter.Get("/chirps/{chirpId}", apiConfig.GetChirp)
	apiRouter.Delete("/chirps/{chirpId}", apiConfig.DeleteChirp)
	apiRouter.Post("/users", apiConfig.CreateUser)
	apiRouter.Put("/users", apiConfig.UpdateUser)
	apiRouter.Post("/login", apiConfig.Login)
	apiRouter.Post("/refresh", apiConfig.RefreshAuth)
	apiRouter.Post("/revoke", apiConfig.RevokeAuth)
	apiRouter.Post("/polka/webhooks", apiConfig.UpgradeUser)

	router.Mount("/api", apiRouter)

	// Admin handlers
	adminRouter := chi.NewRouter()
	adminRouter.Get("/metrics", apiConfig.AdminApiMetrics)

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
