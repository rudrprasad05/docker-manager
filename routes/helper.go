package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

func (routes *Routes) createAndRunContainer(imageName string, containerName string, cmd []string, hostPort string, containerPort string) (string, error){
	ctx := routes.CTX
	cli := routes.Client

	containerInfo, err := routes.findContainerByName(containerName)
	if err == nil {
		if err := routes.Client.ContainerStart(ctx, containerInfo.ID, container.StartOptions{}); err != nil {
			return "", fmt.Errorf("failed to start container: %w", err)
		}
		fmt.Printf("No new container was created. Container %s is running in detached mode.\n", containerName)
		return containerInfo.ID, nil // Return the container ID
	}

	config := &container.Config{
		Image: imageName, // Use the local image
		Cmd:   cmd,       // Command to execute
		Tty:   false,     // Do not allocate a TTY
	}
	hostConfig := &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{
			nat.Port(containerPort + "/tcp"): {{HostPort: hostPort}},
		},
	}

	// Configure the container
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	fmt.Printf("Container %s is running in detached mode.\n", containerName)
	return resp.ID, nil // Return the container ID
}

func (routes *Routes) IsPortInUse(port string) (bool, error) {
	containers, err := routes.Client.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return false, err
	}

	for _, container := range containers {
		for _, p := range container.Ports {
			if fmt.Sprintf("%d", p.PublicPort) == port {
				return true, nil
			}
		}
	}
	return false, nil
}

func (routes *Routes) findContainerByName(containerName string) (*types.ContainerJSON, error) {
	ctx := context.Background()
	cli := routes.Client


	// Inspect the container using its name
	containerInfo, err := cli.ContainerInspect(ctx, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to find container '%s': %w", containerName, err)
	}

	return &containerInfo, nil
}

func sendJSONResponse(w http.ResponseWriter, status int, data string) {
	encodeData := Message{Data: data}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Encode data to JSON and send response
	if err := json.NewEncoder(w).Encode(encodeData); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (routes *Routes) stopContainer(containerID string, timeoutSeconds int) error {
	ctx := routes.CTX
	var c = timeoutSeconds * 60

	stopOptions := container.StopOptions{
		Timeout: &c, // Use pointer to time.Duration
	}
	if err := routes.Client.ContainerStop(ctx, containerID, stopOptions); err != nil {
		return fmt.Errorf("failed to stop container %s: %w", containerID, err)
	}

	fmt.Printf("Container %s stopped successfully.\n", containerID)
	return nil
}


func (routes *Routes) runContainer(id string) (string, error){
	ctx := context.Background()

	if err := routes.Client.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	fmt.Printf("Container %s is running in detached mode.\n", id)
	return id, nil // Return the container ID
}

func StartDockerOnMac() error {
	cmd := exec.Command("open", "-a", "Docker")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Docker on macOS: %w", err)
	}
	fmt.Println("Docker is starting on macOS...")
	return nil
}

func StartDockerOnLinux() error {
	cmd := exec.Command("sudo", "systemctl", "start", "docker")
	if err := cmd.Run(); err != nil {
		// Fallback to older `service` command if systemctl isn't available
		cmd = exec.Command("sudo", "service", "docker", "start")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start Docker on Linux: %w", err)
		}
	}
	fmt.Println("Docker daemon is starting on Linux...")
	return nil
}