import { useCallback, useEffect, useMemo, useState } from "react";
import "./App.css";
import { ConvertMedia, InspectMedia, OpenPath } from "../wailsjs/go/main/App";
import type { main } from "../wailsjs/go/models";
import { OnFileDrop, OnFileDropOff } from "../wailsjs/runtime/runtime";

type MediaKind = "audio" | "video";

type FormatOption = {
  value: string;
  label: string;
  description: string;
};

const formatOptions: Record<MediaKind, FormatOption[]> = {
  video: [
    { value: "mp4", label: "MP4", description: "H.264 video with AAC audio" },
    { value: "m4a", label: "M4A", description: "Extract audio only" },
  ],
  audio: [
    { value: "m4a", label: "M4A", description: "AAC audio" },
    { value: "mp3", label: "MP3", description: "MP3 audio" },
  ],
};

const playbackSpeedOptions = [
  { value: "0.75", label: "0.75x (slower)" },
  { value: "1", label: "1x (normal)" },
  { value: "1.25", label: "1.25x (faster)" },
];

function formatFileSize(bytes?: number): string {
  if (!bytes || bytes <= 0) {
    return "-";
  }
  const units = ["B", "KB", "MB", "GB", "TB"];
  let value = bytes;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }
  const precision = value < 10 && unitIndex > 0 ? 1 : 0;
  return `${value.toFixed(precision)} ${units[unitIndex]}`;
}

function formatDuration(totalSeconds?: number): string {
  if (!totalSeconds || totalSeconds <= 0) {
    return "-";
  }
  const seconds = Math.floor(totalSeconds % 60);
  const minutes = Math.floor((totalSeconds / 60) % 60);
  const hours = Math.floor(totalSeconds / 3600);
  const paddedSeconds = seconds.toString().padStart(2, "0");
  const paddedMinutes = minutes.toString().padStart(2, "0");
  if (hours > 0) {
    return `${hours}:${paddedMinutes}:${paddedSeconds}`;
  }
  return `${minutes}:${paddedSeconds}`;
}

function parseError(error: unknown): string {
  if (!error) {
    return "Something went wrong";
  }
  if (typeof error === "string") {
    return error;
  }
  if (error instanceof Error) {
    return error.message;
  }
  try {
    return JSON.stringify(error);
  } catch (_serializationError) {
    return String(error);
  }
}

