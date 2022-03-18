package modserver

import (
	"errors"
	nethttp "net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"

	"github.com/mercari/modserver/internal/handler/http"
	"github.com/mercari/modserver/internal/repository"
	"github.com/mercari/modserver/internal/usecase"
)

// NewHandler initializes a new http.Handler based on the given server configuration.
func NewHandler(cfg *Config) (nethttp.Handler, error) {
	r := mux.NewRouter()

	if cfg.ModDir == "" {
		return nil, errors.New("invalid config: ModDir must be provided")
	}

	repo := repository.NewModuleRepository(cfg.ModDir)
	uc := usecase.NewProxyUsecase(repo)
	http.NewProxyHandler(r, uc)

	return r, nil
}

// NewTestServer is a convenience function to start a new httptest.Server to serve local modules during testing.
func NewTestServer(cfg *Config) (*httptest.Server, error) {
	h, err := NewHandler(cfg)
	if err != nil {
		return nil, err
	}

	return httptest.NewServer(h), nil
}
