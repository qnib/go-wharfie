package wharfie

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"

)


const (
	// Debug mode.
	Debug = false
	DOCKER_API_VERSION = "v1.30"
	DOCKER_API_HOST = "unix:///var/run/docker.sock"
	DOCKER_CERT_PATH="/etc/docker/"
)

type Wharfie struct {
	do 			Options
	engCli 		*client.Client
}

func New(opts ...Option) Wharfie {
	options := defaultDracerOptions
	for _, o := range opts {
		o(&options)
	}
	return Wharfie{
		do: options,
	}
}

func (w *Wharfie) Connect() {
	var err error
	w.engCli, err = client.NewEnvClient()
	if err != nil {
		fmt.Printf("Could not connect docker/docker/client to '%s': %v", w.do.DockerSocket, err)
		return
	}
	info, err := w.engCli.Info(context.Background())
	if err != nil {
		fmt.Printf("Error during Info(): %v >err> %s", info, err)
		return
	} else {
		fmt.Printf("Connected to '%s' / v'%s' (SWARM: %s)\n", info.Name, info.ServerVersion, info.Swarm.LocalNodeState)
	}

}

func (w *Wharfie) Run() {
	w.Connect()
}