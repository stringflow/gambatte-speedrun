package assets

import (
	_ "embed"
)

var (
	//go:embed keys.png
	KeysPng []byte

	//go:embed arrows.png
	ArrowsPng []byte

	//go:embed font.png
	FontPng []byte
)
