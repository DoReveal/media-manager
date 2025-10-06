# Repository Guidelines

## Project Structure & Module Organization
The Go backend lives at the repository root: `main.go` starts the Wails runtime, `app.go` hosts bound methods, and `ffmpeg.go` handles platform-specific FFmpeg bootstrapping. Asset and installer templates reside under `build/`. The web UI code is under `frontend/` with `src/` for React components, generated bindings in `wailsjs/`, and build output in `dist/`. Add new backend packages at the root and keep shared TypeScript helpers under `frontend/src/`. Never edit files in `frontend/wailsjs/go` manually—rerun the Wails generator instead so the RPC stubs remain in sync.

## Build, Test, and Development Commands
- `wails dev` (root): runs the Go backend and Vite dev server with hot reload.
- `wails build`: produces platform-specific bundles and populates `build/bin/`.
- `npm install` (inside `frontend/`): syncs the React/Vite dependencies.
- `npm run dev` (inside `frontend/`): useful for front-end-only work and mirrors the Wails Vite server.
- `npm run build`: runs `tsc` type-checking and generates optimized assets in `frontend/dist/`.

## Coding Style & Naming Conventions
Use `gofmt` (tabs, grouped imports) before committing Go changes; public APIs stay PascalCase while internal helpers use camelCase. Keep context parameters first and prefer returning explicit errors. TypeScript files use 2-space indentation, arrow functions for handlers, and PascalCase component filenames (e.g., `MediaGrid.tsx`). Re-export bound Go calls from `frontend/wailsjs` rather than duplicating fetch logic, and rely on your editor or Prettier for TSX formatting.

## Testing Guidelines
Backend tests belong beside the code as `*_test.go` files using Go’s standard `testing` package; run them with `go test ./...` from the repo root. Mock external binaries by injecting an `execCommand` wrapper so FFmpeg downloads remain testable. The frontend currently lacks an automated test harness—when introducing one, prefer Vitest with React Testing Library under `frontend/src/__tests__/` and document the command in your PR. Until then, exercise new UI flows through `wails dev` and describe manual coverage in review notes.

## Commit & Pull Request Guidelines
Follow the existing short, imperative commit style (`ffmpeg setup`, `initial commit`). Keep subject lines under 60 characters and detail reasoning in the body when behavior changes. Pull requests should describe the feature or fix, note any FFmpeg or build implications, call out test coverage (automated or manual), and attach screenshots or GIFs for UI updates.

## Environment & Configuration
`wails.json` governs window dimensions, icons, and binding metadata—update it alongside UI changes. The app installs FFmpeg on startup into the user cache; if you modify this flow, note platform considerations in the PR and confirm binaries stay executable on macOS and Windows.
