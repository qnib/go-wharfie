package wharfie

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"

	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/mount"
	"os"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
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
		fmt.Printf("Error while NodeList(): %s\n", err)
		return
	}
	for _, node := range nodeList {
		key := fmt.Sprintf("slurm.jobid.%s", w.do.JobId)
		n, _ ,err := w.engCli.NodeInspectWithRaw(context.Background(), node.ID)
		if err != nil {
			fmt.Printf("Error while NodeInspectWithRaw(%s): %s\n", node.ID, err)
			return err
		}
		_, ok := n.Spec.Annotations.Labels[key]; if ok {
			fmt.Printf("Node '%s' already has label '%s'\n", n.Description.Hostname, key)
			continue
		}
		n.Spec.Annotations.Labels[key] = "true"
		fmt.Printf("Add label '%s=true' to %s\n", key, n.Description.Hostname)
		err = w.engCli.NodeUpdate(context.Background(), node.ID, n.Version, n.Spec)
		if err != nil {
			fmt.Printf("Error while NodeUpdate(): %s\n", err)
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
		n, _ ,err := w.engCli.NodeInspectWithRaw(context.Background(), node.ID)
		if err != nil {
			fmt.Printf("Error while NodeInspectWithRaw(%s): %s\n", node.ID, err)
			continue
		}
		key := fmt.Sprintf("slurm.jobid.%s", w.do.JobId)
		_, ok := n.Spec.Annotations.Labels[key]; if ok {
			fmt.Printf("Remove label '%s=true' to %s\n", key, n.Description.Hostname)
			delete(n.Spec.Labels, key)
			err = w.engCli.NodeUpdate(context.Background(), node.ID, n.Version, n.Spec)
		}
	}
	return
}


func (w *Wharfie) CreateService(jobid int) {
	srvAnnotations := map[string]string{"job.id": string(jobid)}
	rep := uint64(3)
	env := []string{
		"DOCKER_HOST", os.Getenv("DOCKER_HOST"),
		"DOCKER_CERT_PATH", os.Getenv("DOCKER_CERT_PATH"),
	}
	homeMount := mount.Mount{Source: "/home/", Target: "/home",}
	contSpec := swarm.ContainerSpec{
		Image: w.do.DockerImage,
		Command: []string{"tail", "-f", "/dev/null"},
		Mounts: []mount.Mount{homeMount},
		Env: env,
	}
	taskTemp := swarm.TaskSpec{
		ContainerSpec: contSpec,
		Placement: &swarm.Placement{Constraints: []string{
			fmt.Sprintf("job.id=%d", jobid),
		}},
	}
	srvSpec := swarm.ServiceSpec{
		Annotations: swarm.Annotations{Name: "srv", Labels: srvAnnotations},
		TaskTemplate: taskTemp,
		Mode: swarm.ServiceMode{Replicated: &swarm.ReplicatedService{Replicas: &rep}},
	}
	w.engCli.ServiceCreate(context.Background(), srvSpec, types.ServiceCreateOptions{})

}

func (w *Wharfie) Run() {
	w.Connect()
	w.AddJobIdLabel()
	w.RmJobIdLabel()
}