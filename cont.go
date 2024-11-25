package main

import (
	"encoding/json"
	"log"
	"net/http"

	container "github.com/docker/docker/api/types/container"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

func cont() {
	r := mux.NewRouter()
	r.HandleFunc("/containers", listContainers).Methods("GET")
	r.HandleFunc("/containers/{id}/start", startContainer).Methods("POST")
	r.HandleFunc("/containers/{id}/stop", stopContainer).Methods("POST")

	log.Println("API running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func listContainers(w http.ResponseWriter, r *http.Request) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	containers, err := cli.ContainerList(context.Background(), containertypes.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containers)
}

func startContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerID := vars["id"]

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := cli.ContainerStart(context.Background(), containerID, container.StartOptions{}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func stopContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerID := vars["id"]

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	noWaitTimeout := 0
	if err := cli.ContainerStop(context.Background(), containerID, containertypes.StopOptions{Timeout: &noWaitTimeout}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
