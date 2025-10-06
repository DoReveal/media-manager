# DoReveal Tools

DoReveal Tools is a desktop suite of media utilities for inspecting and converting audio or video files. Use it to change formats, adjust quality, or prepare assets for different devices with a streamlined, offline workflow powered by a Go backend and React UI.

## Prerequisites
- macOS with [Homebrew](https://brew.sh/) available in your shell
- Git for cloning this repository

### Install Go with Homebrew
```bash
brew update
brew install go

go version
```
If `go` is not on your PATH after installation, add `$(go env GOPATH)/bin` to your shell profile (e.g. `~/.zshrc`).

### Install Node.js (for the React app)
```bash
brew install node

node --version
npm --version
```

### Install the Wails CLI
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```
Make sure `$(go env GOPATH)/bin` is on your PATH so the `wails` command is available.

## Project Setup
```bash
git clone https://github.com/DoReveal/media-manager.git doreveal-tools
cd doreveal-tools

# Install frontend dependencies
cd frontend
npm install
cd ..
```

## Development Server
Run the app with hot reload:
```bash
wails dev
```
This starts the Go backend and the Vite dev server together.

## Build the App
Create a production build:
```bash
wails build
```
The packaged binaries will be written to `build/bin/`.
