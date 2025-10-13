package nebula

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

// TODO: Add add an `onReady` and wait for Nebula to be fully up.
// e.g. after "handshake message received"
type ServerArgs struct {
	// Custom Logging
	Log *logrus.Logger
	// Nebula config
	Config *config.Config
	// Hook called after Nebula server starts
	OnStart func(*Control)
	// Hook called before Nebula shuts down.
	// Will block until function returns.
	OnShutdown    func()
	deviceFactory overlay.DeviceFactory
}

type Server struct {
	Ctrl *Control
	args *ServerArgs
}

func NewServer(args *ServerArgs) (*Server, error) {
	log := args.Log
	if log == nil {
		return nil, errors.New("expected Log arg")
	}
	nebulaConfig := args.Config
	if nebulaConfig == nil {
		return nil, errors.New("expected west config")
	}

	c, err := CreateNebulaConfigCtrl(nebulaConfig, log)
	if err != nil {
		return nil, err
	}

	return NewServerWithConfigCtrl(args, c)
}

func NewServerWithConfigCtrl(args *ServerArgs, c *NebulaConfigCtrl) (*Server, error) {
	log := args.Log
	if log == nil {
		log = logrus.StandardLogger()
	}
	nebulaConfig := args.Config
	if nebulaConfig == nil {
		return nil, errors.New("expected west config")
	}

	port := nebulaConfig.Listen.Port

	// TODO: Ensure mknod is still upstream

	ctrl, err := nebula.Main(c, false, Build, log, args.deviceFactory, nil)

	if err != nil {
		switch v := err.(type) {
		case *util.ContextualError:
			v.Log(log)
			return nil, v.Unwrap()
		default:
			// TODO: Move this port error decoration up into harbor
			return nil, fmt.Errorf("nebula error on port %d: %w", port, err)
		}
	}

	return &Server{Ctrl: ctrl, args: args}, nil
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
func NewServerWithYamlConfig(args *ServerArgs, yamlCfg []byte) (*Server, error) {
	log := args.Log
	if log == nil {
		log = logrus.StandardLogger()
	}
	c := nebulaCfg.NewC(log)
	// TODO: Just run yaml unmarshel to the any map and set in Settings
	err := c.LoadString(string(yamlCfg))
	if err != nil {
		return nil, fmt.Errorf("failed to load nebula config: %w", err)
	}

	ctrl, err := nebula.Main(c, false, Build, log, args.deviceFactory, nil)
	if err != nil {
		switch v := err.(type) {
		case *util.ContextualError:
			v.Log(log)
			return nil, v.Unwrap()
		default:
			return nil, err
		}
	}

	// TODO: Use mergo and merge west config from the yaml version
	return &Server{Ctrl: ctrl, args: args}, nil
}

func (s *Server) Listen(ctx context.Context) error {
	log := s.args.Log
	port := s.args.Config.Listen.Port
	// Start Nebula Server
	s.Ctrl.Start()
	// Wait for OnStart hook
	if s.args.OnStart != nil {
		s.args.OnStart(s.Ctrl)
	}
	// Wait for server to be stopped (context cancelled)
	<-ctx.Done()
	// Wait for OnShutdown hook
	if s.args.OnShutdown != nil {
		s.args.OnShutdown()
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