function App() {
  const [media, setMedia] = useState<main.MediaInfo | null>(null);
  const [targetFormat, setTargetFormat] = useState<string>("");
  const [playbackSpeed, setPlaybackSpeed] = useState<string>("1");
  const [error, setError] = useState<string | null>(null);
  const [isInspecting, setIsInspecting] = useState(false);
  const [isConverting, setIsConverting] = useState(false);
  const [result, setResult] = useState<main.ConversionResult | null>(null);

  useEffect(() => {
    if (!media) {
      setTargetFormat("");
      setPlaybackSpeed("1");
      return;
    }
    const options = formatOptions[media.kind as MediaKind] ?? [];
    setTargetFormat(options[0]?.value ?? "");
    setPlaybackSpeed("1");
    setResult(null);
  }, [media]);

  const availableFormats = useMemo(() => {
    if (!media) {
      return [];
    }
    return formatOptions[media.kind as MediaKind] ?? [];
  }, [media]);

  useEffect(() => {
    if (targetFormat !== "mp4") {
      setPlaybackSpeed("1");
    }
  }, [targetFormat]);

  const mediaMeta = useMemo(() => {
    if (!media) {
      return [] as string[];
    }
    const pieces: string[] = [];
    if (media.duration && media.duration > 0) {
      pieces.push(formatDuration(media.duration));
    }
    if (media.size && media.size > 0) {
      pieces.push(formatFileSize(media.size));
    }
    return pieces;
  }, [media]);

  const resultPath = result?.output?.path ?? "";

  const processPath = useCallback((filePath: string) => {
    if (!filePath) {
      setError("Drop a media file to begin.");
      return;
    }

    setIsInspecting(true);
    setError(null);
    setResult(null);

    InspectMedia(filePath)
      .then((info) => {
        setMedia(info);
      })
      .catch((inspectError) => {
        setMedia(null);
        setError(parseError(inspectError));
      })
      .finally(() => {
        setIsInspecting(false);
      });
  }, []);

  useEffect(() => {
    OnFileDrop((_x, _y, paths) => {
      if (!paths || paths.length === 0) {
        setError("Drop a media file to begin.");
        return;
      }
      processPath(paths[0]);
    }, true);
    return () => {
      OnFileDropOff();
    };
  }, [processPath]);

  const handleConvert = useCallback(() => {
    if (!media || !targetFormat) {
      return;
    }
    const selectedSpeed =
      targetFormat === "mp4" ? Number(playbackSpeed) || 1 : 1;
    setIsConverting(true);
    setError(null);

    ConvertMedia(media.path, targetFormat, selectedSpeed)
      .then((conversion) => {
        setResult(conversion);
      })
      .catch((convertError) => {
        setResult(null);
        setError(parseError(convertError));
      })
      .finally(() => {
        setIsConverting(false);
      });
  }, [media, targetFormat, playbackSpeed]);

  const handleOpenOutput = useCallback(() => {
    if (!resultPath) {
      return;
    }
    OpenPath(resultPath).catch((openError) => {
      setError(parseError(openError));
    });
  }, [resultPath]);

  return (
    <div id="App">
      <header className="app-header">
        <h1>Media Converter</h1>
        <p>Drop a video or audio file anywhere in the window to get started.</p>
      </header>

      <section className="dropzone" role="presentation">
        <div className="dropzone__content">
          <p className="dropzone__title">Drag &amp; drop a media file here</p>
          <p className="dropzone__hint">
            Supports audio and video files from your computer.
          </p>
        </div>
      </section>

      {isInspecting && <div className="app-message">Inspecting media…</div>}
      {error && <div className="app-message app-message--error">{error}</div>}

      {media && (
        <section
          className={`workspace${result ? " workspace--with-result" : ""}`}
        >
          <div className="workspace__main">
            <div className="selected-media">
              <div className="selected-media__heading">
                <h2>
                  {media.name}
                  <br />
                  <span className="selected-media__path">{media.path}</span>
                </h2>

                {mediaMeta.length > 0 && (
                  <div className="selected-media__meta">
                    {mediaMeta.map((item, index) => (
                      <span key={`${item}-${index}`}>{item}</span>
                    ))}
                  </div>
                )}
                <span className="selected-media__badge">
                  {media.kind === "video" ? "Video" : "Audio"}
                </span>
              </div>
            </div>

            {availableFormats.length > 0 && (
              <div className="conversion-panel">
                <div className="conversion-panel__header">
                  <h3>Select a output format</h3>
                </div>
                <div className="format-options">
                  {availableFormats.map((option) => (
                    <label
                      key={option.value}
                      className={`format-option${
                        targetFormat === option.value
                          ? " format-option--selected"
                          : ""
                      }`}
                    >
                      <input
                        type="radio"
                        name="target-format"
                        value={option.value}
                        checked={targetFormat === option.value}
                        onChange={(event) =>
                          setTargetFormat(event.target.value)
                        }
                      />
                      <div className="format-option__content">
                        <span className="format-option__label">
                          {option.label}
                        </span>
                        <span className="format-option__description">
                          {option.description}
                        </span>
                      </div>
                    </label>
                  ))}
                </div>
                {media.kind === "video" && targetFormat === "mp4" && (
                  <div className="speed-selector">
                    <label
                      htmlFor="playback-speed"
                      className="speed-selector__label"
                    >
                      Playback speed
                    </label>
                    <select
                      id="playback-speed"
                      className="speed-selector__control"
                      value={playbackSpeed}
                      onChange={(event) => setPlaybackSpeed(event.target.value)}
                      disabled={isConverting}
                    >
                      {playbackSpeedOptions.map((option) => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      ))}
                    </select>
                  </div>
                )}
                <button
                  type="button"
                  className="primary-button"
                  onClick={handleConvert}
                  disabled={isConverting || !targetFormat}
                >
                  {isConverting
                    ? "Converting…"
                    : `Convert to ${targetFormat.toUpperCase()}`}
                </button>
              </div>
            )}
          </div>

          {result && (
            <aside className="workspace__result">
              <button
                type="button"
                className="result-card"
                onClick={handleOpenOutput}
                disabled={!resultPath}
              >
                <div className="result-card__visual" aria-hidden="true">
                  {(result.target ?? targetFormat).toUpperCase() || ""}
                </div>
                <div className="result-card__text">
                  <h3>Conversion complete</h3>
                  <p className="result-card__name">
                    {result.output?.name ?? "File created"}
                  </p>
                  {resultPath && (
                    <p className="result-card__path" title={resultPath}>
                      {resultPath}
                    </p>
                  )}
                </div>
              </button>
            </aside>
          )}
        </section>
      )}
    </div>
  );
}

export default App;
