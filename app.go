package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi"

	"github.com/vladislav-chunikhin/lib-go/pkg/healthcheck"
	"github.com/vladislav-chunikhin/lib-go/pkg/logger"
	"github.com/vladislav-chunikhin/lib-go/pkg/shutdown"
	"github.com/vladislav-chunikhin/lib-go/pkg/startup"
)

type App struct {
	Config      *Config
	Logger      logger.Logger
	Context     context.Context
	Closer      shutdown.Closer
	HealthCheck healthcheck.HealthChecker
	HttpServer  *http.Server

	cancel  context.CancelFunc
	readyCh chan struct{}

	httpMtServer *http.Server
}

func NewApp() *App {
	cfg := createFromEnv()

	ctx, cancel := context.WithCancel(context.Background())
	app := &App{
		Context:     ctx,
		HealthCheck: healthcheck.NewHealthCheck(),
		Closer:      shutdown.New(),
		Config:      cfg,
		cancel:      cancel,
		readyCh:     make(chan struct{}, 1),
	}

	return app
}

func (a *App) LoadConfig(configStruct interface{}, defCfgFile string) error {
	if configStruct == nil {
		return fmt.Errorf("the service's configuration must be adjusted")
	}

	if a.Config.ConfigFile == "" {
		a.Config.ConfigFile = defCfgFile
	}

	return configure(configStruct, a.Config.ConfigFile)
}

func (a *App) Init() error {
	err := a.init()
	if err != nil {
		return err
	}

	a.initHTTP()
	a.initHTTPMaintenance()

	if err = startup.SetMaxGoProcs(a.Logger); err != nil {
		a.Logger.Fatalf("set max go procs error: %v", err)
	}

	return nil
}

func (a *App) init() (err error) {
	a.Logger, err = logger.New(a.Config.LoggerLevel)
	if err != nil {
		return fmt.Errorf("logger create error: %s", err)
	}

	return nil
}

func (a *App) initHTTP() {
	a.HttpServer = &http.Server{
		ReadTimeout:  time.Duration(a.Config.HTTPServerReadTimeout) * time.Second,
		WriteTimeout: time.Duration(a.Config.HTTPServerWriteTimeout) * time.Second,
	}

	a.Logger.Infof("init http server success")
}

func (a *App) initHTTPMaintenance() {
	r := chi.NewRouter()

	a.HealthCheck.RegisterLive("ctx.Done", healthcheck.ContextDone(a.Context))

	r.Mount("/health", healthcheck.Handler(a.HealthCheck))

	a.httpMtServer = &http.Server{
		Handler:           r,
		ReadHeaderTimeout: time.Second,
	}

	a.Logger.Infof("init maintenance http server success")
}

func (a *App) Run() {
	err := a.runHTTP()
	if err != nil {
		a.Logger.Fatalf("run http server error: %v", err)
	}

	err = a.runHTTPMaintenance()
	if err != nil {
		a.Logger.Fatalf("run http maintenance server error: %v", err)
	}

	close(a.readyCh)

	a.Logger.Infof("app wait signal")
	err = a.Closer.Wait(a.Context)
	if err != nil && !errors.Is(err, shutdown.ErrTermSig) && !errors.Is(err, context.Canceled) {
		a.Logger.Fatalf("failed to caught signal: %v", err)
	}
	a.cancel()
	a.Logger.Infof("termination signal received")

	a.Closer.CloseAll(a.Logger)

	a.Logger.Infof("app stopped")
}

func (a *App) runHTTP() error {
	if a.HttpServer.Handler == nil {
		a.Logger.Infof("handler of httpServer not set")
		return nil
	}

	r := chi.NewMux()
	r.Mount("/api", a.HttpServer.Handler)

	a.HttpServer.Handler = r

	lsPub, err := net.Listen("tcp", ":"+a.Config.Port)
	if err != nil {
		return fmt.Errorf("listen address %s error: %v", ":"+a.Config.Port, err)
	}
	go func() {
		a.Logger.Infof("http server serve on address %s", lsPub.Addr())
		err = a.HttpServer.Serve(lsPub)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Logger.Errorf("http server serve error: %v", err)
		}
	}()

	a.Closer.Add(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		err = a.HttpServer.Shutdown(ctx)
		if err != nil {
			return fmt.Errorf("http server shutdown error: %v", err)
		}
		a.Logger.Infof("http server shutdown")
		return nil
	})
	return err
}

func (a *App) runHTTPMaintenance() error {
	lsMt, err := net.Listen("tcp", ":"+a.Config.MtPort)
	if err != nil {
		return fmt.Errorf("listen address %s error: %v", ":"+a.Config.MtPort, err)
	}

	go func() {
		a.Logger.Infof("http maintenance server serve on address %s", lsMt.Addr())
		err := a.httpMtServer.Serve(lsMt)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Logger.Errorf("http maintenance server serve error: %v", err)
		}
	}()

	a.Closer.Add(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		err := a.httpMtServer.Shutdown(ctx)
		if err != nil {
			return fmt.Errorf("http maintenance server shutdown error: %v", err)
		}
		a.Logger.Infof("http maintenance server shutdown")
		return nil
	})
	return err
}

func (a *App) Shutdown() {
	a.cancel()
}

func (a *App) WaitReady() {
	<-a.readyCh
}
