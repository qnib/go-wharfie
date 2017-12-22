package main

import (
	"github.com/qnib/go-wharfie/wharfie"
	"log"
	"github.com/zpatrick/go-config"
	"github.com/codegangsta/cli"

	"os"
)

var (
	dockerSocketFlag = cli.StringFlag{
		Name:  "docker-host",
		Usage: "Docker host to connect to.",
		EnvVar: "DOCKER_HOST",
	}
	dockerCertFlag = cli.StringFlag{
		Name:  "docker-cert-path",
		Usage: "Path to certificates.",
		EnvVar: "DOCKER_CERT_PATH",
	}
	debugFlag = cli.BoolFlag{
		Name: "debug",
		Usage: "Print proxy requests",
		EnvVar: "WHARFY_DEBUG",
	}
	dockerImageFlag = cli.StringFlag{
		Name:  "docker-image",
		Usage: "Docker Image to use for JOB.",
	}
	jobIdFlag = cli.StringFlag{
		Name:  "job-id",
		Usage: "Job ID.",
		EnvVar: "SLURM_JOB_ID",
	}
	nodeListFlag = cli.StringFlag{
		Name:  "node-list",
		Usage: "Comma separated list of nodes (container names)",
		EnvVar: "WHARFY_NODE_LIST",
	}

)

func EvalOptions(cfg *config.Config) (po []wharfie.Option) {
	dockerSock, _ := cfg.String("docker-host")
	po = append(po, wharfie.WithDockerSocket(dockerSock))
	dockerImage, _ := cfg.String("docker-image")
	po = append(po, wharfie.WithDockerImage(dockerImage))
	nodeList, _ := cfg.String("node-list")
	po = append(po, wharfie.WithNodeList(nodeList))
	jobId, _ := cfg.String("job-id")
	po = append(po, wharfie.WithJobId(jobId))
	dockerCertPath, _ := cfg.String("docker-cert-path")
	po = append(po, wharfie.WithDockerCertPath(dockerCertPath))
	debug, _ := cfg.Bool("debug")
	po = append(po, wharfie.WithDebugValue(debug))
	return
}

func RunApp(ctx *cli.Context) {
	log.Printf("[II] Start Version: %s", ctx.App.Version)
	cfg := config.NewConfig([]config.Provider{config.NewCLI(ctx, true)})
	po := EvalOptions(cfg)
	p := wharfie.New(po...)
	p.Run()
}

func main() {
	app := cli.NewApp()
	app.Name = "CLI to help mpirun to use docker container."
	app.Usage = "go-wharfie [options]"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		debugFlag,
		dockerSocketFlag,
		dockerImageFlag,
		jobIdFlag,
		nodeListFlag,
	}
	app.Action = RunApp
	app.Run(os.Args)
}
