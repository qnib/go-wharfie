package wharfie

import "strings"

type Option func(*Options)

type Options struct {
	DockerSocket,DockerCertPath,DockerImage,Username,Homedir	string
	Constraints,JobId											string
	NodeList,Volumes											[]string
	Debug 														bool
	Replicas 													int
}

var defaultDracerOptions = Options {
	DockerSocket: DOCKER_API_HOST,
	DockerCertPath: DOCKER_CERT_PATH,
	Debug: Debug,
	Volumes: DOCKER_VOLUMES,
	Replicas: 0,
}

func WithNodeList(nl string) Option {
	return func(o *Options) {
		o.NodeList = strings.Split(nl, ",")
	}
}

func WithVolumes(s string) Option {
	return func(o *Options) {
		o.Volumes = strings.Split(s, ",")
	}
}
func WithReplicas(rep int) Option {
	return func(o *Options) {
		o.Replicas = rep
	}
}

func WithDockerSocket(s string) Option {
	return func(o *Options) {
		o.DockerSocket = s
	}
}

func WithUsername(s string) Option {
	return func(o *Options) {
		o.Username = s
	}
}

func WithConstraints(s string) Option {
	return func(o *Options) {
		o.Constraints = s
	}
}

func WithHomedir(s string) Option {
	return func(o *Options) {
		o.Homedir = s
	}
}

func WithJobId(s string) Option {
	return func(o *Options) {
		o.JobId = s
	}
}

func WithDockerImage(s string) Option {
	return func(o *Options) {
		o.DockerImage = s
	}
}

func WithDockerCertPath(s string) Option {
	return func(o *Options) {
		o.DockerCertPath = s
	}
}

func WithDebugValue(d bool) Option {
	return func(o *Options) {
		o.Debug = d
	}
}