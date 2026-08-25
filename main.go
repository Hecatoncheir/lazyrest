package main

import (
	"flag"
	"log"
	"os"

	appconfig "github.com/Hecatoncheir/lazyrest/config"
	"github.com/Hecatoncheir/lazyrest/environment"
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
	settings, configPath, err := appconfig.LoadDefault()
	if err != nil {
		log.Fatal(err)
	}

	err = ui.Run(rootDirectoryPath, ui.Config{
		Keybindings: settings.Keybindings,
		Locale:      settings.Locale,
		Theme:       settings.Theme,
		ConfigPath:  configPath,
		Runner: runner.Config{
			Timeout:          *timeout,
			MaxResponseBytes: *maxResponseBytes,
			HurlExecutable:   *hurlExecutable,
		},
		Environment: environment.Config{
			Name:        *environmentName,
			PublicFile:  *environmentFile,
			PrivateFile: *privateEnvironmentFile,
		},
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
