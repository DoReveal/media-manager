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

func ensureFFmpeg() error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get cache dir: %w", err)
	}
	appCache := filepath.Join(cacheDir, appName)
	installedFlag := filepath.Join(appCache, "installed")

	// Check if already installed
	if _, err := os.Stat(installedFlag); err == nil {
		setFFmpegPath(appCache)
		return nil
	}

	// Install if needed
	if err := os.MkdirAll(appCache, 0755); err != nil {
		return fmt.Errorf("failed to create cache dir: %w", err)
	}
	if err := downloadAndExtractFFmpeg(appCache); err != nil {
		return fmt.Errorf("FFmpeg setup failed: %w", err)
	}

	// Mark as installed
	if err := os.WriteFile(installedFlag, []byte("1"), 0644); err != nil {
		return fmt.Errorf("failed to write install flag: %w", err)
	}

	setFFmpegPath(appCache)
	return nil
}

func downloadAndExtractFFmpeg(targetDir string) error {
	var url string
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			url = "https://ffmpeg.martin-riedl.de/redirect/latest/macos/arm64/release/ffmpeg.zip"
		} else { // amd64/386 fallback to amd64
			url = "https://ffmpeg.martin-riedl.de/redirect/latest/macos/amd64/release/ffmpeg.zip"
		}
	case "windows":
		if runtime.GOARCH != "amd64" {
			return errors.New("windows arm64 not supported; use x64 build")
		}
		url = "https://ffmpeg.martin-riedl.de/redirect/latest/windows/amd64/release/ffmpeg.zip"
	default:
		return errors.New("unsupported OS")
	}

	// Download ZIP to temp
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	tmpZip := filepath.Join(os.TempDir(), "ffmpeg.zip")
	out, err := os.Create(tmpZip)
	if err != nil {
		return fmt.Errorf("create temp file failed: %w", err)
	}
	defer out.Close()
	defer os.Remove(tmpZip) // Clean up

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("copy download failed: %w", err)
	}

	// Extract ZIP
	r, err := zip.OpenReader(tmpZip)
	if err != nil {
		return fmt.Errorf("open ZIP failed: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			continue // Skip corrupted
		}
		defer rc.Close()

		fPath := filepath.Join(targetDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fPath, f.Mode())
			continue
		}

		wf, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("extract %s failed: %w", f.Name, err)
		}
		defer wf.Close()

		if _, err := io.Copy(wf, rc); err != nil {
			return fmt.Errorf("copy %s failed: %w", f.Name, err)
		}
	}

	// Make executable on macOS
	if runtime.GOOS == "darwin" {
		binPath := filepath.Join(targetDir, "ffmpeg")
		if err := os.Chmod(binPath, 0755); err != nil {
			return fmt.Errorf("chmod failed: %w", err)
		}
	}

	// Quick verify
	cmd := execCommand("ffmpeg", "-version")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("FFmpeg verify failed: %w; output: %s", err, out)
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

func execCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Dir = "" // Inherit cwd
	return cmd
}
