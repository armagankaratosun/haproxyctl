package configuration

// mapGlobalFromAPI converts a generic API response for the "global" section
// into a GlobalConfig manifest structure.
func mapGlobalFromAPI(obj map[string]interface{}) GlobalConfig {
	cfg := GlobalConfig{
		APIVersion: "haproxyctl/v1",
		Kind:       "Global",
	}

	if v, ok := obj["daemon"].(bool); ok {
		cfg.Daemon = v
	}
	if v, ok := getInt(obj, "nbproc"); ok {
		cfg.Nbproc = v
	}
	if v, ok := getInt(obj, "maxconn"); ok {
		cfg.Maxconn = v
	}
	if v, ok := obj["log"].(string); ok {
		cfg.Log = v
	}
	if v, ok := obj["log_send_hostname"].(string); ok {
		cfg.LogSendHost = v
	}
	if v, ok := obj["stats_socket"].(string); ok {
		cfg.StatsSocket = v
	}
	if v, ok := obj["stats_timeout"].(string); ok {
		cfg.StatsTimeout = v
	}
	if v, ok := getInt(obj, "spread_checks"); ok {
		cfg.SpreadChecks = v
	}

	return cfg
}

// mapDefaultsFromAPI converts a generic API response for the "defaults"
// section into a DefaultsConfig manifest structure.
func mapDefaultsFromAPI(obj map[string]interface{}) DefaultsConfig {
	cfg := DefaultsConfig{
		APIVersion: "haproxyctl/v1",
		Kind:       "Defaults",
	}

	if v, ok := obj["name"].(string); ok {
		cfg.Name = v
	}

	if v, ok := obj["mode"].(string); ok {
		cfg.Mode = v
	}
	if v, ok := obj["timeout_client"].(string); ok {
		cfg.TimeoutClient = v
	}
	if v, ok := obj["timeout_server"].(string); ok {
		cfg.TimeoutServer = v
	}
	if v, ok := obj["timeout_connect"].(string); ok {
		cfg.TimeoutConnect = v
	}
	if v, ok := obj["timeout_queue"].(string); ok {
		cfg.TimeoutQueue = v
	}
	if v, ok := obj["timeout_tunnel"].(string); ok {
		cfg.TimeoutTunnel = v
	}
	if v, ok := obj["balance"].(string); ok {
		cfg.Balance = v
	}
	if v, ok := obj["log"].(string); ok {
		cfg.Log = v
	}

	return cfg
}

// getInt extracts an integer field from a map that may contain JSON numbers
// as float64 values.
func getInt(obj map[string]interface{}, key string) (int, bool) {
	v, ok := obj[key]
	if !ok || v == nil {
		return 0, false
	}

	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}

// GlobalConfig defines the subset of the HAProxy "global" section that
// haproxyctl exposes via manifests and edit/apply flows.
type GlobalConfig struct {
	APIVersion string `yaml:"apiVersion,omitempty" json:"-"`
	Kind       string `yaml:"kind,omitempty" json:"-"`

	Daemon  bool `yaml:"daemon,omitempty" json:"daemon,omitempty"`
	Nbproc  int  `yaml:"nbproc,omitempty" json:"nbproc,omitempty"`
	Maxconn int  `yaml:"maxconn,omitempty" json:"maxconn,omitempty"`

	// Log and stats settings
	Log          string `yaml:"log,omitempty" json:"log,omitempty"`
	LogSendHost  string `yaml:"logSendHost,omitempty" json:"log_send_hostname,omitempty"` //nolint:tagliatelle // Data Plane API field name
	StatsSocket  string `yaml:"statsSocket,omitempty" json:"stats_socket,omitempty"`      //nolint:tagliatelle // Data Plane API field name
	StatsTimeout string `yaml:"statsTimeout,omitempty" json:"stats_timeout,omitempty"`    //nolint:tagliatelle // Data Plane API field name

	// Misc tuning knobs
	SpreadChecks int `yaml:"spreadChecks,omitempty" json:"spread_checks,omitempty"` //nolint:tagliatelle // JSON field comes from Data Plane API
}

// DefaultsConfig represents a minimal, manifest-friendly view of the
// "defaults" section.
type DefaultsConfig struct {
	APIVersion string `yaml:"apiVersion,omitempty" json:"-"`
	Kind       string `yaml:"kind,omitempty" json:"-"`

	Name string `yaml:"name,omitempty" json:"name,omitempty"`

	Mode string `yaml:"mode,omitempty" json:"mode,omitempty"`

	TimeoutClient  string `yaml:"timeoutClient,omitempty" json:"timeout_client,omitempty"`   //nolint:tagliatelle // Data Plane API field name
	TimeoutServer  string `yaml:"timeoutServer,omitempty" json:"timeout_server,omitempty"`   //nolint:tagliatelle // Data Plane API field name
	TimeoutConnect string `yaml:"timeoutConnect,omitempty" json:"timeout_connect,omitempty"` //nolint:tagliatelle // Data Plane API field name
	TimeoutQueue   string `yaml:"timeoutQueue,omitempty" json:"timeout_queue,omitempty"`     //nolint:tagliatelle // Data Plane API field name
	TimeoutTunnel  string `yaml:"timeoutTunnel,omitempty" json:"timeout_tunnel,omitempty"`   //nolint:tagliatelle // Data Plane API field name

	Balance string `yaml:"balance,omitempty" json:"balance,omitempty"`
	Log     string `yaml:"log,omitempty" json:"log,omitempty"`
}

// isEmpty reports whether the GlobalConfig has no meaningful settings
// (i.e., all fields other than apiVersion/kind are zero values).
func (g GlobalConfig) isEmpty() bool {
	return !g.Daemon &&
		g.Nbproc == 0 &&
		g.Maxconn == 0 &&
		g.Log == "" &&
		g.LogSendHost == "" &&
		g.StatsSocket == "" &&
		g.StatsTimeout == "" &&
		g.SpreadChecks == 0
}

// isEmpty reports whether the DefaultsConfig has no meaningful settings.
func (d DefaultsConfig) isEmpty() bool {
	return d.Mode == "" &&
		d.TimeoutClient == "" &&
		d.TimeoutServer == "" &&
		d.TimeoutConnect == "" &&
		d.TimeoutQueue == "" &&
		d.TimeoutTunnel == "" &&
		d.Balance == "" &&
		d.Log == ""
}
