package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ffprobeFormat struct {
	Duration string `json:"duration"`
}

type ffprobeStream struct {
	CodecType string `json:"codec_type"`
}

type ffprobeOutput struct {
	Format  ffprobeFormat   `json:"format"`
	Streams []ffprobeStream `json:"streams"`
}

func loadMediaInfo(path string) (*MediaInfo, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("file not found: %w", err)
		}
		return nil, fmt.Errorf("stat file: %w", err)
	}
	if stat.IsDir() {
		return nil, errors.New("expected a file but got a directory")
	}

	kind, duration, err := probeMedia(path)
	if err != nil {
		return nil, err
	}

	info := &MediaInfo{
		Path: path,
		Name: filepath.Base(path),
		Kind: kind,
		Size: stat.Size(),
	}
	if duration > 0 {
		info.Duration = duration
	}
	return info, nil
}

func probeMedia(path string) (string, float64, error) {
	cmd := execCommand("ffprobe", "-v", "error", "-show_streams", "-show_format", "-of", "json", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		trimmed := strings.TrimSpace(string(out))
		if trimmed == "" {
			return "", 0, fmt.Errorf("ffprobe failed: %w", err)
		}
		return "", 0, fmt.Errorf("ffprobe failed: %w; %s", err, trimmed)
	}

	var parsed ffprobeOutput
	if err := json.Unmarshal(out, &parsed); err != nil {
		return "", 0, fmt.Errorf("parse ffprobe output: %w", err)
	}

	kind := ""
	for _, stream := range parsed.Streams {
		if stream.CodecType == "video" {
			kind = "video"
			break
		}
		if stream.CodecType == "audio" && kind == "" {
			kind = "audio"
		}
	}
	if kind == "" {
		return "", 0, errors.New("unsupported media: no audio or video stream")
	}

	var duration float64
	if parsed.Format.Duration != "" {
		duration, err = strconv.ParseFloat(parsed.Format.Duration, 64)
		if err != nil {
			duration = 0
		}
	}

	return kind, duration, nil
}

func convertMedia(path, targetFormat string) (*ConversionResult, error) {
	sourceInfo, err := loadMediaInfo(path)
	if err != nil {
		return nil, err
	}

	format := strings.ToLower(strings.TrimPrefix(targetFormat, "."))
	if format == "" {
		return nil, errors.New("invalid target format")
	}

	outputPath, err := buildOutputPath(sourceInfo.Path, format)
	if err != nil {
		return nil, err
	}

	args := []string{"-y", "-i", sourceInfo.Path}

	switch format {
	case "mp4":
		if sourceInfo.Kind != "video" {
			return nil, fmt.Errorf("cannot convert %s to mp4", sourceInfo.Kind)
		}
		args = append(args, "-c:v", "libx264", "-preset", "medium", "-crf", "23", "-c:a", "aac", "-movflags", "+faststart", outputPath)
	case "m4a":
		if sourceInfo.Kind == "video" {
			args = append(args, "-vn")
		}
		args = append(args, "-c:a", "aac", outputPath)
	case "mp3":
		if sourceInfo.Kind != "audio" {
			return nil, fmt.Errorf("cannot convert %s to mp3", sourceInfo.Kind)
		}
		args = append(args, "-codec:a", "libmp3lame", "-qscale:a", "2", outputPath)
	default:
		return nil, fmt.Errorf("unsupported target format: %s", format)
	}

	cmd := execCommand("ffmpeg", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		trimmed := strings.TrimSpace(string(out))
		if trimmed == "" {
			return nil, fmt.Errorf("ffmpeg failed: %w", err)
		}
		return nil, fmt.Errorf("ffmpeg failed: %w; %s", err, trimmed)
	}

	outputInfo, err := loadMediaInfo(outputPath)
	if err != nil {
		return nil, fmt.Errorf("conversion succeeded but inspecting output failed: %w", err)
	}

	return &ConversionResult{
		Source: sourceInfo,
		Output: outputInfo,
		Target: format,
	}, nil
}

func buildOutputPath(inputPath, extension string) (string, error) {
	dir := filepath.Dir(inputPath)
	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	if base == "" {
		base = "converted"
	}

	candidate := filepath.Join(dir, fmt.Sprintf("%s_converted.%s", base, extension))
	index := 1

	for {
		_, err := os.Stat(candidate)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return candidate, nil
			}
			return "", fmt.Errorf("check output path: %w", err)
		}
		candidate = filepath.Join(dir, fmt.Sprintf("%s_converted_%d.%s", base, index, extension))
		index++
	}
}
