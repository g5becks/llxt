package registry

import (
	"sort"
	"strings"

	errs "github.com/g5becks/llxt/internal/errors"
)

// Lookup finds an entry by key (case-insensitive).
func Lookup(key string) (*Entry, error) {
	k := strings.ToLower(key)
	if entry, ok := entries[k]; ok {
		return entry, nil
	}
	return nil, errs.RegistryErr.
		Code(errs.CodeNotFound).
		With("key", key).
		Hint("Use 'llxt list' to see available sources or 'llxt add' to add a new one").
		Errorf("source %q not found in registry", key)
}

// List returns all registry entries sorted by key.
func List() []*Entry {
	result := make([]*Entry, 0, len(entries))
	for _, e := range entries {
		result = append(result, e)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})
	return result
}

// ListByCategory returns entries filtered by category.
func ListByCategory(category string) []*Entry {
	result := make([]*Entry, 0)
	cat := strings.ToLower(category)
	for _, e := range entries {
		if strings.ToLower(e.Category) == cat {
			result = append(result, e)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})
	return result
}

// Count returns the total number of entries.
func Count() int {
	return len(entries)
}

// Keys returns all registry keys sorted.
func Keys() []string {
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
