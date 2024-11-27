package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/rudrprasad05/go-logs/logs"
)

type Routes struct {
	LOG *logs.Logger
	CTX context.Context
	Client *client.Client
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
	respErr := routes.stopContainer(containerID, 10)
	if respErr != nil {
		data := Message{Data: "invalid json"}
		fmt.Println("400 bad request; invalid json", err)
		sendJSONResponse(w, http.StatusBadRequest, data)
		return
	}

	msg := Message{Data: "container stopped"}
	sendJSONResponse(w, http.StatusOK, msg)

}



func (routes *Routes) PostRunCont(w http.ResponseWriter, r *http.Request){
	var imageProps ContainerRun

	err := json.NewDecoder(r.Body).Decode(&imageProps)
	if err != nil {
		fmt.Println("400 bad request; invalid json", err)

		data := Message{Data: "invalid json"}
		sendJSONResponse(w, http.StatusBadRequest, data)
		return
	}

	// Create and run the container
	resp, respErr := routes.createAndRunContainer(imageProps.ImageName, imageProps.ContainerName, imageProps.CMD, imageProps.HostPort, imageProps.ContainerPort)
	if respErr != nil {
		fmt.Println("400 bad request; invalid json", err, resp)
		
		data := Message{Data: "invalid json"}
		sendJSONResponse(w, http.StatusBadRequest, data)
		return
	}
	
	json.NewEncoder(w).Encode("resp")
	sendJSONResponse(w, http.StatusOK, "a")
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



