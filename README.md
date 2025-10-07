# DoReveal Tools

DoReveal Tools is a desktop suite of media utilities for inspecting and converting audio or video files. Use it to change formats, adjust quality, or prepare assets for different devices with a streamlined, offline workflow powered by a Go backend and React UI.

<p align="center">
  <img width="40%" alt="Screenshot 2025-10-07 at 2 01 23 PM" src="https://github.com/user-attachments/assets/56341cc8-9b6b-4077-a9e5-8f0252274eda" />
  <img width="40%" alt="Screenshot 2025-10-07 at 2 02 04 PM" src="https://github.com/user-attachments/assets/d31a48e4-10f9-4466-8e8b-eeac4f77e102" />
</p>

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

## Package for macOS
- Build a universal macOS bundle locally with:
  ```bash
  wails build -clean -platform darwin/universal
  ```
  This regenerates the frontend, produces `DoReveal-Tools.app`, and creates a signed disk image at `build/bin/DoReveal-Tools.dmg` (or a similarly named `.dmg` if you change the app name).
- Upload the `.dmg` when drafting a GitHub release, or copy it to other machines for distribution.
- You can also run the **macOS Release Build** workflow from the Actions tab (or push a tag) to rebuild on GitHub-hosted macOS runners. The workflow keeps the DMG as a downloadable artifact and automatically attaches it to the tag’s release so you can publish without rebuilding locally.
