package chromium

import "runtime"

func Platform() string {
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "mac-arm64"
		}
		return "mac-x64"
	case "linux":
		return "linux64"
	case "windows":
		return "win64"
	default:
		return "linux64"
	}
}

func ExecutableName() string {
	if runtime.GOOS == "windows" {
		return "chrome.exe"
	}
	return "chrome"
}
