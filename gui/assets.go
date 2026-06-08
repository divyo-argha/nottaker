package main

import "embed"

// assets embeds the entire frontend directory into the binary.
// Wails serves these files directly from memory — no web server required.
//
//go:embed frontend
var assets embed.FS
