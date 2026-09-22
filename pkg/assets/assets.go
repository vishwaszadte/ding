package assets

import (
	_ "embed"
	"os"
	"path/filepath"
)

// Go 1.16 introduced the `//go:embed` directive.
// At compile time, the Go compiler reads `icon.png` and `chime.wav` from disk
// and embeds their raw bytes directly into the compiled binary!
// This makes the ding executable completely self-contained.

//go:embed icon.png
var DefaultIcon []byte

//go:embed chime.wav
var DefaultSound []byte

// EnsureAssetsWritten writes the embedded icon and sound to ~/.ding/assets/
// on first run if they don't already exist on disk, returning their absolute paths.
func EnsureAssetsWritten() (iconPath string, soundPath string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}
	assetsDir := filepath.Join(home, ".ding", "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return "", "", err
	}

	iconPath = filepath.Join(assetsDir, "ding-icon.png")
	if _, err := os.Stat(iconPath); os.IsNotExist(err) && len(DefaultIcon) > 0 {
		_ = os.WriteFile(iconPath, DefaultIcon, 0644)
	}

	soundPath = filepath.Join(assetsDir, "chime.wav")
	if _, err := os.Stat(soundPath); os.IsNotExist(err) && len(DefaultSound) > 0 {
		_ = os.WriteFile(soundPath, DefaultSound, 0644)
	}

	return iconPath, soundPath, nil
}
