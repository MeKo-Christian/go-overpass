package turbo

import "strings"

// SQLDataConfig describes a Postpass-style SQL backend configuration.
type SQLDataConfig struct {
	Server string
	Params map[string]string
}

// SQLDataConfigFromResult builds a SQLDataConfig from Result.Data when mode is "sql".
// Returns nil if no sql data source is present.
func SQLDataConfigFromResult(res Result) *SQLDataConfig {
	if res.Data == nil {
		return nil
	}
	if !strings.EqualFold(res.Data.Mode, "sql") {
		return nil
	}

	cfg := &SQLDataConfig{
		Server: res.DataServer,
		Params: map[string]string{},
	}
	for k, v := range res.Data.Options {
		if k == "server" {
			continue
		}
		cfg.Params[k] = v
	}
	return cfg
}
