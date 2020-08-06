package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io"
	"os"
)

type Manager struct {
	*client.Client
}

func NewManager() (*Manager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	reader, err := cli.ImagePull(context.Background(), imageName, types.ImagePullOptions{})
	if err != nil {
		return nil, err
	}

	//containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	//container := containers[0]
	//container.NetworkSettings.Networks

	// TODO: maybe remove
	io.Copy(os.Stdout, reader)

	return &Manager{cli}, nil
}

