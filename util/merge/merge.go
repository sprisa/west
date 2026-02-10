package merge

import (
	"os"
	l "reship/util/logger/v2"

	"dario.cat/mergo"
	"github.com/goccy/go-yaml"
	"github.com/sprisa/west/config"
	"github.com/sprisa/x/env"
	"github.com/sprisa/x/errutil"
)

func MaybeMergeCustomNebulaCfg(cfg *config.Config) error {
	// Merge in custom Nebula config if exists
	customCfgPath := env.WithDefault("WEST_NEBULA_CONFIG", "")
	if customCfgPath == "" {
		return nil
	}

	data, err := os.ReadFile(customCfgPath)
	if err != nil {
		return errutil.WrapErr(err, "error reading custom nebula config: %s", customCfgPath)
	}
	l.Log.Debug().Msg("Custom Nebula Config")
	l.Log.Debug().Msgf("\n%s", string(data))
	var envConfig config.Config
	err = yaml.Unmarshal(data, &envConfig)
	if err != nil {
		return errutil.WrapErr(err, "error parsing custom nebula config: %s", customCfgPath)
	}
	err = mergo.Merge(cfg, envConfig, mergo.WithOverride)
	if err != nil {
		return errutil.WrapErr(err, "error merging custom nebula config: %s", customCfgPath)
	}
	redactedCfg := *cfg
	if redactedCfg.Pki.Key != "" {
		redactedCfg.Pki.Key = "<redacted>"
	}
	yamlStr, _ := yaml.Marshal(&redactedCfg)
	l.Log.Debug().Msg("Combined Nebula Config")
	l.Log.Debug().Msgf("\n%s", string(yamlStr))
	return nil
}
