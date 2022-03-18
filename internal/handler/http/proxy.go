package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mercari/modserver/internal/domain"
	"github.com/mercari/modserver/internal/usecase"
)

// InfoResponse represents an info query response.
type InfoResponse struct {
	Version string `json:"Version"`
}

// NewProxyHandler initializes a new proxy handler on the given router.
func NewProxyHandler(r *mux.Router, uc usecase.ProxyUsecase) {
	h := proxyHandler{
		uc: uc,
	}
	r.HandleFunc("/{module:.+}/@v/list", h.GETList).Methods(http.MethodGet)
	r.HandleFunc("/{module:.+}/@v/{version:.+}.info", h.GETInfo).Methods(http.MethodGet)
	r.HandleFunc("/{module:.+}/@v/{version:.+}.mod", h.GETMod).Methods(http.MethodGet)
	r.HandleFunc("/{module:.+}/@v/{version:.+}.zip", h.GETZip).Methods(http.MethodGet)
}

type proxyHandler struct {
	uc usecase.ProxyUsecase
}

// GETList handles list queries.
func (h *proxyHandler) GETList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	path, _ := parsePathVersion(r)

	mod, err := h.uc.LoadModule(ctx, path)
	if err != nil {
		respondError(w, statusCode(err), err)
		return
	}

	for _, ver := range mod.Versions {
		fmt.Fprintln(w, ver)
	}
}

// GETInfo handles info queries.
func (h *proxyHandler) GETInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	path, ver := parsePathVersion(r)

	mod, err := h.uc.LoadModule(ctx, path)
	if err != nil {
		respondError(w, statusCode(err), err)
		return
	}
	if !mod.HasVersion(ver) {
		respondError(w, statusCode(domain.ErrNotFound), fmt.Errorf("%w: \"%s\": unknown revision %s", domain.ErrNotFound, path, ver))
		return
	}

	i, err := json.Marshal(InfoResponse{Version: ver})
	if err != nil {
		respondError(w, statusCode(err), err)
		return
	}

	fmt.Fprintln(w, string(i))
}

// GETMod handles mod file queries.
func (h *proxyHandler) GETMod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	path, ver := parsePathVersion(r)

	mod, err := h.uc.LoadModule(ctx, path)
	if err != nil {
		respondError(w, statusCode(err), err)
		return
	}
	if !mod.HasVersion(ver) {
		respondError(w, statusCode(domain.ErrNotFound), fmt.Errorf("%w: \"%s\": unknown revision %s", domain.ErrNotFound, path, ver))
		return
	}

	f, err := h.uc.GetModFile(ctx, mod, ver)
	if err != nil {
		respondError(w, statusCode(err), err)
		return
	}

	fmt.Fprintln(w, f)
}

// GETZip handles module zip file queries.
func (h *proxyHandler) GETZip(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	path, ver := parsePathVersion(r)

	mod, err := h.uc.LoadModule(ctx, path)
	if err != nil {
		respondError(w, statusCode(err), err)
		return
	}
	if !mod.HasVersion(ver) {
		respondError(w, statusCode(domain.ErrNotFound), fmt.Errorf("%w: \"%s\": unknown revision %s", domain.ErrNotFound, path, ver))
		return
	}

	var buf bytes.Buffer
	err = h.uc.ZipModule(ctx, mod, ver, &buf)
	if err != nil {
		respondError(w, statusCode(err), err)
		return
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		respondError(w, statusCode(err), err)
		return
	}
}

func parsePathVersion(r *http.Request) (string, string) {
	return mux.Vars(r)["module"], mux.Vars(r)["version"]
}

func respondError(w http.ResponseWriter, code int, errs ...interface{}) {
	w.WriteHeader(code)
	fmt.Fprintln(w, errs...)
}

func statusCode(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusGone
	}
	return http.StatusInternalServerError
}
