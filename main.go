package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/rudrprasad05/go-logs/logs"
)

func main() {
	logger, err := logs.NewLogger()
	if err != nil{
		log.Println("err", err)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/docker/images/list", handle)
	// mux.HandleFunc("/convert", routes.ConvertFile)
	// mux.HandleFunc("/download", routes.DownloadImageHandler)

	corsHandler := enableCORS(mux)
	loggedHandler := logs.LoggingMiddleware(logger, corsHandler)


	log.Println("Server running on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", loggedHandler))
}

func handle(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	containers, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Println(container.ID)
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "https://imageconverter.rudrprasad.com")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}