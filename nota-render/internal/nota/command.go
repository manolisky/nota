package nota

import (
	"os"
	"os/exec"
	"path/filepath"
)

func lookCommand(name string) (string, error) {
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}
	for _, dir := range []string{"/opt/homebrew/bin", "/usr/local/bin", "/Library/TeX/texbin"} {
		candidate := filepath.Join(dir, name)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() && st.Mode()&0o111 != 0 {
			return candidate, nil
		}
	}
	return "", exec.ErrNotFound
}
