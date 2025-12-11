package backends

import (
	"fmt"
	"haproxyctl/cmd/servers"
	"haproxyctl/internal"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v2"
)

// ApplyBackendFromYAML applies a backend manifest in a declarative way:
//   - If the backend does not exist, it is created along with its servers.
//   - If it exists, it is replaced via PUT and servers are reconciled using
//     the same diff logic as the interactive edit flow.
func ApplyBackendFromYAML(data []byte, outputFormat string, dryRun bool) error {
	var manifest backendWithServers
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse backend manifest: %w", err)
	}

	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("invalid backend configuration: %w", err)
	}

	// Preview/dry-run behaviour mirrors `create backends`.
	if outputFormat != "" || dryRun {
		if outputFormat == "" {
			outputFormat = internal.OutputFormatYAML
			internal.FormatOutput(manifest, outputFormat)
		} else {
			internal.FormatOutput(manifest.toPayload(), outputFormat)
		}
		if dryRun {
			internal.PrintDryRun()
		}
		return nil
	}

	name := manifest.Name

	// Determine whether the backend exists and, if it does, capture its
	// current configuration so we can detect no-op applies.
	path := "/services/haproxy/configuration/backends/" + name
	rawBackend, err := internal.GetResource(path)
	notFound := false
	if err != nil {
		if internal.IsNotFoundError(err) {
			notFound = true
			rawBackend = nil
		} else {
			return fmt.Errorf("failed to check backend existence: %w", err)
		}
	}

	version, err := internal.GetConfigurationVersion()
	if err != nil {
		return fmt.Errorf("failed to fetch HAProxy configuration version: %w", err)
	}

	payload := manifest.toPayload()

	if notFound {
		// Create backend, then create servers to match manifest.
		if _, err := internal.SendRequest(
			"POST",
			"/services/haproxy/configuration/backends",
			map[string]string{"version": strconv.Itoa(version)},
			payload,
		); err != nil {
			return fmt.Errorf("failed to create backend %q: %w", name, err)
		}

		for _, srv := range manifest.Servers {
			srv.Backend = name
			if err := servers.CreateServer(srv, "", false); err != nil {
				return fmt.Errorf("failed to create server %q for backend %q: %w", srv.Name, name, err)
			}
		}

		internal.PrintStatus("Backend", name, internal.ActionCreated)
		return nil
	}

	// Fetch current servers to compute the diff and to detect no-op applies.
	rawServers, err := internal.GetResourceList(
		"/services/haproxy/configuration/backends/" + name + "/servers",
	)
	if err != nil && !internal.IsNotFoundError(err) {
		return fmt.Errorf("failed to fetch existing servers for backend %q: %w", name, err)
	}

	var before []servers.ServerConfig
	for _, srv := range rawServers {
		sc := mapServerFromAPI(name, srv)
		if sc.Name != "" && sc.Address != "" && sc.Port != 0 {
			before = append(before, sc)
		}
	}

	// Build a normalized view of the current backend (config + servers) to
	// detect if this apply would be a no-op.
	var current backendWithServers
	if rawBackend != nil {
		populateBackendConfigFromMap(&current.backendConfig, rawBackend)
		current.Servers = before
	}

	if reflect.DeepEqual(current.backendConfig, manifest.backendConfig) &&
		serversEqualByName(before, manifest.Servers) {
		internal.PrintStatus("Backend", name, internal.ActionUnchanged)
		return nil
	}

	// Update existing backend via PUT, then reconcile servers using the
	// same diff logic as the interactive edit flow.
	if _, err := internal.SendRequest(
		"PUT",
		path,
		map[string]string{"version": strconv.Itoa(version)},
		payload,
	); err != nil {
		return fmt.Errorf("failed to update backend %q: %w", name, err)
	}

	if err := applyServerDiff(name, before, manifest.Servers); err != nil {
		return fmt.Errorf("failed to apply server changes for backend %q: %w", name, err)
	}

	internal.PrintStatus("Backend", name, internal.ActionConfigured)
	return nil
}

// serversEqualByName compares two slices of ServerConfig using the same
// semantics as applyServerDiff: identity by name, and equality via
// serverConfigEqual.
func serversEqualByName(a, b []servers.ServerConfig) bool {
	if len(a) != len(b) {
		return false
	}

	aByName := make(map[string]servers.ServerConfig, len(a))
	for _, s := range a {
		aByName[s.Name] = s
	}

	for _, s := range b {
		existing, ok := aByName[s.Name]
		if !ok {
			return false
		}
		if !serverConfigEqual(existing, s) {
			return false
		}
	}

	return true
}
