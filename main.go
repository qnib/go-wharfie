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
		EnvVar: "WHARFIE_DEBUG",
	}
	bundleFlag = cli.BoolFlag{
		Name: "bundle",
		Usage: "Use UCP client bundle",
		EnvVar: "WHARFIE_BUNDLE",
	}
	dockerImageFlag = cli.StringFlag{
		Name:  "docker-image",
		Usage: "Docker Image to use for JOB.",
		EnvVar: "WHARFIE_DOCKER_IMAGE",
	}
	jobIdFlag = cli.StringFlag{
		Name:  "job-id",
		Usage: "Job ID.",
		EnvVar: "SLURM_JOB_ID",
	}
	constraintsFlag = cli.StringFlag{
		Name:  "constraints",
		Usage: "Label to limit container on.",
		EnvVar: "WHARFIE_CONSTRAINTS",
	}
	usernameFlag = cli.StringFlag{
		Name:  "username",
		Usage: "username to define home-dir.",
		Value: "cluser",
		EnvVar: "WHARFIE_USERNAME",
	}
	userFlag = cli.StringFlag{
		Name:  "user",
		Usage: "UID[:GID] to run container with.",
		Value: "1000:1000",
		EnvVar: "WHARFIE_USER",
	}
	homedirFlag = cli.StringFlag{
		Name:  "homedir",
		Usage: "Homedir prefix. Workdir of containers are going to be '${homedir}/${username}'.",
		Value: "/home",
		EnvVar: "WHARFIE_HOMEDIR",
	}
	replicaFlag = cli.IntFlag{
		Name:  "replicas",
		Usage: "Service replicas, 0 creates a global service.",
		Value: 0,
		EnvVar: "WHARFIE_SERVICE_REPLICAS",
	}
	mountsFlag = cli.StringFlag{
		Name:  "volumes",
		Usage: "Comma separated list of bind-mounts",
		EnvVar: "WHARFIE_VOLUMES",
	}

	nodeListFlag = cli.StringFlag{
		Name:  "node-list",
		Usage: "Comma separated list of nodes (container names)",
		EnvVar: "WHARFIE_NODE_LIST",
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
	bundle, _ := cfg.Bool("bundle")
	po = append(po, wharfie.WithBundleValue(bundle))
	replicas, _ := cfg.Int("replicas")
	po = append(po, wharfie.WithReplicas(replicas))
	vols, _ := cfg.String("volumes")
	po = append(po, wharfie.WithVolumes(vols))
	hdir, _ := cfg.String("homedir")
	po = append(po, wharfie.WithHomedir(hdir))
	uname, _ := cfg.String("username")
	po = append(po, wharfie.WithUsername(uname))
	user, _ := cfg.String("user")
	po = append(po, wharfie.WithUser(user))
	constraints , _ := cfg.String("constraints")
	po = append(po, wharfie.WithConstraints(constraints))

	return
}

func SshTasks(ctx *cli.Context) {
	log.Printf("[II] Start Version: %s", ctx.App.Version)
	cfg := config.NewConfig([]config.Provider{config.NewCLI(ctx, true)})
	po := EvalOptions(cfg)
	p := wharfie.New(po...)
	p.Ssh(ctx)
}

func StageService(ctx *cli.Context) {
	log.Printf("[II] Start Version: %s", ctx.App.Version)
	cfg := config.NewConfig([]config.Provider{config.NewCLI(ctx, true)})
	po := EvalOptions(cfg)
	p := wharfie.New(po...)
	p.Stage()
}

func RemoveService(ctx *cli.Context) {
	log.Printf("[II] Start Version: %s", ctx.App.Version)
	cfg := config.NewConfig([]config.Provider{config.NewCLI(ctx, true)})
	po := EvalOptions(cfg)
	p := wharfie.New(po...)
	p.Remove()
}

func main() {
	app := cli.NewApp()
	app.Name = "CLI to help mpirun to use docker container."
	app.Usage = "go-wharfie [options]"
	app.Version = "0.1.5"
	app.Flags = []cli.Flag{
		debugFlag,
		bundleFlag,
		dockerSocketFlag,
		jobIdFlag,
		nodeListFlag,
	}
	app.Commands = []cli.Command{
		{
			Name:    "stage",
			Usage:   "Create service and wait for all tasks to be up.",
			Action: StageService,
			Flags: []cli.Flag{
				mountsFlag,
				replicaFlag,
				usernameFlag,
				userFlag,
				homedirFlag,
				dockerImageFlag,
				constraintsFlag,
			},
		},{
			Name:    "remove",
			Usage:   "Remove service and wait for all tasks to be removed.",
			Action: RemoveService,
		},
	}
	app.Action = SshTasks
	app.Run(os.Args)
}
