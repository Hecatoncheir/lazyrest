package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	appconfig "github.com/Hecatoncheir/lazyrest/config"
	"github.com/Hecatoncheir/lazyrest/environment"
	"github.com/Hecatoncheir/lazyrest/runner"
	"github.com/Hecatoncheir/lazyrest/ui"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func run(arguments []string, output io.Writer) error {
	flags := flag.NewFlagSet("lazyrest", flag.ContinueOnError)
	flags.SetOutput(output)
	timeout := flags.Duration("timeout", runner.DefaultTimeout, "request and Hurl execution timeout")
	maxResponseBytes := flags.Int64("max-response-bytes", runner.DefaultMaxResponseBytes, "maximum response bytes kept in memory")
	hurlExecutable := flags.String("hurl", "hurl", "path or name of the Hurl executable")
	environmentName := flags.String("env", "", "environment profile from the HTTP client environment files")
	environmentFile := flags.String("env-file", environment.DefaultPublicFile, "public HTTP client environment file")
	privateEnvironmentFile := flags.String("private-env-file", environment.DefaultPrivateFile, "private HTTP client environment file")
	configFile := flags.String("config", "", "additional configuration file with highest priority")
	generateConfig := flags.Bool("generate-config", false, "create a default configuration file")
	printConfig := flags.Bool("print-config", false, "print the resolved configuration and exit")
	validateConfig := flags.Bool("validate-config", false, "validate the resolved configuration and exit")
	if err := flags.Parse(arguments); err != nil {
		return err
	}

	rootDirectoryPath, err := getRootDirectoryPath(flags.Args())
	if err != nil {
		return err
	}
	userConfigPath, err := appconfig.DefaultPath()
	if err != nil {
		return err
	}
	if *generateConfig {
		target := userConfigPath
		if *configFile != "" {
			target = *configFile
		}
		if err := appconfig.Generate(target); err != nil {
			return err
		}
		_, err := fmt.Fprintf(output, "created %s\n", target)
		return err
	}

	configPaths := []string{userConfigPath, appconfig.ProjectPath(rootDirectoryPath)}
	if *configFile != "" {
		if _, err := os.Stat(*configFile); err != nil {
			return fmt.Errorf("read explicit config %s: %w", *configFile, err)
		}
		configPaths = append(configPaths, *configFile)
	}
	settings, err := appconfig.LoadFiles(configPaths)
	if err != nil {
		return err
	}
	if *validateConfig {
		_, err := fmt.Fprintln(output, "configuration is valid")
		return err
	}
	if *printConfig {
		contents, err := appconfig.Marshal(settings.Document)
		if err != nil {
			return err
		}
		_, err = output.Write(contents)
		return err
	}
	historyPath, err := appconfig.HistoryPath()
	if err != nil {
		return err
	}

	return ui.Run(rootDirectoryPath, ui.Config{
		Keybindings: settings.Keybindings,
		Locale:      settings.Locale,
		Theme:       settings.Theme,
		ConfigPath:  userConfigPath,
		ConfigPaths: configPaths,
		HistoryPath: historyPath,
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
}

func getRootDirectoryPath(arguments []string) (string, error) {
	if len(arguments) > 0 {
		return arguments[0], nil
	}
	return os.Getwd()
}
