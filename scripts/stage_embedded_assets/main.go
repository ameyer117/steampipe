package main

import (
	"context"
	"flag"
	"log"

	spociinstaller "github.com/turbot/steampipe/pkg/ociinstaller"
)

func main() {
	targetOS := flag.String("os", "", "target operating system")
	targetArch := flag.String("arch", "", "target architecture")
	outputRoot := flag.String("output", "", "output directory for bundled assets")
	flag.Parse()

	if *targetOS == "" || *targetArch == "" || *outputRoot == "" {
		log.Fatal("os, arch and output are required")
	}

	if err := spociinstaller.StageBundledAssets(context.Background(), *targetOS, *targetArch, *outputRoot); err != nil {
		log.Fatal(err)
	}
}
