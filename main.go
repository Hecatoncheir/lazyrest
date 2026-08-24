package main

import (
	"flag"
	"log"
	"os"

	"github.com/Hecatoncheir/lazyrest/environment"
	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
	"github.com/Hecatoncheir/lazyrest/ui"
)

func main() {
	timeout := flag.Duration("timeout", runner.DefaultTimeout, "request and Hurl execution timeout")
	maxResponseBytes := flag.Int64("max-response-bytes", runner.DefaultMaxResponseBytes, "maximum response bytes kept in memory")
	hurlExecutable := flag.String("hurl", "hurl", "path or name of the Hurl executable")
	environmentName := flag.String("env", "", "environment profile from the HTTP client environment files")
	environmentFile := flag.String("env-file", environment.DefaultPublicFile, "public HTTP client environment file")
	privateEnvironmentFile := flag.String("private-env-file", environment.DefaultPrivateFile, "private HTTP client environment file")
	flag.Parse()

	rootDirectoryPath, err := getRootDirectoryPath()
	if err != nil {
		log.Fatal(err)
	}

	selectedEnvironment, err := environment.Load(rootDirectoryPath, environment.Config{
		Name:        *environmentName,
		PublicFile:  *environmentFile,
		PrivateFile: *privateEnvironmentFile,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = ui.Run(rootDirectoryPath, ui.Config{
		Runner: runner.Config{
			Timeout:          *timeout,
			MaxResponseBytes: *maxResponseBytes,
			HurlExecutable:   *hurlExecutable,
		},
		ParseOptions: parserhttp.ParseOptions{
			Variables:       selectedEnvironment.Values,
			SecretVariables: selectedEnvironment.SecretVariables,
		},
		EnvironmentName: selectedEnvironment.Name,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func getRootDirectoryPath() (string, error) {
	if flag.NArg() > 0 {
		directoryPath := flag.Arg(0)
		return directoryPath, nil
	}

	directoryPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return directoryPath, nil
}
