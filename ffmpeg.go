package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const appName = "doreveal-tools"

type binaryAsset struct {
	name string
	url  string
	path string
}

func ensureFFmpeg() error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get cache dir: %w", err)
	}
	appCache := filepath.Join(cacheDir, appName)
	installedFlag := filepath.Join(appCache, "installed")
	binaries, err := platformBinaryAssets(appCache)
	if err != nil {
		return err
	}

	// Check if already installed
	if binariesReady(binaries) {
		setFFmpegPath(appCache)
		return nil
	}

	// Install if needed
	if err := os.MkdirAll(appCache, 0755); err != nil {
		return fmt.Errorf("failed to create cache dir: %w", err)
	}
	if err := downloadAndExtractFFmpeg(appCache, binaries); err != nil {
		return fmt.Errorf("FFmpeg setup failed: %w", err)
	}

	// Mark as installed
	if err := os.WriteFile(installedFlag, []byte("1"), 0644); err != nil {
		return fmt.Errorf("failed to write install flag: %w", err)
	}

	setFFmpegPath(appCache)
	return nil
}

func platformBinaryAssets(targetDir string) ([]binaryAsset, error) {
	switch runtime.GOOS {
	case "darwin":
		arch := "amd64"
		if runtime.GOARCH == "arm64" {
			arch = "arm64"
		}
		base := fmt.Sprintf("https://ffmpeg.martin-riedl.de/redirect/latest/macos/%s/release/", arch)
		return []binaryAsset{
			{name: "ffmpeg", url: base + "ffmpeg.zip", path: ffmpegBinaryPath(targetDir)},
			{name: "ffprobe", url: base + "ffprobe.zip", path: ffprobeBinaryPath(targetDir)},
		}, nil
	case "windows":
		if runtime.GOARCH != "amd64" {
			return nil, errors.New("windows arm64 not supported; use x64 build")
		}
		base := "https://ffmpeg.martin-riedl.de/redirect/latest/windows/amd64/release/"
		return []binaryAsset{
			{name: "ffmpeg", url: base + "ffmpeg.zip", path: ffmpegBinaryPath(targetDir)},
			{name: "ffprobe", url: base + "ffprobe.zip", path: ffprobeBinaryPath(targetDir)},
		}, nil
	default:
		return nil, errors.New("unsupported OS")
	}
}

func binariesReady(binaries []binaryAsset) bool {
	for _, bin := range binaries {
		info, err := os.Stat(bin.path)
		if err != nil {
			return false
		}
		if info.IsDir() {
			return false
		}
	}
	return true
}

func downloadAndExtractFFmpeg(targetDir string, binaries []binaryAsset) error {
	for _, bin := range binaries {
		if err := downloadAndExtract(bin.url, targetDir); err != nil {
			return fmt.Errorf("download %s failed: %w", bin.name, err)
		}
		if _, err := os.Stat(bin.path); err != nil {
			return fmt.Errorf("%s missing after download: %w", bin.name, err)
		}
		if runtime.GOOS == "darwin" {
			if err := os.Chmod(bin.path, 0755); err != nil {
				return fmt.Errorf("chmod %s failed: %w", bin.name, err)
			}
		}
		cmd := execCommand(bin.path, "-version")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("%s verify failed: %w; output: %s", bin.name, err, out)
		}
	}
	return nil
}

func downloadAndExtract(url, targetDir string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	out, err := os.CreateTemp("", "ffmpeg-download-*.zip")
	if err != nil {
		return fmt.Errorf("create temp file failed: %w", err)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		os.Remove(out.Name())
		return fmt.Errorf("copy download failed: %w", err)
	}
	out.Close()
	defer os.Remove(out.Name())

	r, err := zip.OpenReader(out.Name())
	if err != nil {
		return fmt.Errorf("open ZIP failed: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		fPath := filepath.Join(targetDir, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fPath, f.Mode()); err != nil {
				return fmt.Errorf("create dir %s failed: %w", f.Name, err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fPath), 0755); err != nil {
			return fmt.Errorf("create parent dir for %s failed: %w", f.Name, err)
		}
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open %s failed: %w", f.Name, err)
		}
		wf, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("extract %s failed: %w", f.Name, err)
		}
		if _, err := io.Copy(wf, rc); err != nil {
			wf.Close()
			rc.Close()
			return fmt.Errorf("copy %s failed: %w", f.Name, err)
		}
		wf.Close()
		rc.Close()
	}

	return nil
}

func setFFmpegPath(binDir string) {
	fmt.Println("FFmpeg installed and verified.", ffmpegBinaryPath(binDir))

	sep := ":"
	if runtime.GOOS == "windows" {
		sep = ";"
	}
	newPath := binDir + sep + os.Getenv("PATH")
	os.Setenv("PATH", newPath)
}

func ffmpegBinaryPath(dir string) string {
	name := "ffmpeg"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(dir, name)
}

func ffprobeBinaryPath(dir string) string {
	name := "ffprobe"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(dir, name)
}

func execCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Dir = "" // Inherit cwd
	return cmd
}
