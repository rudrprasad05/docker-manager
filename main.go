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
)

func main() {
	logger, err := logs.NewLogger()
	if err != nil{
		log.Println("err", err)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/docker/images/list", handle)
	mux.HandleFunc("/docker/container/run", func(w http.ResponseWriter, r *http.Request) {
		imageName := "go-http-server"
		containerName := "go-http-server-container"
		cmd := []string{"./main"}
		hostPort := "8082"
		containerPort := "8082"
	
		// Create and run the container
		createAndRunContainer(imageName, containerName, cmd, hostPort, containerPort)
	
	})
	// mux.HandleFunc("/download", routes.DownloadImageHandler)

	corsHandler := enableCORS(mux)
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