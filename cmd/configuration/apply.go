package configuration

import (
	"fmt"
	"haproxyctl/internal"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
)

func applyConfig[T any](
	data []byte,
	outputFormat string,
	dryRun bool,
	kind string,
	getCurrent func() (T, error),
	putFn func(int, T) error,
) error {
	var manifest T
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse %s manifest: %w", strings.ToLower(kind), err)
	}

	if outputFormat != "" || dryRun {
		if outputFormat == "" {
			outputFormat = internal.OutputFormatYAML
			internal.FormatOutput(manifest, outputFormat)
		} else {
			internal.FormatOutput(manifest, outputFormat)
		}
		if dryRun {
			internal.PrintDryRun()
		}
		return nil
	}

	current, err := getCurrent()
	if err != nil {
		return err
	}

	if reflect.DeepEqual(current, manifest) {
		internal.PrintStatus(kind, "config", internal.ActionUnchanged)
		return nil
	}

	version, err := internal.GetConfigurationVersion()
	if err != nil {
		return fmt.Errorf("failed to fetch HAProxy configuration version: %w", err)
	}

	if err := putFn(version, manifest); err != nil {
		return err
	}

	internal.PrintStatus(kind, "config", internal.ActionConfigured)
	return nil
}

// ApplyGlobalFromYAML applies a GlobalConfig manifest declaratively.
func ApplyGlobalFromYAML(data []byte, outputFormat string, dryRun bool) error {
	return applyConfig(
		data,
		outputFormat,
		dryRun,
		"Global",
		func() (GlobalConfig, error) {
			obj, err := internal.GetResource("/services/haproxy/configuration/global")
			if err != nil && !internal.IsNotFoundError(err) {
				return GlobalConfig{}, fmt.Errorf("failed to fetch current global configuration: %w", err)
			}
			if obj == nil {
				return GlobalConfig{}, nil
			}
			return mapGlobalFromAPI(obj), nil
		},
		putGlobal,
	)
}

// ApplyDefaultsFromYAML applies a DefaultsConfig manifest declaratively.
func ApplyDefaultsFromYAML(data []byte, outputFormat string, dryRun bool) error {
	var currentName string

	return applyConfig(
		data,
		outputFormat,
		dryRun,
		"Defaults",
		func() (DefaultsConfig, error) {
			list, err := internal.GetResourceList("/services/haproxy/configuration/defaults")
			if err != nil && !internal.IsNotFoundError(err) {
				return DefaultsConfig{}, fmt.Errorf("failed to fetch current defaults configuration: %w", err)
			}
			if len(list) == 0 {
				currentName = ""
				return DefaultsConfig{}, nil
			}
			cfg := mapDefaultsFromAPI(list[0])
			currentName = cfg.Name
			return cfg, nil
		},
		func(version int, cfg DefaultsConfig) error {
			// If the manifest did not specify a name, fall back to the
			// current defaults section name (best-effort primary defaults).
			if cfg.Name == "" {
				cfg.Name = currentName
			}
			return putDefaults(version, cfg)
		},
	)
}
