# DoReveal Tools

A Wails desktop app that combines a Go backend with a React frontend.

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
git clone <repo-url>
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
