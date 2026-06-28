// Package toolconfig manages config files for file-configured tools (those that
// read a config file rather than environment variables), writing atomically and
// keeping a timestamped backup of anything it replaces.
package toolconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// now is overridable in tests; production uses the wall clock.
var now = time.Now

// backup copies path to "path.eg-bak-<timestamp>" if it exists. Returns the
// backup path (empty if there was nothing to back up).
func backup(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	bak := fmt.Sprintf("%s.eg-bak-%s", path, now().Format("20060102-150405"))
	if err := os.WriteFile(bak, data, 0o600); err != nil {
		return "", err
	}
	return bak, nil
}

// atomicWrite writes data to path via a temp file + rename, creating parent
// directories as needed.
func atomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".eg-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op if the rename succeeded
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
