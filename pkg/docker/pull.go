package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func PullImage(repository string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	authConfig := types.AuthConfig{
		Username: "admin",
		Password: "P8wk%pU9D!#bSp",
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return err
	}

	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	reader, err := cli.ImagePull(ctx, repository, types.ImagePullOptions{RegistryAuth: authStr})
	if err != nil {
		return err
	}
	defer reader.Close()

	io.Copy(os.Stdout, reader)

	return nil
}
