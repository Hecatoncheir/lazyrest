package finder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ignoredDirectories = map[string]struct{}{
	".git":         {},
	".hg":          {},
	".svn":         {},
	".cache":       {},
	"node_modules": {},
	"vendor":       {},
}

func FindFilesInDirectory(directoryPath string, extensions []string) (Directory, error) {
	directoryPath = filepath.Clean(directoryPath)
	directoryName := filepath.Base(directoryPath)

	directory := Directory{
		Name:        directoryName,
		Path:        directoryPath,
		Directories: []Directory{},
		Files:       []File{},
		Warnings:    []string{},
	}

	entities, err := os.ReadDir(directoryPath)
	if err != nil {
		return directory, err
	}

	for _, entity := range entities {
		entityName := entity.Name()
		entityPath := filepath.Join(directory.Path, entityName)

		if entity.IsDir() {
			if _, ignored := ignoredDirectories[entityName]; ignored {
				continue
			}
			directoryWithFiles, err := FindFilesInDirectory(entityPath, extensions)
			if err != nil {
				directory.Warnings = append(directory.Warnings, fmt.Sprintf("%s: %v", entityPath, err))
				continue
			}
			directory.Warnings = append(directory.Warnings, directoryWithFiles.Warnings...)
			if len(directoryWithFiles.Directories) == 0 && len(directoryWithFiles.Files) == 0 {
				continue
			}
			directory.Directories = append(directory.Directories, directoryWithFiles)
		} else {
			entityExtension := filepath.Ext(entityName)
			matched := false
			for _, ext := range extensions {
				if strings.EqualFold(entityExtension, ext) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
			file := File{
				Name: entityName,
				Path: entityPath,
			}
			directory.Files = append(directory.Files, file)
		}
	}

	return directory, nil
}
