package modserver_test

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	nethttp "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mercari/modserver"
)

const modDir = "testdata/modules"

var update bool

func init() {
	flag.BoolVar(&update, "update", false, "update bin files")
}
func TestProxyServer(t *testing.T) {
	t.Parallel()

	type mode int
	const (
		modeList mode = iota
		modeInfo
		modeMod
		modeZip
	)

	cases := map[string]struct {
		mode    mode
		module  string
		version string

		wantStatusCode int
	}{
		"list-ok":        {modeList, "github.com/mercari/example", "", nethttp.StatusOK},
		"list-nomodule":  {modeList, "github.com/mercari/badexample", "", nethttp.StatusGone},
		"list-badmodule": {modeList, "github", "", nethttp.StatusGone},

		"info-ok":         {modeInfo, "github.com/mercari/example", "v0.2.0", nethttp.StatusOK},
		"info-nomodule":   {modeInfo, "github.com/mercari/badexample", "v0.1.0", nethttp.StatusGone},
		"info-noversion":  {modeInfo, "github.com/mercari/example", "v0.7899.0", nethttp.StatusGone},
		"info-badmodule":  {modeInfo, "github", "v0.2.0", nethttp.StatusGone},
		"info-badversion": {modeInfo, "github.com/mercari/example", "badver", nethttp.StatusGone},

		"mod-ok":         {modeMod, "github.com/mercari/example", "v0.2.0", nethttp.StatusOK},
		"mod-nomodule":   {modeMod, "github.com/mercari/badexample", "v0.1.0", nethttp.StatusGone},
		"mod-noversion":  {modeMod, "github.com/mercari/example", "v0.7899.0", nethttp.StatusGone},
		"mod-badmodule":  {modeMod, "github", "v0.2.0", nethttp.StatusGone},
		"mod-badversion": {modeMod, "github.com/mercari/example", "badver", nethttp.StatusGone},

		"zip-ok":         {modeZip, "github.com/mercari/example", "v0.2.0", nethttp.StatusOK},
		"zip-nomodule":   {modeZip, "github.com/mercari/badexample", "v0.1.0", nethttp.StatusGone},
		"zip-noversion":  {modeZip, "github.com/mercari/example", "v0.7899.0", nethttp.StatusGone},
		"zip-badmodule":  {modeZip, "github", "v0.2.0", nethttp.StatusGone},
		"zip-badversion": {modeZip, "github.com/mercari/example", "badver", nethttp.StatusGone},
	}

	url := func(srv *httptest.Server, module, version string, mode mode) string {
		switch mode {
		case modeList:
			return fmt.Sprintf("%s/%s/@v/list", srv.URL, module)
		case modeInfo:
			return fmt.Sprintf("%s/%s/@v/%s.info", srv.URL, module, version)
		case modeMod:
			return fmt.Sprintf("%s/%s/@v/%s.mod", srv.URL, module, version)
		case modeZip:
			return fmt.Sprintf("%s/%s/@v/%s.zip", srv.URL, module, version)
		}
		return ""
	}

	if update {
		RemoveAllBody(t, "testdata")
	}

	for name, tt := range cases {
		name, tt := name, tt
		t.Run(name, func(t *testing.T) {
			srv, err := modserver.NewTestServer(&modserver.Config{
				ModDir: modDir,
			})
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			defer srv.Close()

			resp, err := srv.Client().Get(url(srv, tt.module, tt.version, tt.mode))
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if resp.StatusCode != tt.wantStatusCode {
				t.Fatal("unexpected status code:", resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}

			if update {
				UpdateBody(t, "testdata", name, body)
			}

			diff := DiffBody(t, "testdata", name, body)
			if diff != "" {
				t.Fatal("response body mismatch:", diff)
			}
		})
	}

}

func RemoveAllBody(t *testing.T, testdata string) {
	t.Helper()

	err := filepath.WalkDir(testdata, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(path) != ".bin" {
			return nil
		}

		if err := os.Remove(path); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		t.Fatal("unexpected error", err)
	}
}

func UpdateBody(t *testing.T, testdata, name string, body []byte) {
	t.Helper()

	fo, err := os.Create(filepath.Join(testdata, name+".bin"))
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	_, err = io.Copy(fo, bytes.NewBuffer(body))
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	err = fo.Close()
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
}

func DiffBody(t *testing.T, testdata, name string, got []byte) string {
	t.Helper()

	want, err := os.ReadFile(filepath.Join(testdata, name+".bin"))
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	return cmp.Diff(string(want), string(got))
}
