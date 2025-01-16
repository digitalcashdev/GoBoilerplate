package goboilerplate

import (
	"encoding/json"
	"net/http"
)

type GoBoilerplateConfig map[string]any

type RPCConfig struct {
	BaseURL  string
	Username string
	Password string
}

type Server struct {
	RPCConfig RPCConfig
	Config    GoBoilerplateConfig
}

func (srv *Server) GoBoilerplateHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	if len(key) > 0 {
		_ = encoder.Encode(srv.Config[key])
		return
	}

	_ = encoder.Encode(srv.Config)
}
