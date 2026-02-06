package registry

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed directory_entries.json
var directoryEntriesJSON []byte

//go:embed websites.json
var websitesJSON []byte

// DirectoryEntry matches directory_entries.json format.
type DirectoryEntry struct {
	Key         string  `json:"key"`
	Name        string  `json:"name"`
	Domain      string  `json:"domain"`
	LLMsURL     string  `json:"llms_url"`
	LLMsFullURL *string `json:"llms_full_url"`
}

// WebsiteEntry matches websites.json format.
type WebsiteEntry struct {
	Name           string  `json:"name"`
	Domain         string  `json:"domain"`
	Description    string  `json:"description"`
	LLMsTxtURL     string  `json:"llmsTxtUrl"`
	LLMsFullTxtURL *string `json:"llmsFullTxtUrl,omitempty"`
	Category       string  `json:"category"`
}

//nolint:gochecknoglobals // registry entries are loaded once at init
var entries map[string]*Entry

//nolint:gochecknoinits // init required to load embedded data
func init() {
	entries = make(map[string]*Entry)
	loadDirectoryEntries()
	loadWebsiteEntries()
}

func loadDirectoryEntries() {
	var dirEntries []DirectoryEntry
	_ = json.Unmarshal(directoryEntriesJSON, &dirEntries)
	for _, e := range dirEntries {
		entries[e.Key] = &Entry{
			Key:         e.Key,
			Name:        e.Name,
			Domain:      e.Domain,
			LLMsURL:     e.LLMsURL,
			LLMsFullURL: e.LLMsFullURL,
		}
	}
}

func loadWebsiteEntries() {
	var webEntries []WebsiteEntry
	_ = json.Unmarshal(websitesJSON, &webEntries)
	for _, e := range webEntries {
		key := slugify(e.Name)
		if existing, ok := entries[key]; ok {
			existing.Description = e.Description
			existing.Category = e.Category
		} else {
			entries[key] = &Entry{
				Key:         key,
				Name:        e.Name,
				Domain:      e.Domain,
				Description: e.Description,
				Category:    e.Category,
				LLMsURL:     e.LLMsTxtURL,
				LLMsFullURL: e.LLMsFullTxtURL,
			}
		}
	}
}

func slugify(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
