package domain

import (
	"context"
)

// Module represents a Go module
type Module struct {
	// Path is the module path.
	Path string

	// Versions are the list of versions the module has.
	Versions []string
}

// HasVersions returns true if the module has the given version.
func (m *Module) HasVersion(ver string) bool {
	if m == nil {
		return false
	}

	for _, v := range m.Versions {
		if v == ver {
			return true
		}
	}

	return false
}

// ModuleRepository represents a module repository.
type ModuleRepository interface {
	// LoadByPath loads a module by the given module path.
	LoadByPath(ctx context.Context, path string) (*Module, error)

	// LoadModFile returns the module's go.mod file.
	LoadModFile(ctx context.Context, mod *Module, ver string) ([]byte, error)

	// GetPaths returns the base directory of the module collection and the path to the module's directory.
	GetPaths(mod *Module, ver string) (string, string)
}
