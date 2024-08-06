package finder

import (
	"os"
	"path/filepath"
)

func FindFilesInDirectory(directoryPath, extension string) (Directory, error) {
	_, directoryName := filepath.Split(directoryPath)

	directory := Directory{
		Name:        directoryName,
		Path:        directoryPath,
		Directories: []Directory{},
		Files:       []File{},
	}

	entities, err := os.ReadDir(directoryPath)
	if err != nil {
		return directory, err
	}

	for _, entity := range entities {
		entityName := entity.Name()
		entityPath := filepath.Join(directory.Path, entityName)

		if entity.IsDir() {
			directoryWithFiles, err := FindFilesInDirectory(entityPath, extension)
			if err != nil {
				continue
			}
			if len(directoryWithFiles.Directories) == 0 && len(directoryWithFiles.Files) == 0 {
				continue
			}
			directory.Directories = append(directory.Directories, directoryWithFiles)
		} else {
			entityExtension := filepath.Ext(entityName)
			if entityExtension != extension {
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
