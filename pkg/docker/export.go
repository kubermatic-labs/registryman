package docker

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func ExportImages(repo, destinationPath string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	images, err := searchImage(repo, &ctx, cli)
	if err != nil {
		return err
	}

	imageIDs := []string{}
	for _, image := range images {
		fmt.Println(image.ID)
		imageIDs = append(imageIDs, image.ID)
	}

	reader, err := cli.ImageSave(ctx, imageIDs)
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	destination, err := os.Create(destinationPath)
	if err != nil {
		panic(err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, reader)
	if err != nil {
		panic(err)
	}

	return nil
}

func searchImage(repo string, ctx *context.Context, cli *client.Client) ([]types.ImageSummary, error) {
	args := filters.NewArgs(filters.KeyValuePair{
		Key:   "reference",
		Value: fmt.Sprintf("%s:*", repo),
	})
	images, err := cli.ImageList(*ctx, types.ImageListOptions{
		Filters: args,
	})
	if err != nil {
		return nil, err
	}

	return images, nil
}
