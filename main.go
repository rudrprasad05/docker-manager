package main

import (
	"context"
	"docker_manager/routes"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rudrprasad05/go-logs/logs"
)


type ContainerRun struct {
	ImageName string 		`json:"imageName"`
	ContainerName string 	`json:"containerName"`
	CMD []string 			`json:"cmd"`
	HostPort string 		`json:"hostPort"`
	ContainerPort string 	`json:"containerPort"`
}

func main() {
	router := mux.NewRouter()
	ctx := context.Background()
	logger, err := logs.NewLogger()
	if err != nil{
		log.Println("err", err)
		return
	}
	routes := &routes.Routes{LOG: logger, CTX: ctx}

	router.HandleFunc("/docker/status/check", routes.GetDockerStatus)
	router.HandleFunc("/docker/status/start", routes.GetStartDocker)
	router.HandleFunc("/docker/images/list", routes.GetImageList)
	router.HandleFunc("/docker/container/run", routes.PostRunCont).Methods("POST")
	router.HandleFunc("/docker/container/stop", routes.PostStopCont).Methods("POST")
	// mux.HandleFunc("/download", routes.DownloadImageHandler)

	corsHandler := enableCORS(router)
	loggedHandler := logs.LoggingMiddleware(logger, corsHandler)

	log.Println("Server running on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", loggedHandler))
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


