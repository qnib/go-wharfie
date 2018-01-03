package wharfie

import (
	"log"
	"github.com/codegangsta/cli"
	"github.com/docker/docker/api/types"
	"strings"
	"fmt"
	"io"
	"bufio"
	"os/exec"
)

var (
	out, stderr io.Writer
)

func (w *Wharfie) Ssh(ctx *cli.Context) {
	w.Connect()
	log.Printf("%v", ctx.Args())
	node := ctx.Args()[0]
	task, err := w.GetServiceTask(node)
	if err != nil {
		log.Fatal(err.Error())
	}
	cmd := types.ExecConfig{
		User: "cluser", //fmt.Sprintf("%d", os.Getuid()),
		Privileged: false,
		Cmd: ctx.Args()[1:],
		Tty: true,
		//Env: os.Environ(),
	}
	cmdStr := fmt.Sprintf("docker exec -t -u %s %v %s", cmd.User, task.Status.ContainerStatus.ContainerID, strings.Join(cmd.Cmd, " "))
	log.Println(cmdStr)
	RunExec(cmdStr)
	/*
	cont := context.Background()
	eresp, err := w.engCli.ContainerExecCreate(cont, task.Status.ContainerStatus.ContainerID, cmd)
	if err != nil {
		log.Fatalf("%v", eresp)
	}
	log.Println("ExecCreate done..")
	err = w.engCli.ContainerExecStart(cont, eresp.ID, types.ExecStartCheck{Detach: false, Tty: true})
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	log.Println("ExecStart done..")
	log.Printf("Executing exec ID = %s", eresp.ID)
	_, err = w.engCli.ContainerExecAttach(cont, eresp.ID,
		types.ExecConfig{
			Detach: false,
			Tty:    true,
		},
	)
	//log.Printf("%s",string(r))
	resp, err := w.engCli.ContainerExecInspect(cont,eresp.ID)
	_ = resp
	os.Exit(0)
	*/
}

func RunExec(cmd string) {
	commd := exec.Command("bash", "-c", cmd)
	commd.Dir = "/tmp/"
	stdout, _ := commd.StdoutPipe()
	stderr, _ := commd.StderrPipe()
	commd.Start()
	ch := make(chan string, 100)
	stdoutScan := bufio.NewScanner(stdout)
	stderrScan := bufio.NewScanner(stderr)
	go func() {
		for stdoutScan.Scan() {
			line := stdoutScan.Text()
			ch <- line
		}
	}()
	go func() {
		for stderrScan.Scan() {
			line := stderrScan.Text()
			ch <- line
		}
	}()
	go func() {
		commd.Wait()
		close(ch)
	}()
	for line := range ch {
		fmt.Println(line)
	}
}