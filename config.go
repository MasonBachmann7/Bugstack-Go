package bugstack

// Config holds the configuration for the BugStack SDK.
type Config struct {
	// APIKey is your BugStack API key (required).
	APIKey string

	// Endpoint is the BugStack API endpoint.
	// Default: "https://api.bugstack.dev/api/capture"
	Endpoint string

	// ProjectID is an optional project identifier.
	ProjectID string

	// Environment is the deployment environment name.
	// Default: "production"
	Environment string

	// AutoFix enables AI-powered autonomous error fixing.
	AutoFix bool

	// Enabled is a kill switch. Set to false to disable all capture.
	// Default: true (when nil or unset)
	Enabled *bool

	// Debug enables verbose SDK logging to stderr.
	Debug bool

	// DryRun logs events to stderr instead of sending them.
	DryRun bool

	// DeduplicationWindow is how long (in seconds) to suppress
	// duplicate errors. Default: 300 (5 minutes).
	DeduplicationWindow float64

	// Timeout is the HTTP timeout in seconds. Default: 5.
	Timeout float64

	// MaxRetries is the max number of retry attempts. Default: 3.
	MaxRetries int

	// IgnoredErrors is a list of error message substrings to ignore.
	IgnoredErrors []string

	// BeforeSend is a hook that can inspect, modify, or drop events.
	// Return nil to drop the event.
	BeforeSend func(*Event) *Event
}

// Bool returns a pointer to a bool value. Convenience helper for Config.Enabled.
func Bool(v bool) *bool { return &v }

func (c *Config) setDefaults() {
	if c.Endpoint == "" {
		c.Endpoint = "https://api.bugstack.dev/api/capture"
	}
	if c.Environment == "" {
		c.Environment = "production"
	}
	if c.DeduplicationWindow == 0 {
		c.DeduplicationWindow = 300
	}
	if c.Timeout == 0 {
		c.Timeout = 5
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}
	// Enabled defaults to true when not explicitly set
	if c.Enabled == nil {
		t := true
		c.Enabled = &t
	}
}
