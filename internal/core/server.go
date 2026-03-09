package core

import (
	genericoptions "business-dev-bone/internal/pkg/options"
	genericapiserver "business-dev-bone/internal/pkg/server"
	"business-dev-bone/internal/pkg/storage"
	"business-dev-bone/pkg/component-base/log"
	"business-dev-bone/pkg/component-base/shutdown"
	"business-dev-bone/pkg/component-base/shutdown/shutdownmanagers/posixsignal"
	"context"
	"fmt"
	"time"

	"business-dev-bone/internal/core/config"
)

type apiServer struct {
	gs           *shutdown.GracefulShutdown
	generic      *genericapiserver.GenericAPIServer
	redisOptions *genericoptions.RedisOptions
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

	s := &apiServer{
		gs:           gs,
		generic:      generic,
		redisOptions: cfg.RedisOptions,
	}
	return s, nil
}

func (s *apiServer) prepareRun() *apiServer {
	installRoutes(s.generic.Engine)
	if err := s.initRedisStore(); err != nil {
		log.Fatalf("init redis store failed: %v", err)
	}

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

func (s *apiServer) initRedisStore() error {
	ctx, cancel := context.WithCancel(context.Background())
	s.gs.AddShutdownCallback(shutdown.ShutdownFunc(func(string) error {
		cancel()

		return nil
	}))

	config := &storage.Config{
		Host:                  s.redisOptions.Host,
		Port:                  s.redisOptions.Port,
		Addrs:                 s.redisOptions.Addrs,
		MasterName:            s.redisOptions.MasterName,
		Username:              s.redisOptions.Username,
		Password:              s.redisOptions.Password,
		Database:              s.redisOptions.Database,
		MaxIdle:               s.redisOptions.MaxIdle,
		MaxActive:             s.redisOptions.MaxActive,
		Timeout:               s.redisOptions.Timeout,
		EnableCluster:         s.redisOptions.EnableCluster,
		UseSSL:                s.redisOptions.UseSSL,
		SSLInsecureSkipVerify: s.redisOptions.SSLInsecureSkipVerify,
	}

	// try to connect to redis
	go storage.ConnectToRedis(ctx, config)

	// wait for redis to be connected
	for i := 0; i < 30; i++ {
		if storage.Connected() {
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("redis connection timeout")
}
