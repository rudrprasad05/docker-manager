package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

func (routes *Routes) PostCreateAndRunCont(w http.ResponseWriter, r *http.Request){
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





