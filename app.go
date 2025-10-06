package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	wailsruntime.OnFileDrop(ctx, func(x, y int, paths []string) {
		fmt.Printf("Dropped files at (%d, %d): %v\n", x, y, paths)
	})
}

type MediaInfo struct {
	Path     string  `json:"path"`
	Name     string  `json:"name"`
	Kind     string  `json:"kind"`
	Duration float64 `json:"duration"`
	Size     int64   `json:"size"`
}

type ConversionResult struct {
	Source *MediaInfo `json:"source"`
	Output *MediaInfo `json:"output"`
	Target string     `json:"target"`
}

func (a *App) InspectMedia(path string) (*MediaInfo, error) {
	if path == "" {
		return nil, errors.New("path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}
	info, err := loadMediaInfo(absolute)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (a *App) ConvertMedia(sourcePath, targetFormat string) (*ConversionResult, error) {
	if sourcePath == "" {
		return nil, errors.New("sourcePath is required")
	}
	if targetFormat == "" {
		return nil, errors.New("targetFormat is required")
	}
	absolute, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}
	result, err := convertMedia(absolute, targetFormat)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *App) OpenPath(path string) error {
	if path == "" {
		return errors.New("path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}
	if _, err := os.Stat(absolute); err != nil {
		return fmt.Errorf("stat path: %w", err)
	}

	var cmd *exec.Cmd
	switch goruntime.GOOS {
	case "darwin":
		cmd = exec.Command("open", absolute)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", absolute)
	default:
		cmd = exec.Command("xdg-open", absolute)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open path: %w", err)
	}
	return nil
}
