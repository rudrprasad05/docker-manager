package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
)

func (routes *Routes) DeleteImage(w http.ResponseWriter, r *http.Request) {
	// Only allow DELETE method
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse image ID or name from query parameter or request body
	imageID := r.URL.Query().Get("id") // Check query parameter first
	if imageID == "" {
		var body struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" {
			http.Error(w, "Invalid or missing 'id'", http.StatusBadRequest)
			return
		}
		imageID = body.ID
	}

	// Remove the Docker image

	removedItems, err := routes.Client.ImageRemove(routes.CTX, imageID, image.RemoveOptions{
		Force:         true,
		PruneChildren: true,
	})
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete Docker image: %v", err))
		return
	}

	// Send the response back to the client
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(removedItems); err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, fmt.Sprintf("Failed to encode response"))
		return
	}
}


func (routes *Routes) DeleteContainer(w http.ResponseWriter, r *http.Request) {
	// Only allow DELETE method
	if r.Method != http.MethodDelete {
		sendJSONResponse(w, http.StatusMethodNotAllowed, "failed to delete")
		return
	}

	// Parse image ID or name from query parameter or request body
	imageID := r.URL.Query().Get("id") // Check query parameter first
	if imageID == "" {
		var body struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" {
			sendJSONResponse(w, http.StatusBadRequest, "missing info")
			return
		}
		imageID = body.ID
	}
	// Remove the Docker image

	err := routes.Client.ContainerRemove(routes.CTX, imageID, container.RemoveOptions{
		Force: true,
	})
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, "failed to delete")
		routes.LOG.Error(err.Error())
		return
	}

	// Send the response back to the client
	sendJSONResponse(w, http.StatusOK, "container deleted")
	return
}