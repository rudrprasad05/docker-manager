package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type CmdPort struct {
	Cmd  []string `json:"cmd"`
	Port []nat.Port   `json:"port"`
}

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

	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(images)
}

func (routes *Routes) GetContainerList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Create Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer cli.Close()

	// Get the list of running containers
	containers, err := cli.ContainerList(routes.CTX, container.ListOptions{})
	if err != nil {
		fmt.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Is Docker running?"})
		return
	}

	// Return container data as JSON
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(containers)
}


func (routes *Routes) GetDockerStatus(w http.ResponseWriter, r *http.Request) {
	// Check if Docker is runnin
	if routes.isDockerRunning() {
		sendJSONResponse(w, http.StatusBadRequest, "docker is running")
	} else {
		sendJSONResponse(w, http.StatusServiceUnavailable, "docker not running")
	}
}

func (routes *Routes) GetCMDStatus(w http.ResponseWriter, r *http.Request) {

	// Parse image name from query or body
	imageName := r.URL.Query().Get("image") // Example: "nginx:latest"
	if imageName == "" {
		sendJSONResponse(w, http.StatusBadRequest,  "Image name is required")
		return
	}

	// Inspect the image
	imageInfo, _, err := routes.Client.ImageInspectWithRaw(routes.CTX, imageName)
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, "Failed to inspect image")
	}

	// Determine command
	cmd := imageInfo.Config.Cmd
	entrypoint := imageInfo.Config.Entrypoint

	if len(cmd) == 0 && len(entrypoint) == 0 {
		sendJSONResponse(w, http.StatusMethodNotAllowed, "Image has no CMD or ENTRYPOINT defined")
		return
	}

	portProtocol := imageInfo.Config.ExposedPorts

	var ports []nat.Port
	for port := range portProtocol {
		ports = append(ports, port[:len(port)-len("/tcp")]) // Remove "/tcp" suffix
	}

	data := CmdPort{
		Cmd:  cmd,
		Port: ports,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (routes *Routes) GetStartDocker(w http.ResponseWriter, r *http.Request) {
	// Check if Docker is runnin

	if routes.isDockerRunning() {
		// Respond with the container and image details, if Docker is running

		sendJSONResponse(w, http.StatusBadRequest, "docker is running")
	} else {
		// Respond with a message if Docker is down
		err := startDockerDesktop()
		if err != nil{

			sendJSONResponse(w, http.StatusServiceUnavailable, "docker is running")
		}
		sendJSONResponse(w, http.StatusOK, "docker is running")
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