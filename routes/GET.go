package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

func (routes *Routes) GetImageList(w http.ResponseWriter, r *http.Request) {
	// var imgArr []string
	w.Header().Set("Content-Type", "application/json")
	
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	defer cli.Close()

	images, err := cli.ImageList(routes.CTX, image.ListOptions{})
	if err != nil {
		fmt.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		data := Message{Data: "is docker running?"}
		json.NewEncoder(w).Encode(data)
		return
	}

	// for _, image := range images {
	// 	if len(image.RepoTags) > 0 {
	// 		for _, tag := range image.RepoTags {
	// 			imgArr = append(imgArr, tag)
	// 		}
	// 	} else {
	// 		fmt.Println("<untagged image>")
	// 	}
	// }
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(images)
}

func (routes *Routes) GetDockerStatus(w http.ResponseWriter, r *http.Request) {
	// Check if Docker is runnin

	if routes.isDockerRunning() {
		// Respond with the container and image details, if Docker is running
		data := Message{Data: "docker is running"}
		sendJSONResponse(w, http.StatusBadRequest, data)
	} else {
		// Respond with a message if Docker is down
		data := Message{Data: "docker not running"}
		sendJSONResponse(w, http.StatusServiceUnavailable, data)
	}
}

func (routes *Routes) GetStartDocker(w http.ResponseWriter, r *http.Request) {
	// Check if Docker is runnin

	if routes.isDockerRunning() {
		// Respond with the container and image details, if Docker is running
		data := Message{Data: "docker is running"}
		sendJSONResponse(w, http.StatusBadRequest, data)
	} else {
		// Respond with a message if Docker is down
		err := startDockerDesktop()
		if err != nil{
			data := Message{Data: "an error occured while starting docker"}
			sendJSONResponse(w, http.StatusServiceUnavailable, data)
		}
		data := Message{Data: "docker started"}
		sendJSONResponse(w, http.StatusOK, data)
	}
}

func startDockerDesktop() error {
	// Run the open command to start Docker Desktop on macOS
	cmd := exec.Command("open", "-a", "Docker")

	// Run the command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to start Docker Desktop: %w", err)
	}

	return nil
}

func (routes *Routes) isDockerRunning() bool {
	ctx := routes.CTX

	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println("Error connecting to Docker:", err)
		return false
	}
	defer cli.Close()

	// Try to get the Docker version
	_, err = cli.ServerVersion(ctx)
	if err != nil {
		// If an error occurs (e.g., Docker is not running), return false
		fmt.Println("Docker is not running:", err)
		return false
	}

	// If no error, Docker is running
	return true
}