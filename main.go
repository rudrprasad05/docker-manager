package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/rudrprasad05/go-logs/logs"

	"github.com/docker/docker/api/types/container"
	"github.com/gorilla/mux"
)

type Message struct {
	Data string `json:"data"`
}

type ContainerRun struct {
	ImageName string 		`json:"imageName"`
	ContainerName string 	`json:"containerName"`
	CMD []string 			`json:"cmd"`
	HostPort string 		`json:"hostPort"`
	ContainerPort string 	`json:"containerPort"`
}

type Routes struct {
	LOG *logs.Logger
}

func main() {
	router := mux.NewRouter()
	logger, err := logs.NewLogger()
	if err != nil{
		log.Println("err", err)
		return
	}


	router.HandleFunc("/docker/images/list", handle)
	router.HandleFunc("/docker/container/run", PostRunCont).Methods("POST")
	router.HandleFunc("/docker/container/stop", PostStopCont).Methods("POST")
	// mux.HandleFunc("/download", routes.DownloadImageHandler)

	corsHandler := enableCORS(router)
	loggedHandler := logs.LoggingMiddleware(logger, corsHandler)

	

	log.Println("Server running on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", loggedHandler))
}

func handle(w http.ResponseWriter, r *http.Request) {
	var imgArr []string
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		if len(image.RepoTags) > 0 {
			for _, tag := range image.RepoTags {
				imgArr = append(imgArr, tag)
			}
		} else {
			fmt.Println("<untagged image>")
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(imgArr)
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

func PostRunCont(w http.ResponseWriter, r *http.Request){
	var imageProps ContainerRun

	err := json.NewDecoder(r.Body).Decode(&imageProps)
	if err != nil {
		data := Message{Data: "invalid json"}
		fmt.Println("400 bad request; invalid json", err)
		sendJSONResponse(w, http.StatusBadRequest, data)
		return
	}

	// Create and run the container
	resp, respErr := createAndRunContainer(imageProps.ImageName, imageProps.ContainerName, imageProps.CMD, imageProps.HostPort, imageProps.ContainerPort)
	if respErr != nil {
		data := Message{Data: "invalid json"}
		fmt.Println("400 bad request; invalid json", err)
		sendJSONResponse(w, http.StatusBadRequest, data)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func PostStopCont(w http.ResponseWriter, r *http.Request){
	var data map[string]string

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		data := Message{Data: "invalid json"}
		fmt.Println("400 bad request; invalid json", err)
		sendJSONResponse(w, http.StatusBadRequest, data)
		return
	}

	// Create and run the container
	containerID := data["id"]
	respErr := stopContainer(containerID, 10)
	if respErr != nil {
		data := Message{Data: "invalid json"}
		fmt.Println("400 bad request; invalid json", err)
		sendJSONResponse(w, http.StatusBadRequest, data)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	msg := Message{Data: "container stopped"}
	json.NewEncoder(w).Encode(msg)
}

func createAndRunContainer(imageName string, containerName string, cmd []string, hostPort string, containerPort string) (string, error){
	ctx := context.Background()

	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	// Configure the container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName, // Use the local image
		Cmd:   cmd,       // Command to execute
		Tty:   false,     // Do not allocate a TTY
	}, &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{
			nat.Port(containerPort + "/tcp"): {{HostPort: hostPort}},
		},
	}, nil, nil, containerName)
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	fmt.Printf("Container %s is running in detached mode.\n", containerName)
	return resp.ID, nil // Return the container ID
}

func stopContainer(containerID string, timeoutSeconds int) error {
	ctx := context.Background()

	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer cli.Close()

	
	// Stop the container
	var c = timeoutSeconds * 60

	stopOptions := container.StopOptions{
		Timeout: &c, // Use pointer to time.Duration
	}
	if err := cli.ContainerStop(ctx, containerID, stopOptions); err != nil {
		return fmt.Errorf("failed to stop container %s: %w", containerID, err)
	}

	fmt.Printf("Container %s stopped successfully.\n", containerID)
	return nil
}

func sendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Encode data to JSON and send response
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}