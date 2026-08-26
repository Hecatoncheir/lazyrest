package http

import (
	"fmt"
	"os"
	"path/filepath"
)

// maxExternalBodyBytes caps the size of a file a request may send as its body.
const maxExternalBodyBytes = int64(10 << 20)

// loadExternalBody replaces a body that names a file with the contents of that
// file, resolved against the directory of the parsed file.
func loadExternalBody(suite *HttpSuite, baseDirectory string) error {
	path, external := externalBodyPath(suite.Body)
	if !external {
		return nil
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDirectory, path)
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("read request body file: %w", err)
	}
	if info.Size() > maxExternalBodyBytes {
		return fmt.Errorf("request body file %s is larger than %d bytes", path, maxExternalBodyBytes)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read request body file: %w", err)
	}

	suite.Body = string(contents)
	return nil
}
