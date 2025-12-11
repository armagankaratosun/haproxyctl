package frontends

import (
	"fmt"
	"haproxyctl/internal"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v2"
)

// ApplyFrontendFromYAML applies a frontend manifest declaratively:
//   - If the frontend does not exist, it is created along with its binds.
//   - If it exists, it is replaced via PUT and binds are reconciled using
//     the same diff logic as the interactive edit flow.
func ApplyFrontendFromYAML(data []byte, outputFormat string, dryRun bool) error {
	var manifest frontendWithBinds
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse frontend manifest: %w", err)
	}

	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("invalid frontend configuration: %w", err)
	}

	// Preview/dry-run behaviour mirrors `create frontends`.
	if outputFormat != "" || dryRun {
		if outputFormat == "" {
			outputFormat = internal.OutputFormatYAML
			internal.FormatOutput(manifest, outputFormat)
		} else {
			internal.FormatOutput(manifest.ToPayload(), outputFormat)
		}
		if dryRun {
			internal.PrintDryRun()
		}
		return nil
	}

	name := manifest.Name

	// Determine whether the frontend exists and, if it does, capture its
	// current configuration so we can detect no-op applies.
	path := "/services/haproxy/configuration/frontends/" + name
	rawFrontend, err := internal.GetResource(path)
	notFound := false
	if err != nil {
		if internal.IsNotFoundError(err) {
			notFound = true
			rawFrontend = nil
		} else {
			return fmt.Errorf("failed to check frontend existence: %w", err)
		}
	}

	version, err := internal.GetConfigurationVersion()
	if err != nil {
		return fmt.Errorf("failed to fetch HAProxy configuration version: %w", err)
	}

	payload := manifest.ToPayload()

	if notFound {
		// Create frontend, then create binds to match manifest.
		if _, err := internal.SendRequest(
			"POST",
			"/services/haproxy/configuration/frontends",
			map[string]string{"version": strconv.Itoa(version)},
			payload,
		); err != nil {
			return fmt.Errorf("failed to create frontend %q: %w", name, err)
		}

		for _, b := range manifest.Binds {
			if err := createBind(name, b); err != nil {
				return fmt.Errorf("failed to create bind on frontend %q: %w", name, err)
			}
		}

		internal.PrintStatus("Frontend", name, internal.ActionCreated)
		return nil
	}

	rawBinds, err := internal.GetResourceList(
		"/services/haproxy/configuration/frontends/" + name + "/binds",
	)
	if err != nil && !internal.IsNotFoundError(err) {
		return fmt.Errorf("failed to fetch existing binds for frontend %q: %w", name, err)
	}

	var before []BindConfig
	for _, raw := range rawBinds {
		bc := mapBindFromAPI(raw)
		if bc.Address != "" && bc.Port != 0 {
			before = append(before, bc)
		}
	}

	// Build a normalized view of the current frontend (config + binds) to
	// detect if this apply would be a no-op.
	var current frontendWithBinds
	if rawFrontend != nil {
		populateFrontendConfigFromMap(&current.frontendConfig, rawFrontend)
		current.Binds = before
	}

	if reflect.DeepEqual(current.frontendConfig, manifest.frontendConfig) &&
		bindsEqualByKey(before, manifest.Binds) {
		internal.PrintStatus("Frontend", name, internal.ActionUnchanged)
		return nil
	}

	// Update existing frontend via PUT, then reconcile binds using the
	// same diff logic as the interactive edit flow.
	if _, err := internal.SendRequest(
		"PUT",
		path,
		map[string]string{"version": strconv.Itoa(version)},
		payload,
	); err != nil {
		return fmt.Errorf("failed to update frontend %q: %w", name, err)
	}

	if err := applyBindDiff(name, before, manifest.Binds); err != nil {
		return fmt.Errorf("failed to apply bind changes for frontend %q: %w", name, err)
	}

	internal.PrintStatus("Frontend", name, internal.ActionConfigured)
	return nil
}

// bindsEqualByKey compares two slices of BindConfig using the same
// semantics as applyBindDiff: identity by address:port, and equality via
// bindConfigEqual.
func bindsEqualByKey(a, b []BindConfig) bool {
	if len(a) != len(b) {
		return false
	}

	aByKey := make(map[string]BindConfig, len(a))
	for _, bind := range a {
		key := fmt.Sprintf("%s:%d", bind.Address, bind.Port)
		aByKey[key] = bind
	}

	for _, bind := range b {
		key := fmt.Sprintf("%s:%d", bind.Address, bind.Port)
		existing, ok := aByKey[key]
		if !ok {
			return false
		}
		if !bindConfigEqual(existing, bind) {
			return false
		}
	}

	return true
}
