package wharfie


type Option func(*Options)

type Options struct {
	DockerSocket,DockerCertPath	string
	Debug 						bool
}

var defaultDracerOptions = Options {
	DockerSocket: DOCKER_API_HOST,
	DockerCertPath: DOCKER_CERT_PATH,
	Debug: Debug,
}

func WithDockerSocket(s string) Option {
	return func(o *Options) {
		o.DockerSocket = s
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