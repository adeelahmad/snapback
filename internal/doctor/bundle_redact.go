package doctor

import (
	"github.com/adeelahmad/snapback/internal/config"
)

// redactForBundle prepares the config, doctor report and daemon log for a
// diagnostic bundle.
func redactForBundle(cfg *config.Config, doctorJSON, daemonLog []byte) ([]byte, []byte, []byte, error) {
	configYAML, err := config.Marshal(cfg)
	if err != nil {
		return nil, nil, nil, err
	}
	return configYAML, doctorJSON, daemonLog, nil
}
