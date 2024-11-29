package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"sort"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type CmdPort struct {
	Cmd  []string `json:"cmd"`
	Port []nat.Port   `json:"port"`
}

func (routes *Routes) IsDockerUp(){
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		routes.LOG.Error("")
		return 
	}
	defer cli.Close()

	_, err = routes.Client.Ping(routes.CTX)
	if err != nil {
		var errN error
		switch runtime.GOOS {
			case "darwin": // macOS
				errN = StartDockerOnMac()
			case "linux": // Linux
				errN = StartDockerOnLinux()
			default:
				errN = fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
		}

		if errN != nil {
			fmt.Printf("Failed to start Docker: %v\n", errN)
		} else {
			fmt.Println("Docker started successfully. Please wait a few seconds before retrying.")
		}
	}
} 

func (routes *Routes) GetImageList(w http.ResponseWriter, r *http.Request) {
	// var imgArr []string
	w.Header().Set("Content-Type", "application/json")
	
	images, err := routes.Client.ImageList(routes.CTX, image.ListOptions{})
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

func (routes *Routes) GetImageById(w http.ResponseWriter, r *http.Request) {
	// var imgArr []string
	w.Header().Set("Content-Type", "application/json")

	images, err := routes.Client.ImageList(routes.CTX, image.ListOptions{})
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
	// Get the list of running containers
	containers, err := routes.Client.ContainerList(routes.CTX, container.ListOptions{
		All: true,
	})
	if err != nil {
		fmt.Print(err)
		sendJSONResponse(w, http.StatusInternalServerError, "is it running")
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


func (routes *Routes) SearchImages(w http.ResponseWriter, r *http.Request) {
	// Set the response content type
	w.Header().Set("Content-Type", "application/json")

	// Get the search query from the URL
	query := r.URL.Query().Get("search")

	var results []registry.SearchResult
	var err error;
	if query == "" {
		// Default behavior: Fetch top images by popularity
		 
		results, err = routes.Client.ImageSearch(routes.CTX, "popular", registry.SearchOptions{Limit: 10})
	} else {
		// Search for images matching the query
		results, err = routes.Client.ImageSearch(routes.CTX, query, registry.SearchOptions{Limit: 20})
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to search images: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].StarCount > results[j].StarCount
	})

	// Transform the Docker search results into a response structure
	var response []registry.SearchResult
	for _, result := range results {
		response = append(response, registry.SearchResult{
			Name:        result.Name,
			Description: result.Description,
			StarCount:   result.StarCount,
			IsOfficial:  result.IsOfficial,
		})
	}

	// Encode the response into JSON and send it
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}