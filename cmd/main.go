package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/danvixent/aboki-africa-assessment/routes"
	"github.com/dimfeld/httptreemux"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danvixent/aboki-africa-assessment/config"
	"github.com/danvixent/aboki-africa-assessment/datastore/postgres"
	"github.com/danvixent/aboki-africa-assessment/handler"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var configPath *string

func init() {
	configPath = flag.String("config_path", "", "path to config file")
	flag.Parse()
	if configPath == nil {
		log.Fatalln("-config_path flag is required")
	}
}

func main() {
	file, err := os.Open(*configPath)
	if err != nil {
		log.Fatalf("unable to open config file: %v", err)
	}

	cfg := &config.BaseConfig{}
	err = yaml.NewDecoder(file).Decode(cfg)
	if err != nil {
		log.Fatalf("failed to decode config file: %v", err)
	}

	postgresClient := postgres.New(context.Background(), cfg.Postgres)
	userRepo := postgres.NewUserRepository(postgresClient)
	userReferralRepo := postgres.NewUserReferralRepository(postgresClient)
	userPointsRepo := postgres.NewUserPointsRepository(postgresClient)

	h := handler.NewHandler(userRepo, userReferralRepo, userPointsRepo, postgresClient.BeginTx)

	router := httptreemux.New()
	routes.SetupRoutes(router, h)

	srv := &http.Server{
		Addr:    ":" + cfg.ServePort,
		Handler: router,
	}

	log.Printf("serving at http://localhost:%s", cfg.ServePort)

	// start server in new goroutine so we can listen for CLI signals
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Panicf("unable to listen: %s", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so no need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Print("shutdown server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}
	select {
	case <-ctx.Done():
		log.Print("timeout of 1 seconds.")
	}
	log.Print("server exiting")
}

func unmarshalRequestBody(respBody io.ReadCloser, data interface{}) error {
	responseData, err := ioutil.ReadAll(respBody)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(responseData, data); err != nil {
		return err
	}

	return nil
}
