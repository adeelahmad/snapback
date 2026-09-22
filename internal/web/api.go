package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/adeelahmad/snapback/internal/config"
	"github.com/adeelahmad/snapback/internal/errcode"
)

// SetupValidator checks a candidate configuration against its repository.
type SetupValidator interface {
	Validate(ctx context.Context, c *config.Config) error
}

type configPayload struct {
	Revision config.Revision `json:"revision"`
	Config   *config.Config  `json:"config"`
}

type validateResult struct {
	OK    bool   `json:"ok"`
	Code  string `json:"code,omitempty"`
	Error string `json:"error,omitempty"`
}

// writeJSON writes v as an uncached JSON response with status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes the {code, error} envelope.
func writeError(w http.ResponseWriter, status int, code errcode.Code, err error) {
	writeJSON(w, status, map[string]string{"code": string(code), "error": err.Error()})
}

func (s *Server) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.opts.Backend.Status())
}

func (s *Server) handleAPIConfigGet(w http.ResponseWriter, r *http.Request) {
	c, rev, err := s.opts.Backend.Config()
	if err != nil {
		writeError(w, http.StatusInternalServerError, errcode.Of(err), err)
		return
	}
	writeJSON(w, http.StatusOK, configPayload{Revision: rev, Config: c})
}

func (s *Server) handleAPIConfigPut(w http.ResponseWriter, r *http.Request) {
	var p configPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, errcode.InvalidConfig, err)
		return
	}
	if p.Config == nil {
		writeError(w, http.StatusBadRequest, errcode.InvalidConfig, errors.New("missing config"))
		return
	}
	rev, err := s.opts.Backend.SaveConfig(p.Config, p.Revision)
	if errors.Is(err, config.ErrRevisionConflict) {
		writeError(w, http.StatusConflict, errcode.StaleState, err)
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, errcode.InvalidConfig, err)
		return
	}
	writeJSON(w, http.StatusOK, configPayload{Revision: rev, Config: p.Config})
}

func (s *Server) handleAPISetupValidate(w http.ResponseWriter, r *http.Request) {
	if s.opts.Validator == nil {
		writeError(w, http.StatusServiceUnavailable, errcode.PrereqMissing, errors.New("no setup validator"))
		return
	}
	var p configPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, errcode.InvalidConfig, err)
		return
	}
	if p.Config == nil {
		writeError(w, http.StatusBadRequest, errcode.InvalidConfig, errors.New("missing config"))
		return
	}
	if err := s.opts.Validator.Validate(r.Context(), p.Config); err != nil {
		writeJSON(w, http.StatusOK, validateResult{Code: string(errcode.Of(err)), Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, validateResult{OK: true})
}

func (s *Server) handleAPIIntegrations(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{})
}
