package wharfie

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"gopkg.in/fatih/set.v0"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/mount"


)


func (w *Wharfie) CreateNetwork() (id string, err error){
	netName := fmt.Sprintf("jobid-network%s",w.do.JobId)
	f := filters.NewArgs()
	f.Add("name", netName)
	nets, err := w.engCli.NetworkList(context.Background(), types.NetworkListOptions{Filters: f})
	for _, net := range nets {
		if net.Name == netName {
			w.Log("info", fmt.Sprintf("Network '%s' already existing... ", netName))
			return net.ID, err
		}
	}
	resp, err := w.engCli.NetworkCreate(context.Background(), netName, types.NetworkCreate{Driver: "overlay"})
	if err != nil {
		return "", err
	}
	id = resp.ID
	return
}

func (w *Wharfie) CreateService() (err error){
	srvAnnotations := map[string]string{"job.id": w.do.JobId}
	env := []string{}
		/*"DOCKER_HOST", os.Getenv("DOCKER_HOST"),
		"DOCKER_CERT_PATH", os.Getenv("DOCKER_CERT_PATH"),
	}*/
	if w.do.DockerImage == "" {
		w.Log("error", "No image defined, please set WHARFY_DOCKER_IMAGE or use --docker-image")
	}
	id, err := w.CreateNetwork()
	if err != nil {
		log.Fatalf("Could not create network for jobid: %s", err.Error())
	}
	mounts := []mount.Mount{}
	for _, vols := range w.do.Volumes {
		slice := strings.Split(vols, ":")
		switch len(slice) {
			case 0:
				continue
			case 2:
				mounts = append(mounts, mount.Mount{Source: slice[0], Target: slice[1]})
			default:
				log.Printf("WARN: Could not parse volume: %s", vols)
		}
	}
	contSpec := swarm.ContainerSpec{
		Image: w.do.DockerImage,
		Mounts: mounts,
		Env: env,
		Hostname: `{{.Service.Name}}.{{.Task.Slot}}.{{.Task.ID}}`,
		Dir: fmt.Sprintf("%s/%s", w.do.Homedir, w.do.Username),
		User: w.do.Username,
	}
	netConfig := []swarm.NetworkAttachmentConfig{}
	netConfig = append(netConfig, swarm.NetworkAttachmentConfig{Target: id})
	constraints := []string{
		fmt.Sprintf("node.labels.job.id.%s==true", w.do.JobId),
	}
	if w.do.Constraints != "" {
		log.Printf("Add Label '%s'", w.do.Constraints)
		constraints = append(constraints, w.do.Constraints)
	}
	taskTemp := swarm.TaskSpec{
		ContainerSpec: contSpec,
		Networks: netConfig,
		Placement: &swarm.Placement{Constraints: constraints},

	}
	srvSpec := swarm.ServiceSpec{
		Annotations: swarm.Annotations{Name: fmt.Sprintf("jobid%s",w.do.JobId), Labels: srvAnnotations},
		TaskTemplate: taskTemp,
		Mode: swarm.ServiceMode{Global: &swarm.GlobalService{}},
	}
	if w.do.Replicas != 0 {
		u := uint64(w.do.Replicas)
		srvSpec.Mode = swarm.ServiceMode{Replicated: &swarm.ReplicatedService{Replicas: &u}}
	}
	resp, err := w.engCli.ServiceCreate(ctx, srvSpec, types.ServiceCreateOptions{})
	log.Printf("Response: %v\n", resp)
	if err != nil {
		log.Printf("Error while ServiceCreate(): %s\n", err.Error())
	}
	return
}

func (w *Wharfie) RemoveService() (err error) {
	tasks, _ := w.GetTasks()
	srvName := fmt.Sprintf("jobid%s",w.do.JobId)
	err = w.engCli.ServiceRemove(ctx, srvName)
	if err != nil {
		log.Printf("Error during RemoveService(): %s\n", err.Error())
	} else {
		log.Printf("Service '%s' removed\n", srvName)
	}
	for {
		cid := set.New()
		for _, task := range tasks {
			log.Printf("Add %s (on %s) to set", task.Status.ContainerStatus.ContainerID, task.NodeID)
			cid.Add(task.Status.ContainerStatus.ContainerID)
		}
		containers, err := w.engCli.ContainerList(ctx, types.ContainerListOptions{})
		if err != nil {
			log.Fatal(err)
		}
		for _, cnt := range containers {
			if cid.Has(cnt.ID) {
				log.Printf("remove %s from set: %v", cnt.ID, cid.String())
				cid.Remove(cnt.ID)
			}
		}
		if cid.IsEmpty() {
			break
		}
	}
	return
}

func (w *Wharfie) GetTasks() (tasks []swarm.Task, err error){
	srvName := fmt.Sprintf("jobid%s",w.do.JobId)
	f := filters.NewArgs()
	f.Add("service", srvName)
	tasks, err = w.engCli.TaskList(context.Background(), types.TaskListOptions{Filters: f})
	return
}

func (w *Wharfie) GetServiceTask(node string) (task swarm.Task, err error){
	f := filters.NewArgs()
	f.Add("service", fmt.Sprintf("jobid%s",w.do.JobId))
	f.Add("node", node)
	f.Add("desired-state","running")
	tasks, err := w.engCli.TaskList(context.Background(), types.TaskListOptions{Filters: f})
	if err != nil {
		return swarm.Task{}, err
	}
	if len(tasks) != 1 {
		for _, task := range tasks {
			log.Printf("tid: %s // cid: %s", task.NodeID, task.Status.ContainerStatus.ContainerID)
		}
		return swarm.Task{}, fmt.Errorf("only a single task per node is allowed, found %d", len(tasks))
	}
	return tasks[0], err
}

func (w *Wharfie) WaitForService() (err error) {
	srvName := fmt.Sprintf("jobid%s",w.do.JobId)
	f := filters.NewArgs()
	f.Add("name", srvName)
	srvList, err := w.engCli.ServiceList(context.Background(), types.ServiceListOptions{Filters: f})
	if err != nil {
		log.Fatalf("Error during ServiceList()")
		return
	}
	if len(srvList) == 0 {
		return fmt.Errorf("No service found with name '%s'", srvName)
	}
	srv := srvList[0]
	srvInfo, _, err := w.engCli.ServiceInspectWithRaw(context.Background(), srv.ID)
	log.Printf("Service: %v\n", srvInfo.Spec.Name)

	old_line := ""
	// start listening for updates and render
	for {
		taskStatus := map[string]int{
			"scheduling": 0,
			"pending": 0,
			"starting": 0,
			"running": 0,
		}
		tasks, _ := w.GetTasks()
		for _, task := range tasks {
			taskStatus[string(task.Status.State)] += 1
		}
		statStr := []string{}
		for k, v := range taskStatus {
			if v != 0 {
				statStr = append(statStr, fmt.Sprintf("%s=%d", k, v))
			}
		}
		new_line := strings.Join(statStr, "/")
		if old_line != new_line {
			w.Log("debug",  new_line)
			old_line = new_line
		}
		if taskStatus["running"] == len(tasks) {
			break
		}
		time.Sleep(time.Duration(1)*time.Second)
	}
	return
}

