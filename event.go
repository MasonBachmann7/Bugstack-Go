package bugstack

import "time"

// Event represents an error event to be sent to BugStack.
type Event struct {
	Message       string          `json:"message"`
	StackTrace    string          `json:"stackTrace"`
	File          string          `json:"file"`
	Function      string          `json:"function"`
	Fingerprint   string          `json:"fingerprint"`
	ExceptionType string          `json:"exceptionType"`
	Request       *RequestContext `json:"request,omitempty"`
	Environment   EnvironmentInfo `json:"environment"`
	Timestamp     time.Time       `json:"timestamp"`
	Metadata      map[string]any  `json:"metadata,omitempty"`
}

// RequestContext holds HTTP request information.
type RequestContext struct {
	Route       string            `json:"route"`
	Method      string            `json:"method"`
	QueryParams map[string]string `json:"queryParams,omitempty"`
}

// EnvironmentInfo holds runtime environment details.
type EnvironmentInfo struct {
	Language         string `json:"language"`
	LanguageVersion  string `json:"languageVersion"`
	Framework        string `json:"framework,omitempty"`
	FrameworkVersion string `json:"frameworkVersion,omitempty"`
	OS               string `json:"os"`
	SDKVersion       string `json:"sdkVersion"`
}

// CaptureOption is a functional option for CaptureError/CaptureMessage.
type CaptureOption func(*captureOptions)

type captureOptions struct {
	request  *RequestContext
	metadata map[string]any
}

// WithRequest attaches HTTP request context to the error event.
func WithRequest(req *RequestContext) CaptureOption {
	return func(o *captureOptions) {
		o.request = req
	}
}

// WithMetadata attaches arbitrary key-value metadata to the event.
func WithMetadata(meta map[string]any) CaptureOption {
	return func(o *captureOptions) {
		o.metadata = meta
	}
}

// toPayload serializes the event to the BugStack API payload format.
func (e *Event) toPayload(cfg Config) map[string]any {
	payload := map[string]any{
		"apiKey": cfg.APIKey,
		"error": map[string]any{
			"message":    e.Message,
			"stackTrace": e.StackTrace,
			"file":       e.File,
			"function":   e.Function,
			"fingerprint": e.Fingerprint,
		},
		"environment": map[string]any{
			"language":        e.Environment.Language,
			"languageVersion": e.Environment.LanguageVersion,
			"framework":       e.Environment.Framework,
			"frameworkVersion": e.Environment.FrameworkVersion,
			"os":              e.Environment.OS,
			"sdkVersion":      e.Environment.SDKVersion,
		},
		"timestamp": e.Timestamp.UTC().Format(time.RFC3339),
	}

	if e.Request != nil {
		payload["request"] = map[string]any{
			"route":  e.Request.Route,
			"method": e.Request.Method,
		}
	}

	if cfg.ProjectID != "" {
		payload["projectId"] = cfg.ProjectID
	}

	meta := e.Metadata
	if cfg.AutoFix {
		if meta == nil {
			meta = make(map[string]any)
		}
		meta["autoFix"] = true
	}
	if len(meta) > 0 {
		payload["metadata"] = meta
	}

	return payload
}
