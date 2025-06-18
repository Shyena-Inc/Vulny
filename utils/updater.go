package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
)

// UpdateURL is where Vulny releases are hosted
const UpdateURL = "https://github.com/Shyena-Inc/Vulny/releases/latest/download"

// GetBinaryName returns the expected binary filename for current OS
func GetBinaryName() string {
	switch runtime.GOOS {
	case "linux":
		return "vulny-linux"
	case "darwin":
		return "vulny-macos"
	case "windows":
		return "vulny-windows.exe"
	default:
		return "vulny"
	}
}

// UpdateBinary downloads the latest version and replaces the current executable
func UpdateBinary() error {
	binaryName := GetBinaryName()
	url := fmt.Sprintf("%s/%s", UpdateURL, binaryName)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download update: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("update server returned: %s", resp.Status)
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not get executable path: %v", err)
	}

	tmpPath := exePath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("could not create temp file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed writing binary: %v", err)
	}

	if runtime.GOOS != "windows" {
		os.Chmod(tmpPath, 0755)
	}

	err = os.Rename(tmpPath, exePath)
	if err != nil {
		return fmt.Errorf("failed replacing binary: %v", err)
	}

	return nil
}
