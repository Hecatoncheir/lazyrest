package finder

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/git-pkgs/gitignore"
)

// gitIgnoreMatcher loads each .gitignore only when its directory is reached.
// This preserves Git's scoping and precedence without walking ignored trees a
// second time before the real scan starts.
type gitIgnoreMatcher struct {
	root       string
	matcher    *gitignore.Matcher
	errorCount int
}

func newGitIgnoreMatcher(root string) *gitIgnoreMatcher {
	return &gitIgnoreMatcher{
		root:    root,
		matcher: gitignore.New(""),
	}
}

func (matcher *gitIgnoreMatcher) load(directoryPath string) []string {
	ignorePath := filepath.Join(directoryPath, ".gitignore")
	contents, err := os.ReadFile(ignorePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return []string{fmt.Sprintf("%s: %v", ignorePath, err)}
	}

	relativeDirectory, err := filepath.Rel(matcher.root, directoryPath)
	if err != nil {
		return []string{fmt.Sprintf("%s: resolve .gitignore scope: %v", ignorePath, err)}
	}
	if relativeDirectory == "." {
		relativeDirectory = ""
	}
	matcher.matcher.AddPatterns(contents, filepath.ToSlash(relativeDirectory))

	patternErrors := matcher.matcher.Errors()
	warnings := make([]string, 0, len(patternErrors)-matcher.errorCount)
	for _, patternError := range patternErrors[matcher.errorCount:] {
		warnings = append(warnings, fmt.Sprintf(
			"%s:%d: invalid .gitignore pattern %q: %s",
			ignorePath,
			patternError.Line,
			patternError.Pattern,
			patternError.Message,
		))
	}
	matcher.errorCount = len(patternErrors)
	return warnings
}

func (matcher *gitIgnoreMatcher) matches(path string, isDirectory bool) bool {
	relativePath, err := filepath.Rel(matcher.root, path)
	if err != nil {
		return false
	}
	return matcher.matcher.MatchPath(filepath.ToSlash(relativePath), isDirectory)
}
