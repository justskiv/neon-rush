# Neon Rush

A vertical arcade racing game built as an experiment to test the [Ebiten](https://ebitengine.org/) game engine and explore AI-agent capabilities in game development.

This is a toy project — not intended for production use.

## Build

Requires Go 1.26+ and [Task](https://taskfile.dev/).

```bash
task build          # macOS
task build:win      # Windows (cross-compile)
task build:web      # Browser (WASM)
task run            # Run locally
```

Binaries go to `build/`. The web build produces a self-contained `build/web/` folder ready for static hosting (e.g. GitHub Pages).