package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

func (routes *Routes) GetImageList(w http.ResponseWriter, r *http.Request) {
	var imgArr []string
	
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	images, err := cli.ImageList(routes.CTX, image.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		if len(image.RepoTags) > 0 {
			for _, tag := range image.RepoTags {
				imgArr = append(imgArr, tag)
			}
		} else {
			fmt.Println("<untagged image>")
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(imgArr)
}
