package routes

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"

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

type TPostStopCont struct {
	ID string `json:"id"`
}


func (routes *Routes) PostStopCont(w http.ResponseWriter, r *http.Request){
	var reqBody TPostStopCont
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create and run the container
	containerID := reqBody.ID
	respErr := routes.stopContainer(containerID, 10)
	if respErr != nil {
		fmt.Println("400 bad request; invalid json", respErr)
		sendJSONResponse(w, http.StatusBadRequest, "invalid json")
		return
	}

	sendJSONResponse(w, http.StatusOK, "container stopped")

}

func (routes *Routes) PostCreateAndRunCont(w http.ResponseWriter, r *http.Request){
	var imageProps ContainerRun

	err := json.NewDecoder(r.Body).Decode(&imageProps)
	if err != nil {
		fmt.Println("400 bad request; invalid json", err)

		sendJSONResponse(w, http.StatusBadRequest, "invalid json")
		return
	}

	isUsed, _ := routes.IsPortInUse(imageProps.HostPort)
	if isUsed == true{
		sendJSONResponse(w, http.StatusForbidden, "port in use")
		return
	}

	// Create and run the container
	resp, respErr := routes.createAndRunContainer(imageProps.ImageName, imageProps.ContainerName, imageProps.CMD, imageProps.HostPort, imageProps.ContainerPort)
	if respErr != nil {
		fmt.Println("500 internal server", respErr)
		
		sendJSONResponse(w, http.StatusInternalServerError, "err in starting cont")
		return
	}
	
	json.NewEncoder(w).Encode(resp)
	sendJSONResponse(w, http.StatusOK, "a")
}

func (routes *Routes) PostRunCont(w http.ResponseWriter, r *http.Request){
	var props TPostStopCont
	var cli = routes.Client
	var ctx = routes.CTX

	err := json.NewDecoder(r.Body).Decode(&props)
	if err != nil {
		fmt.Println("400 bad request; invalid json", err)

		sendJSONResponse(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := cli.ContainerStart(ctx, props.ID, container.StartOptions{}); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, "invalid json")
		return
	}

	sendJSONResponse(w, http.StatusOK, props.ID)
	return

	// Create and run the container
	
}

func (routes *Routes) PostPullImage(w http.ResponseWriter, r *http.Request) {
	// Parse image name from the request body
	imageName := r.URL.Query().Get("image")
	if imageName == "" {
		sendJSONResponse(w, http.StatusBadRequest,  "Missing image name")
		return
	}

	// Pull the image
	reader, err := routes.Client.ImagePull(routes.CTX, imageName, image.PullOptions{})
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to pull image: %v", err))
		return
	}
	defer reader.Close()
	
	encoder := json.NewEncoder(w)
	scanner := bufio.NewScanner(reader)
	w.WriteHeader(http.StatusOK)

	for scanner.Scan() {
		var progress map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &progress); err != nil {
			continue // Skip invalid progress updates
		}

		// Send the progress as JSON
		if err := encoder.Encode(progress); err != nil {
			fmt.Printf("Failed to encode progress: %v\n", err)
			break
		}

		// Flush response writer to send updates immediately
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Scanner error: %v\n", err)
	}

	// Stream the pull progress
	// var pullOutput strings.Builder
	// io.Copy(&pullOutput, reader)

	// Return success
	// sendJSONResponse(w, http.StatusOK, "pull ok")
}

