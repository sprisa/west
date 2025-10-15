package west

import (
	"context"
	"errors"
	"fmt"

	"github.com/sprisa/west/config"

	"github.com/sirupsen/logrus"
	"github.com/slackhq/nebula"
	nebulaCfg "github.com/slackhq/nebula/config"
	"github.com/slackhq/nebula/overlay"
	"github.com/slackhq/nebula/util"
	"gopkg.in/yaml.v3"
)

// A version string that can be set with
//
//	-ldflags "-X main.Build=SOMEVERSION"
//
// at compile-time.
var Build string = "west v0.0.1"

type Control = nebula.Control
type NebulaConfigCtrl = nebulaCfg.C

type OnStartFunc = func(*Control)

// TODO: Add add an `onReady` and wait for Nebula to be fully up.
// e.g. after "handshake message received"
type ServerOpts struct {
	// Custom Logging
	Log *logrus.Logger
	// Nebula config
	Config *config.Config
	// Hook called after Nebula server starts
	OnStart OnStartFunc
	// Hook called before Nebula shuts down.
	// Will block until function returns.
	OnShutdown    func()
	deviceFactory overlay.DeviceFactory
}

type Server struct {
	Ctrl *Control
	opts *ServerOpts
}

func NewServer(opts *ServerOpts) (*Server, error) {
	log := opts.Log
	if log == nil {
		log = logrus.StandardLogger()
	}
	nebulaConfig := opts.Config
	if nebulaConfig == nil {
		return nil, errors.New("expected west config")
	}

	c, err := CreateNebulaConfigCtrl(nebulaConfig, log)
	if err != nil {
		return nil, err
	}

	return NewServerWithConfigCtrl(opts, c)
}

func NewServerWithConfigCtrl(opts *ServerOpts, c *NebulaConfigCtrl) (*Server, error) {
	if opts.Log == nil {
		opts.Log = logrus.StandardLogger()
	}
	nebulaConfig := opts.Config
	if nebulaConfig == nil {
		return nil, errors.New("expected west config")
	}

	port := nebulaConfig.Listen.Port

	// TODO: Ensure mknod is still upstream

	ctrl, err := nebula.Main(c, false, Build, opts.Log, opts.deviceFactory, nil)

	if err != nil {
		switch v := err.(type) {
		case *util.ContextualError:
			v.Log(opts.Log)
			return nil, v.Unwrap()
		default:
			// TODO: Move this port error decoration up into harbor
			return nil, fmt.Errorf("nebula error on port %d: %w", port, err)
		}
	}

	return &Server{Ctrl: ctrl, opts: opts}, nil
}

func CreateNebulaConfigCtrl(cfg *config.Config, log *logrus.Logger) (*NebulaConfigCtrl, error) {
	nebulaYaml, err := configToYaml(cfg)
	if err != nil {
		return nil, err
	}
	nebulaYamlStr := string(nebulaYaml)
	println("CONFIG")
	println(nebulaYamlStr)

	c := nebulaCfg.NewC(log)
	// Can use ReloadConfigString to handle static-host-map changes
	// c.ReloadConfigString()
	// TODO: Could just turn the struct into a record map
	// or commit the types upstream.
	err = c.LoadString(nebulaYamlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to load nebula config: %w", err)
	}

	return c, nil
}

// TODO: Not sure if I want this and instead just use a JWT to provision
func NewServerWithYamlConfig(args *ServerOpts, yamlCfg []byte) (*Server, error) {
	if args.Log == nil {
		args.Log = logrus.StandardLogger()
	}
	c := nebulaCfg.NewC(args.Log)
	// TODO: Just run yaml unmarshel to the any map and set in Settings
	err := c.LoadString(string(yamlCfg))
	if err != nil {
		return nil, fmt.Errorf("failed to load nebula config: %w", err)
	}

	ctrl, err := nebula.Main(c, false, Build, args.Log, args.deviceFactory, nil)
	if err != nil {
		switch v := err.(type) {
		case *util.ContextualError:
			v.Log(args.Log)
			return nil, v.Unwrap()
		default:
			return nil, err
		}
	}

	// TODO: Use mergo and merge west config from the yaml version
	return &Server{Ctrl: ctrl, opts: args}, nil
}

func (s *Server) Listen(ctx context.Context) error {
	log := s.opts.Log
	port := s.opts.Config.Listen.Port
	// Start Nebula Server
	s.Ctrl.Start()
	// Wait for OnStart hook
	if s.opts.OnStart != nil {
		s.opts.OnStart(s.Ctrl)
	}
	// Wait for server to be stopped (context cancelled)
	<-ctx.Done()
	// Wait for OnShutdown hook
	if s.opts.OnShutdown != nil {
		s.opts.OnShutdown()
	}
	log.Infof("Shutting down nebula server on port %d \n", port)
	s.Ctrl.Stop()
	return nil
}

func (s *Server) IFaceName() string {
	return s.Ctrl.Device().Name()
}

func configToYaml(cfg *config.Config) ([]byte, error) {
	return yaml.Marshal(cfg)
}
