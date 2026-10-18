package mihomo

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ReleaseBase is where MetaCubeX publishes Mihomo builds.
const ReleaseBase = "https://github.com/MetaCubeX/mihomo/releases"

// Installed is the result of an install.
type Installed struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

// Install downloads the latest Mihomo release for this machine to dest.
// It reads the release's version.txt rather than GitHub's API, which rate
// limits anonymous callers.
func Install(ctx context.Context, dest string) (Installed, error) {
	client := &http.Client{Timeout: 5 * time.Minute}
	version, err := get(ctx, client, ReleaseBase+"/latest/download/version.txt")
	if err != nil {
		return Installed{}, fmt.Errorf("reading the latest Mihomo version: %w", err)
	}
	tag := strings.TrimSpace(string(version))
	asset := AssetName(runtime.GOOS, runtime.GOARCH, tag)
	archive, err := get(ctx, client, ReleaseBase+"/download/"+tag+"/"+asset)
	if err != nil {
		return Installed{}, fmt.Errorf("downloading %s: %w", asset, err)
	}
	binary, err := extract(asset, archive)
	if err != nil {
		return Installed{}, err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return Installed{}, err
	}
	tmp := dest + ".download"
	if err := os.WriteFile(tmp, binary, 0o755); err != nil {
		return Installed{}, err
	}
	if err := os.Rename(tmp, dest); err != nil {
		return Installed{}, err
	}
	return Installed{Path: dest, Version: tag}, nil
}

// AssetName is the release file for a platform, such as
// mihomo-darwin-arm64-v1.19.32.gz or mihomo-windows-amd64-v1.19.32.zip.
// amd64 builds use the "compatible" variant, which runs on every x86-64 CPU.
func AssetName(goos, goarch, tag string) string {
	arch := goarch
	if goarch == "amd64" {
		arch = "amd64-compatible"
	}
	ext := ".gz"
	if goos == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("mihomo-%s-%s-%s%s", goos, arch, tag, ext)
}

func get(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 256<<20))
}

func extract(name string, archive []byte) ([]byte, error) {
	if strings.HasSuffix(name, ".zip") {
		r, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
		if err != nil {
			return nil, err
		}
		for _, f := range r.File {
			if strings.HasSuffix(strings.ToLower(f.Name), ".exe") {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()
				return io.ReadAll(rc)
			}
		}
		return nil, fmt.Errorf("%s holds no executable", name)
	}
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	return io.ReadAll(gz)
}
