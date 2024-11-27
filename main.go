package main

import (
	"context"
	"docker_manager/routes"
	"log"
	"net/http"

	"github.com/docker/docker/client"

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

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to initialize Docker client: %v", err)
	}

	routes := &routes.Routes{LOG: logger, CTX: ctx, Client: cli}

	router.HandleFunc("/docker/status/check", routes.GetDockerStatus)
	router.HandleFunc("/docker/status/start", routes.GetStartDocker)

	router.HandleFunc("/docker/images/list", routes.GetImageList)
	router.HandleFunc("/docker/images/cmd/status", routes.GetCMDStatus)
	router.HandleFunc("/docker/images/delete", routes.DeleteImage)


	router.HandleFunc("/docker/container/list", routes.GetContainerList).Methods("GET")
	router.HandleFunc("/docker/container/run", routes.PostCreateAndRunCont).Methods("POST")
	router.HandleFunc("/docker/container/run-existing", routes.PostRunCont).Methods("POST")
	router.HandleFunc("/docker/container/stop", routes.PostStopCont).Methods("POST")
	// mux.HandleFunc("/download", routes.DownloadImageHandler)

	corsHandler := EnableCORS(router)
	loggedHandler := logs.LoggingMiddleware(logger, corsHandler)

	log.Println("Server running on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", loggedHandler))
}



