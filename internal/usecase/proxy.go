package usecase

import (
	"archive/zip"
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mercari/modserver/internal/domain"
)

// ProxyUsecase represents the usecases of a proxy.
type ProxyUsecase interface {
	// LoadModule loads a module by the given module path.
	LoadModule(ctx context.Context, path string) (*domain.Module, error)

	// GetModFile returns a module's go.mod file.
	GetModFile(ctx context.Context, mod *domain.Module, ver string) (string, error)

	// ZipModule writes a zip file of the module to the given writer.
	ZipModule(ctx context.Context, mod *domain.Module, ver string, dst io.Writer) error
}

// NewProxyUsecase initializes a new ProxyUsecase.
func NewProxyUsecase(moduleRepo domain.ModuleRepository) ProxyUsecase {
	return &proxyUsecase{
		moduleRepo: moduleRepo,
	}
}

type proxyUsecase struct {
	moduleRepo domain.ModuleRepository
}

var _ ProxyUsecase = (*proxyUsecase)(nil)

func (u *proxyUsecase) LoadModule(ctx context.Context, path string) (*domain.Module, error) {
	mod, err := u.moduleRepo.LoadByPath(ctx, path)
	if err != nil {
		return nil, err
	}

	return mod, nil
}

func (u *proxyUsecase) GetModFile(ctx context.Context, mod *domain.Module, ver string) (string, error) {
	b, err := u.moduleRepo.LoadModFile(ctx, mod, ver)
	if err != nil {
		return "", err
	}

	return string(b), err
}

func (u *proxyUsecase) ZipModule(ctx context.Context, mod *domain.Module, ver string, dst io.Writer) error {
	baseDir, root := u.moduleRepo.GetPaths(mod, ver)

	zw := zip.NewWriter(dst)
	defer zw.Close()

	err := filepath.WalkDir(root, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		fi, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fi.Close()

		rel, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		fo, err := zw.Create(rel)
		if err != nil {
			return err
		}

		_, err = io.Copy(fo, fi)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
