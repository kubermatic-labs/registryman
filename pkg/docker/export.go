package docker

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/client"
)

func TestExport(dest string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	reader, err := cli.ImageSave(ctx, []string{"d4ff818577bc"})

	if err != nil {
		panic(err)
	}

	destination, err := os.Create(dest)
	if err != nil {
		panic(err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, reader)
	if err != nil {
		panic(err)
	}

}
