package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/mercari/modserver/internal/domain"
)

// NewModuleRepository initializes a new ModuleRepository.
func NewModuleRepository(baseDir string) domain.ModuleRepository {
	return &moduleRepository{
		baseDir: filepath.Clean(baseDir),
	}
}

type moduleRepository struct {
	baseDir string
}

var _ domain.ModuleRepository = (*moduleRepository)(nil)

func (r *moduleRepository) LoadByPath(ctx context.Context, path string) (*domain.Module, error) {
	m := &domain.Module{
		Path: path,
	}

	modName := filepath.Base(path)
	modParent := filepath.Dir(path)
	if modParent == "." {
		return nil, fmt.Errorf("%w: invalid import path \"%s\"", domain.ErrNotFound, path)
	}

	modParent = filepath.Join(r.baseDir, modParent)
	_, err := os.Stat(modParent)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w: module \"%s\" not found", domain.ErrNotFound, path)
	}
	if err != nil {
		return nil, err
	}

	dirs, err := os.ReadDir(modParent)
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		s := strings.SplitN(dir.Name(), "@", 2)
		if len(s) != 2 {
			continue
		}

		mod, ver := s[0], s[1]
		if dir.IsDir() && mod == modName && validVersion(ver) {
			m.Versions = append(m.Versions, ver)
		}

	}

	if len(m.Versions) == 0 {
		return nil, fmt.Errorf("%w: module \"%s\" does not have any valid versions", domain.ErrNotFound, path)
	}

	return m, nil
}

func (r *moduleRepository) LoadModFile(ctx context.Context, mod *domain.Module, ver string) ([]byte, error) {
	b, err := os.ReadFile(filepath.Join(r.baseDir, fmt.Sprintf("%s@%s", mod.Path, ver), "go.mod"))
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (r *moduleRepository) GetPaths(mod *domain.Module, ver string) (string, string) {
	rel := fmt.Sprintf("%s@%s", mod.Path, ver)
	return r.baseDir, filepath.Join(r.baseDir, rel)
}

func validVersion(ver string) bool {
	build := semver.Build(ver)
	return semver.IsValid(ver) && (build == "" || build == "+incompatible")
}
