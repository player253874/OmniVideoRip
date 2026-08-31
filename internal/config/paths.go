package config

import (
	"os"
	"path/filepath"
)

type Paths struct {
	RootDir      string
	ConfigPath   string
	BinDir       string
	LogPath      string
	DownloadsDir string
}

func ResolvePaths() (Paths, error) {
	exePath, err := os.Executable()
	if err != nil {
		exePath, _ = os.Getwd()
	}

	appDir := filepath.Dir(exePath)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = appDir
	}

	downloadsDir := filepath.Join(homeDir, "Downloads")
	binDir := filepath.Join(appDir, "bin")

	return Paths{
		RootDir:      appDir,
		ConfigPath:   filepath.Join(appDir, "config.json"),
		BinDir:       binDir, // Dedicated ./bin/ folder keeps root directory neat
		LogPath:      filepath.Join(appDir, "OmniVideoRip.log"),
		DownloadsDir: downloadsDir,
	}, nil
}

func EnsureDirectories(paths Paths) error {
	if paths.BinDir != "" {
		_ = os.MkdirAll(paths.BinDir, 0755)
	}
	if paths.DownloadsDir != "" {
		_ = os.MkdirAll(paths.DownloadsDir, 0755)
	}
	return nil
}
