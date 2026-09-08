//go:build !windows

package sidecars

import "os"

func replaceComponentFile(source, target string) error { return os.Rename(source, target) }
