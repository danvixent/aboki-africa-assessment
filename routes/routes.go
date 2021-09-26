package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/danvixent/aboki-africa-assessment/handler"
	"github.com/dimfeld/httptreemux"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
)

func SetupRoutes(router *httptreemux.TreeMux, h *handler.Handler) {
	router.POST("/register", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		req := &handler.UserRequest{}
		err := getRequestBody(r.Body, req)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to parse request body: %v", err), http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}

		if req.Email == "" {
			http.Error(w, "email is required", http.StatusBadRequest)
			return
		}

		logger := log.WithFields(map[string]interface{}{})
		user, err := h.RegisterUser(context.Background(), req, logger)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		buf, err := json.Marshal(user)
		if err != nil {
			http.Error(w, "failed to marshal response", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(buf)
	})

	router.POST("/transaction", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		req := &handler.TransferPointsRequest{}
		err := getRequestBody(r.Body, req)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to parse request body: %v", err), http.StatusBadRequest)
			return
		}

		if req.UserID == "" {
			http.Error(w, "user id is required", http.StatusBadRequest)
			return
		}

		if req.RecipientUserID == "" {
			http.Error(w, "recipient user id is required", http.StatusBadRequest)
			return
		}

		if req.Points == 0 {
			http.Error(w, "points cannot be zero is required", http.StatusBadRequest)
			return
		}

		logger := log.WithFields(map[string]interface{}{})
		err = h.TransferPoints(context.Background(), req, logger)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}

func getRequestBody(respBody io.ReadCloser, data interface{}) error {
	buf, err := ioutil.ReadAll(respBody)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(buf, data); err != nil {
		return err
	}

	return nil
}
