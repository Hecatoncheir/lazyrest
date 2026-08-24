package main

import (
	"flag"
	"github.com/Hecatoncheir/lazyrest/runner"
	"github.com/Hecatoncheir/lazyrest/ui"
	"log"
	"os"
)

func main() {
	timeout := flag.Duration("timeout", runner.DefaultTimeout, "request and Hurl execution timeout")
	maxResponseBytes := flag.Int64("max-response-bytes", runner.DefaultMaxResponseBytes, "maximum response bytes kept in memory")
	hurlExecutable := flag.String("hurl", "hurl", "path or name of the Hurl executable")
	flag.Parse()

	rootDirectoryPath, err := getRootDirectoryPath()
	if err != nil {
		log.Fatal(err)
	}

	err = ui.Run(rootDirectoryPath, runner.Config{
		Timeout:          *timeout,
		MaxResponseBytes: *maxResponseBytes,
		HurlExecutable:   *hurlExecutable,
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
