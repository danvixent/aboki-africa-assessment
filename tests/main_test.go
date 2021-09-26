package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	app "github.com/danvixent/aboki-africa-assessment"
	"github.com/danvixent/aboki-africa-assessment/routes"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/danvixent/aboki-africa-assessment/config"
	"github.com/danvixent/aboki-africa-assessment/datastore/postgres"
	"github.com/danvixent/aboki-africa-assessment/handler"
	"github.com/dimfeld/httptreemux"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var url = "http://localhost:%s"

type TestHandler struct {
	userRepository         app.UserRepository
	userReferralRepository app.UserReferralRepository
	userPointRepository    app.UserPointRepository
	client                 *postgres.Client
}

var testHandler *TestHandler

func TestMain(m *testing.M) {
	file, err := os.Open("../config/config.yml")
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

	url = fmt.Sprintf(url, cfg.ServePort)
	srv := &http.Server{
		Addr:    ":" + cfg.ServePort,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Panicf("unable to listen: %s", err)
		}
	}()

	// allow the goroutine above start the server
	time.Sleep(time.Second)

	testHandler = &TestHandler{
		userRepository:         userRepo,
		userReferralRepository: userReferralRepo,
		userPointRepository:    userPointsRepo,
		client:                 postgresClient,
	}
	// run the tests
	code := m.Run()

	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("unable to shutdown server gracefully: %v", err)
	}

	os.Exit(code)
}

// serialize obj into json bytes
func serialize(obj interface{}) *bytes.Buffer {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(obj); err != nil {
		log.Fatalf("unable to serialize obj: %v", err)
	}
	return buf
}

func deleteAllFromTable(name string) error {
	_, err := testHandler.client.Exec(context.Background(), fmt.Sprintf("DELETE FROM %s", name))
	return err
}
