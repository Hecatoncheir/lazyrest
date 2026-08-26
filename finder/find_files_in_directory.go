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
		maxDepth:   maxDepth,
		visited:    map[string]struct{}{},
	}
	return scan.walk(ctx, filepath.Clean(directoryPath), 0)
}

type scanner struct {
	extensions []string
	ignored    map[string]struct{}
	maxDepth   int
	// visited holds the directories already walked, resolved through any
	// symbolic link, so that a link back into the tree is not followed twice.
	visited map[string]struct{}
}

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
		directory.Warnings = append(directory.Warnings, fmt.Sprintf("%s: stopped at a depth of %d directories", directoryPath, scan.maxDepth))
		return directory, nil
	}
	if resolved, err := filepath.EvalSymlinks(directoryPath); err == nil {
		if _, seen := scan.visited[resolved]; seen {
			return directory, nil
		}
		scan.visited[resolved] = struct{}{}
	}

	entities, err := os.ReadDir(directoryPath)
	if err != nil {
		return directory, err
	}

	for _, entity := range entities {
		if err := ctx.Err(); err != nil {
			return directory, err
		}
		entityName := entity.Name()
		entityPath := filepath.Join(directory.Path, entityName)

		isDirectory := entity.IsDir()
		if !isDirectory && entity.Type()&os.ModeSymlink != 0 {
			// A symbolic link reports its own type, so what it points at has
			// to be asked for separately.
			info, err := os.Stat(entityPath)
			if err != nil {
				directory.Warnings = append(directory.Warnings, fmt.Sprintf("%s: %v", entityPath, err))
				continue
			}
			isDirectory = info.IsDir()
		}

		if !isDirectory {
			if scan.matches(entityName) {
				directory.Files = append(directory.Files, File{Name: entityName, Path: entityPath})
			}
			continue
		}
		if _, skip := scan.ignored[entityName]; skip {
			continue
		}

		child, err := scan.walk(ctx, entityPath, depth+1)
		if err != nil {
			if ctx.Err() != nil {
				return directory, ctx.Err()
			}
			directory.Warnings = append(directory.Warnings, fmt.Sprintf("%s: %v", entityPath, err))
			continue
		}
		directory.Warnings = append(directory.Warnings, child.Warnings...)
		if len(child.Directories) == 0 && len(child.Files) == 0 {
			continue
		}
		directory.Directories = append(directory.Directories, child)
	}

	return directory, nil
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
