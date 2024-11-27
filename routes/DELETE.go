package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
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

	// Initialize Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create Docker client: %v", err), http.StatusInternalServerError)
		return
	}
	defer cli.Close()

	// Remove the Docker image

	removedItems, err := cli.ImageRemove(routes.CTX, imageID, image.RemoveOptions{
		Force:         false,
		PruneChildren: true,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete Docker image: %v", err), http.StatusInternalServerError)
		return
	}

	// Send the response back to the client
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(removedItems); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}