// Package registry provides access to the bundled llms.txt registry.
package registry

// Entry represents a unified registry entry.
type Entry struct {
	Key         string  `json:"key"`
	Name        string  `json:"name"`
	Domain      string  `json:"domain"`
	Description string  `json:"description,omitempty"`
	Category    string  `json:"category,omitempty"`
	LLMsURL     string  `json:"llms_url"`
	LLMsFullURL *string `json:"llms_full_url,omitempty"`
}
