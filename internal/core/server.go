package core

import (
	genericapiserver "business-dev-bone/internal/framework/server"
	"business-dev-bone/pkg/component-base/log"
	"business-dev-bone/pkg/component-base/shutdown"
	"business-dev-bone/pkg/component-base/shutdown/shutdownmanagers/posixsignal"

	"business-dev-bone/internal/core/config"
)

type apiServer struct {
	gs      *shutdown.GracefulShutdown
	generic *genericapiserver.GenericAPIServer
}

func createServer(cfg *config.Config) (*apiServer, error) {
	gs := shutdown.New()
	gs.AddShutdownManager(posixsignal.NewPosixSignalManager())

	genericConfig, err := buildGenericConfig(cfg)
	if err != nil {
		return nil, err
	}
	generic, err := genericConfig.Complete().New()
	if err != nil {
		return nil, err
	}

	s := &apiServer{gs: gs, generic: generic}
	return s, nil
}

func (s *apiServer) prepareRun() *apiServer {
	installRoutes(s.generic.Engine)
	s.gs.AddShutdownCallback(shutdown.ShutdownFunc(func(string) error {
		s.generic.Close()
		return nil
	}))
	return s
}

func (s *apiServer) run() error {
	if err := s.gs.Start(); err != nil {
		log.Fatalf("shutdown manager start: %s", err.Error())
	}
	return s.generic.Run()
}

func buildGenericConfig(cfg *config.Config) (*genericapiserver.Config, error) {
	c := genericapiserver.NewConfig()
	o := cfg.Options
	if err := o.GenericServerRunOptions.ApplyTo(c); err != nil {
		return nil, err
	}
	if err := o.FeatureOptions.ApplyTo(c); err != nil {
		return nil, err
	}
	if err := o.SecureServing.ApplyTo(c); err != nil {
		return nil, err
	}
	if err := o.InsecureServing.ApplyTo(c); err != nil {
		return nil, err
	}
	return c, nil
}
