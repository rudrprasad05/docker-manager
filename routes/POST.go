package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/rudrprasad05/go-logs/logs"
)

type Routes struct {
	LOG *logs.Logger
	CTX context.Context
}

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


func (routes *Routes) PostStopCont(w http.ResponseWriter, r *http.Request){
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



func (routes *Routes) PostRunCont(w http.ResponseWriter, r *http.Request){
	var imageProps ContainerRun

	err := json.NewDecoder(r.Body).Decode(&imageProps)
	if err != nil {
		data := Message{Data: "invalid json"}
		fmt.Println("400 bad request; invalid json", err)
		sendJSONResponse(w, http.StatusBadRequest, data)
		return
	}

	// Create and run the container
	resp, respErr := routes.createAndRunContainer(imageProps.ImageName, imageProps.ContainerName, imageProps.CMD, imageProps.HostPort, imageProps.ContainerPort)
	if respErr != nil {
		data := Message{Data: "invalid json"}
		fmt.Println("400 bad request; invalid json", err, resp)
		sendJSONResponse(w, http.StatusBadRequest, data)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
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


func runContainer(id string) (string, error){
	ctx := context.Background()

	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	// Configure the container

	if err := cli.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	fmt.Printf("Container %s is running in detached mode.\n", id)
	return id, nil // Return the container ID
}

func (routes *Routes) createAndRunContainer(imageName string, containerName string, cmd []string, hostPort string, containerPort string) (string, error){
	ctx := routes.CTX

	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	containerInfo, err := findContainerByName(containerName)
	if err == nil {
		if err := cli.ContainerStart(ctx, containerInfo.ID, container.StartOptions{}); err != nil {
			return "", fmt.Errorf("failed to start container: %w", err)
		}
		fmt.Printf("No new container was created. Container %s is running in detached mode.\n", containerName)
		return containerInfo.ID, nil // Return the container ID
	}

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
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	fmt.Printf("Container %s is running in detached mode.\n", containerName)
	return resp.ID, nil // Return the container ID
}

func findContainerByName(containerName string) (*types.ContainerJSON, error) {
	ctx := context.Background()

	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer cli.Close()

	// Inspect the container using its name
	containerInfo, err := cli.ContainerInspect(ctx, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to find container '%s': %w", containerName, err)
	}

	return &containerInfo, nil
}

func sendJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Encode data to JSON and send response
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

