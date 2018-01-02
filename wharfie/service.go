package wharfie

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	//"os"

	"gopkg.in/fatih/set.v0"
	"github.com/gosuri/uilive"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/mount"

)

func (w *Wharfie) CreateService() (err error){
	srvAnnotations := map[string]string{"job.id": w.do.JobId}
	env := []string{
	//	"DOCKER_HOST", os.Getenv("DOCKER_HOST"),
	//	"DOCKER_CERT_PATH", os.Getenv("DOCKER_CERT_PATH"),
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
			fmt.Sprintf("node.labels.job.id.%s==true", w.do.JobId),
		}},
	}
	srvSpec := swarm.ServiceSpec{
		Annotations: swarm.Annotations{Name: fmt.Sprintf("JobID%s",w.do.JobId), Labels: srvAnnotations},
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
	srvName := fmt.Sprintf("JobID%s",w.do.JobId)
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
	srvName := fmt.Sprintf("JobID%s",w.do.JobId)
	f := filters.NewArgs()
	f.Add("service", srvName)
	tasks, err = w.engCli.TaskList(context.Background(), types.TaskListOptions{Filters: f})
	return
}

func (w *Wharfie) GetServiceTask(node string) (task swarm.Task, err error){
	f := filters.NewArgs()
	f.Add("service", "job_openmpi") //fmt.Sprintf("JobID%s",w.do.JobId)
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
	srvName := fmt.Sprintf("JobID%s",w.do.JobId)
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

	writer := uilive.New()
	// start listening for updates and render
	writer.Start()
	defer writer.Stop()
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
		fmt.Fprintf(writer, "%s\n",strings.Join(statStr, "/"))
		if taskStatus["running"] == len(tasks) {
			break
		}
		time.Sleep(time.Duration(1)*time.Second)
	}
	return
}

