package cli

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

// CopyToClipboard sends `text` to the system clipboard. It shells out to
// platform-native tooling (pbcopy / wl-copy / xclip / xsel / clip.exe). On
// success it returns the name of the tool used.
func CopyToClipboard(text string) (string, error) {
	var candidates [][]string
	switch runtime.GOOS {
	case "darwin":
		candidates = [][]string{{"pbcopy"}}
	case "windows":
		candidates = [][]string{{"clip"}}
	default:
		candidates = [][]string{
			{"wl-copy"},
			{"xclip", "-selection", "clipboard"},
			{"xsel", "--clipboard", "--input"},
		}
	}
	for _, c := range candidates {
		if _, err := exec.LookPath(c[0]); err != nil {
			continue
		}
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return c[0], nil
		}
	}
	return "", errors.New("no clipboard tool available")
}
