package options

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-envconfig"
	"golang.org/x/exp/slices"
)

const (
	DefaultControlPlane = "localhost:9090"
	e2coreEnvPrefix     = "E2CORE"
	FeatureMultiTenant  = "adminV1"
)

// Options defines options for E2Core.
type Options struct {
	Features         []string      `env:"E2CORE_API_FEATURES"`
	BundlePath       string        `env:"E2CORE_BUNDLE_PATH"`
	RunSchedules     *bool         `env:"E2CORE_RUN_SCHEDULES,default=true"`
	ControlPlane     string        `env:"E2CORE_CONTROL_PLANE"`
	AuthCacheTTL     time.Duration `env:"E2CORE_AUTH_CACHE_TTL,default=10m"`
	UpstreamAddress  string        `env:"E2CORE_UPSTREAM_ADDRESS"`
	EnvironmentToken string        `env:"E2CORE_ENV_TOKEN"`
	StaticPeers      string        `env:"E2CORE_PEERS"`
	AppName          string        `env:"E2CORE_APP_NAME,default=E2Core"`
	Domain           string        `env:"E2CORE_DOMAIN"`
	HTTPPort         int           `env:"E2CORE_HTTP_PORT,default=8080"`
	TLSPort          int           `env:"E2CORE_TLS_PORT,default=443"`
	TracerConfig     TracerConfig  `env:",prefix=E2CORE_TRACER_"`
}

// TracerConfig holds values specific to setting up the tracer. It's only used in proxy mode. All configuration options
// have a prefix of E2CORE_TRACER_ specified in the parent Options struct.
type TracerConfig struct {
	TracerType      string           `env:"TYPE,default=none"`
	ServiceName     string           `env:"SERVICENAME,default=e2core"`
	Probability     float64          `env:"PROBABILITY,default=0.5"`
	Collector       *CollectorConfig `env:",prefix=COLLECTOR_,noinit"`
	HoneycombConfig *HoneycombConfig `env:",prefix=HONEYCOMB_,noinit"`
}

// CollectorConfig holds config values specific to the collector tracer exporter running locally / within your cluster.
// All the configuration values here have a prefix of E2CORE_TRACER_COLLECTOR_, specified in the top level Options struct,
// and the parent TracerConfig struct.
type CollectorConfig struct {
	Endpoint string `env:"ENDPOINT"`
}

// HoneycombConfig holds config values specific to the honeycomb tracer exporter. All the configuration values here have
// a prefix of E2CORE_TRACER_HONEYCOMB_, specified in the top level Options struct, and the parent TracerConfig struct.
type HoneycombConfig struct {
	Endpoint string `env:"ENDPOINT"`
	APIKey   string `env:"APIKEY"`
	Dataset  string `env:"DATASET"`
}

// Modifier defines options for E2Core.
type Modifier func(*Options)

func NewWithModifiers(mods ...Modifier) (*Options, error) {
	opts := &Options{}

	for _, mod := range mods {
		mod(opts)
	}

	err := opts.finalize()
	if err != nil {
		return nil, errors.Wrap(err, "opts.finalize")
	}

	return opts, nil
}

// UseBundlePath sets the bundle path to be used.
func UseBundlePath(path string) Modifier {
	return func(opts *Options) {
		opts.BundlePath = path
	}
}

// AppName sets the app name to be used.
func AppName(name string) Modifier {
	return func(opts *Options) {
		opts.AppName = name
	}
}

// Domain sets the domain to be used.
func Domain(domain string) Modifier {
	return func(opts *Options) {
		opts.Domain = domain
	}
}

// HTTPPort sets the http port to be used.
func HTTPPort(port int) Modifier {
	return func(opts *Options) {
		opts.HTTPPort = port
	}
}

// TLSPort sets the tls port to be used.
func TLSPort(port int) Modifier {
	return func(opts *Options) {
		opts.TLSPort = port
	}
}

// finalize "locks in" the options by overriding any existing options with the version from the environment, and setting the default logger if needed.
func (o *Options) finalize() error {
	envOpts := Options{}
	if err := envconfig.Process(context.Background(), &envOpts); err != nil {
		return errors.Wrap(err, "envconfig.Process")
	}

	o.ControlPlane = strings.TrimSuffix(envOpts.ControlPlane, "/")
	o.AuthCacheTTL = envOpts.AuthCacheTTL

	// set RunSchedules if it was not passed as a flag.
	if o.RunSchedules == nil {
		if envOpts.RunSchedules != nil {
			o.RunSchedules = envOpts.RunSchedules
		}
	}

	// set AppName if it was not passed as a flag.
	if o.AppName == "" {
		o.AppName = envOpts.AppName
	}

	// set Domain if it was not passed as a flag.
	if o.Domain == "" {
		o.Domain = envOpts.Domain
	}

	// set HTTPPort if it was not passed as a flag.
	if o.HTTPPort == 0 {
		o.HTTPPort = envOpts.HTTPPort
	}

	// set TLSPort if it was not passed as a flag.
	if o.TLSPort == 0 {
		o.TLSPort = envOpts.TLSPort
	}

	o.Features = envOpts.Features
	o.EnvironmentToken = ""
	o.TracerConfig = TracerConfig{}
	o.StaticPeers = envOpts.StaticPeers

	o.EnvironmentToken = envOpts.EnvironmentToken
	o.TracerConfig = envOpts.TracerConfig

	return nil
}

func (o *Options) AdminEnabled() bool {
	return slices.Contains(o.Features, FeatureMultiTenant)
}
