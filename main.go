package main

import (
	"lazyrest/ui"
	"log"
	"os"
)

func main() {
	rootDirectoryPath, err := getRootDirectoryPath()
	if err != nil {
		log.Fatal(err)
	}

	err = ui.Run(rootDirectoryPath)
	if err != nil {
		log.Fatal(err)
	}

	// rootDirectoryPath, err := getRootDirectoryPath()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// directoriesWithFiles, err := finder.FindFilesInDirectory(
	// 	rootDirectoryPath,
	// 	".http",
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// if len(directoriesWithFiles.Directories) == 0 && len(directoriesWithFiles.Files) == 0 {
	// 	log.Fatal("No files for parse.")
	// }
	//
	// httpParser, err := http.NewParser()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(directoriesWithFiles.Path)
	// for _, directory := range directoriesWithFiles.Directories {
	// 	fmt.Println(directory.Path)
	// 	fmt.Println(len(directory.Files))
	// 	for _, directory := range directory.Directories {
	// 		fmt.Println(directory.Path)
	// 		fmt.Println(len(directory.Files))
	// 		for _, file := range directory.Files {
	// 			fmt.Println(file.Path)
	// 			suites, err := httpParser.GetSuitesFromFile(file.Path)
	// 			if err != nil {
	// 				log.Fatal(err)
	// 			}
	// 			println(len(suites))
	// 		}
	// 	}
	//
	// 	for _, file := range directory.Files {
	// 		fmt.Println(file.Path)
	// 		suites, err := httpParser.GetSuitesFromFile(file.Path)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		println(len(suites))
	// 	}
	// }
}

func getRootDirectoryPath() (string, error) {
	if len(os.Args) > 1 {
		directoryPath := os.Args[1]
		return directoryPath, nil
	}

	directoryPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return directoryPath, nil
}
