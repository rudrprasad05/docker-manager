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

type TPostStopCont struct {
	ID string `json:"id"`
}


func (routes *Routes) PostStopCont(w http.ResponseWriter, r *http.Request){
	var reqBody TPostStopCont
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		routes.LOG.Error("here1")

		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create and run the container
	containerID := reqBody.ID
	fmt.Println(containerID)
	respErr := routes.stopContainer(containerID, 10)
	if respErr != nil {
		routes.LOG.Error("here2")
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

	// Create and run the container
	resp, respErr := routes.createAndRunContainer(imageProps.ImageName, imageProps.ContainerName, imageProps.CMD, imageProps.HostPort, imageProps.ContainerPort)
	if respErr != nil {
		fmt.Println("400 bad request; invalid json", err, resp)
		
		sendJSONResponse(w, http.StatusBadRequest, "invalid json")
		return
	}
	
	json.NewEncoder(w).Encode("resp")
	sendJSONResponse(w, http.StatusOK, "a")
}

func (routes *Routes) PostRunCont(w http.ResponseWriter, r *http.Request){
	var imageProps ContainerRun
	var cli = routes.Client
	var ctx = routes.CTX

	err := json.NewDecoder(r.Body).Decode(&imageProps)
	if err != nil {
		fmt.Println("400 bad request; invalid json", err)

		sendJSONResponse(w, http.StatusBadRequest, "invalid json")
		return
	}

	oldCont, err := cli.ContainerInspect(ctx, imageProps.ImageName)
	if err != nil {
		resp, respErr := routes.createAndRunContainer(imageProps.ImageName, imageProps.ContainerName, imageProps.CMD, imageProps.HostPort, imageProps.ContainerPort)
		if respErr != nil {
			fmt.Println("400 bad request; invalid json", err, resp)
			
			sendJSONResponse(w, http.StatusBadRequest, "invalid json")
			return
		}
		
		json.NewEncoder(w).Encode("resp")
		sendJSONResponse(w, http.StatusOK, "a")
		return
	}

	if err := cli.ContainerStart(ctx, oldCont.ID, container.StartOptions{}); err != nil {
		sendJSONResponse(w, http.StatusBadRequest, "invalid json")
		return
	}

	sendJSONResponse(w, http.StatusBadRequest, oldCont.ID)
	return

	// Create and run the container
	
}





