package finder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// DefaultIgnoredDirectories are never descended into. They hold no request
// files worth showing and can be very large.
var DefaultIgnoredDirectories = []string{
	".git", ".hg", ".svn", ".cache", ".venv", ".tox",
	"node_modules", "vendor", "target", "dist", "build",
}

// DefaultMaxDepth bounds how deep a scan goes. It is high enough for any real
// project and low enough to stop a pathological tree.
const DefaultMaxDepth = 32

// Options steers a scan.
type Options struct {
	Extensions []string
	// Ignore names directories to skip on top of DefaultIgnoredDirectories.
	Ignore []string
	// MaxDepth bounds the depth of the scan. Zero selects DefaultMaxDepth.
	MaxDepth int
}

func FindFilesInDirectory(directoryPath string, extensions []string) (Directory, error) {
	return FindFilesInDirectoryContext(context.Background(), directoryPath, extensions)
}

func FindFilesInDirectoryContext(ctx context.Context, directoryPath string, extensions []string) (Directory, error) {
	return Find(ctx, directoryPath, Options{Extensions: extensions})
}

// Find walks a directory tree for the files a scan is after. Symbolic links to
// directories are followed once each, so a tree that links back into itself
// terminates.
func Find(ctx context.Context, directoryPath string, options Options) (Directory, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	maxDepth := options.MaxDepth
	if maxDepth <= 0 {
		maxDepth = DefaultMaxDepth
	}
	ignored := make(map[string]struct{}, len(DefaultIgnoredDirectories)+len(options.Ignore))
	for _, name := range slices.Concat(DefaultIgnoredDirectories, options.Ignore) {
		if name = strings.TrimSpace(name); name != "" {
			ignored[name] = struct{}{}
		}
	}

	scan := &scanner{
		extensions: options.Extensions,
		ignored:    ignored,
		gitIgnore:  newGitIgnoreMatcher(filepath.Clean(directoryPath)),
		maxDepth:   maxDepth,
		visited:    map[string]struct{}{},
	}
	return scan.walk(ctx, filepath.Clean(directoryPath), 0)
}

type scanner struct {
	extensions []string
	ignored    map[string]struct{}
	gitIgnore  *gitIgnoreMatcher
	maxDepth   int
	// visited holds the directories already walked, resolved through any
	// symbolic link, so that a link back into the tree is not followed twice.
	visited map[string]struct{}
}

// walk reads one directory and everything under it. A problem with a single
// entry becomes a warning and the walk continues; only a cancelled context
// stops it.
func (scan *scanner) walk(ctx context.Context, directoryPath string, depth int) (Directory, error) {
	directory := Directory{
		Name:        filepath.Base(directoryPath),
		Path:        directoryPath,
		Directories: []Directory{},
		Files:       []File{},
		Warnings:    []string{},
	}
	if err := ctx.Err(); err != nil {
		return directory, err
	}
	if depth > scan.maxDepth {
		directory.Warnings = append(directory.Warnings,
			fmt.Sprintf("%s: stopped at a depth of %d directories", directoryPath, scan.maxDepth))
		return directory, nil
	}
	if scan.alreadyVisited(directoryPath) {
		return directory, nil
	}
	directory.Warnings = append(directory.Warnings, scan.gitIgnore.load(directoryPath)...)

	entities, err := os.ReadDir(directoryPath)
	if err != nil {
		return directory, err
	}
	for _, entity := range entities {
		if err := ctx.Err(); err != nil {
			return directory, err
		}
		if err := scan.addEntry(ctx, &directory, entity, depth); err != nil {
			return directory, err
		}
	}
	return directory, nil
}

// alreadyVisited reports whether this directory has been walked before under
// another name. A symbolic link can point back up the tree, and following one
// twice would not end.
func (scan *scanner) alreadyVisited(directoryPath string) bool {
	resolved, err := filepath.EvalSymlinks(directoryPath)
	if err != nil {
		return false
	}
	if _, seen := scan.visited[resolved]; seen {
		return true
	}
	scan.visited[resolved] = struct{}{}
	return false
}

// addEntry places one entry in the directory: as a file, as a walked child, or
// nowhere at all. It returns an error only when the walk must stop.
func (scan *scanner) addEntry(
	ctx context.Context,
	directory *Directory,
	entity os.DirEntry,
	depth int,
) error {
	name := entity.Name()
	path := filepath.Join(directory.Path, name)

	isDirectory, err := scan.isDirectory(entity, path)
	if err != nil {
		directory.Warnings = append(directory.Warnings, fmt.Sprintf("%s: %v", path, err))
		return nil
	}
	if isDirectory {
		if _, skip := scan.ignored[name]; skip {
			return nil
		}
	}
	if scan.gitIgnore.matches(path, isDirectory) {
		return nil
	}
	if !isDirectory {
		if scan.matches(name) {
			directory.Files = append(directory.Files, File{Name: name, Path: path})
		}
		return nil
	}
	return scan.addChild(ctx, directory, path, depth)
}

// isDirectory answers for the entry itself, and for what a symbolic link points
// at: a link reports its own type, so the target has to be asked for
// separately.
func (scan *scanner) isDirectory(entity os.DirEntry, path string) (bool, error) {
	if entity.IsDir() {
		return true, nil
	}
	if entity.Type()&os.ModeSymlink == 0 {
		return false, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// addChild walks a subdirectory and keeps it only when it holds something. An
// empty branch would be noise in the tree.
func (scan *scanner) addChild(
	ctx context.Context,
	directory *Directory,
	path string,
	depth int,
) error {
	child, err := scan.walk(ctx, path, depth+1)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		directory.Warnings = append(directory.Warnings, fmt.Sprintf("%s: %v", path, err))
		return nil
	}
	directory.Warnings = append(directory.Warnings, child.Warnings...)
	if len(child.Directories) == 0 && len(child.Files) == 0 {
		return nil
	}
	directory.Directories = append(directory.Directories, child)
	return nil
}

func (scan *scanner) matches(name string) bool {
	extension := filepath.Ext(name)
	for _, candidate := range scan.extensions {
		if strings.EqualFold(extension, candidate) {
			return true
		}
	}
	return false
}
