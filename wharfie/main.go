package wharfie

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"

	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"log"

)


const (
	// Debug mode.
	Debug = false
	DOCKER_API_VERSION = "v1.30"
	DOCKER_API_HOST = "unix:///var/run/docker.sock"
	DOCKER_CERT_PATH="/etc/docker/"
)

var (
	ctx = context.Background()
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
		log.Printf("Could not connect docker/docker/client to '%s': %v", w.do.DockerSocket, err)
		return
	}
	info, err := w.engCli.Info(ctx)
	if err != nil {
		log.Printf("Error during Info(): %v >err> %s", info, err)
		return
	} else {
		log.Printf("Connected to '%s' / v'%s' (SWARM: %s)\n", info.Name, info.ServerVersion, info.Swarm.LocalNodeState)
	}
}

func (w *Wharfie) GetNodesFiltered() ([]swarm.Node, error) {
	f := filters.NewArgs()
	for _, n := range w.do.NodeList {
		f.Add("name", n)
	}
	return w.engCli.NodeList(context.Background(), types.NodeListOptions{Filters: f})
}

func (w *Wharfie) AddJobIdLabel() (err error) {
	nodeList, err := w.GetNodesFiltered()
	if err != nil {
		log.Fatalf("Error while NodeList(): %s\n", err)
		return
	}
	for _, node := range nodeList {
		key := fmt.Sprintf("job.id.%s", w.do.JobId)
		n, _ ,err := w.engCli.NodeInspectWithRaw(ctx, node.ID)
		if err != nil {
			log.Printf("Error while NodeInspectWithRaw(%s): %s\n", node.ID, err)
			return err
		}
		_, ok := n.Spec.Annotations.Labels[key]; if ok {
			log.Printf("Node '%s' already has label '%s'\n", n.Description.Hostname, key)
			continue
		}
		n.Spec.Annotations.Labels[key] = "true"
		log.Printf("Add label '%s=true' to %s\n", key, n.Description.Hostname)
		err = w.engCli.NodeUpdate(ctx, node.ID, n.Version, n.Spec)
		if err != nil {
			log.Fatalf("Error while NodeUpdate(): %s\n", err)
			return err
		}
	}
	return
}

func (w *Wharfie) RmJobIdLabel() (err error) {
	if err != nil {
		return
	}
	nodelist, err := w.GetNodesFiltered()
	for _, node := range nodelist {
		n, _ ,err := w.engCli.NodeInspectWithRaw(ctx, node.ID)
		if err != nil {
			log.Fatalf("Error while NodeInspectWithRaw(%s): %s\n", node.ID, err)
			continue
		}
		key := fmt.Sprintf("slurm.jobid.%s", w.do.JobId)
		_, ok := n.Spec.Annotations.Labels[key]; if ok {
			log.Printf("Remove label '%s=true' to %s\n", key, n.Description.Hostname)
			delete(n.Spec.Labels, key)
			err = w.engCli.NodeUpdate(ctx, node.ID, n.Version, n.Spec)
		}
	}
	return
}

func (w *Wharfie) Stage() {
	w.Connect()
	w.AddJobIdLabel()
	w.CreateService()
	w.WaitForService()


}
func (w *Wharfie) Remove() {
	w.Connect()
	w.RemoveService()
	w.RmJobIdLabel()
}