package web

import (
	"context"
	"net/http"

	"github.com/adeelahmad/snapback/internal/config"
)

// SetupValidator checks a candidate configuration against its repository.
type SetupValidator interface {
	Validate(ctx context.Context, c *config.Config) error
}

func (s *Server) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	panic("SUB-AGENT-TODO: write Backend.Status() as JSON with Cache-Control no-store")
}

func (s *Server) handleAPIConfigGet(w http.ResponseWriter, r *http.Request) {
	panic("SUB-AGENT-TODO: write {revision, config} from Backend.Config() as JSON; errors via the {code, error} envelope")
}

func (s *Server) handleAPIConfigPut(w http.ResponseWriter, r *http.Request) {
	panic("SUB-AGENT-TODO: decode {revision, config} (bad JSON -> 400 invalid_config, no save), call Backend.SaveConfig, return {revision, config}; config.ErrRevisionConflict -> 409 stale_state")
}

func (s *Server) handleAPISetupValidate(w http.ResponseWriter, r *http.Request) {
	panic("SUB-AGENT-TODO: nil Options.Validator -> 503 prereq_missing; decode {config}, run Validator.Validate, reply 200 {ok, code} with the errcode of any error")
}

func (s *Server) handleAPIIntegrations(w http.ResponseWriter, r *http.Request) {
	panic("SUB-AGENT-TODO: write the integrations summary as a JSON object")
}
