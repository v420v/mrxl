package main

import (
	"flag"
	"log"
	"os"

	"github.com/v420v/mrxl/internal/gen"
	"github.com/v420v/mrxl/internal/parser"
)

func main() {
	var srcFile string
	var outFile string
	flag.StringVar(&srcFile, "src", "", "input file")
	flag.StringVar(&outFile, "out", "mermaid.out.xlsx", "output file")
	flag.Parse()

	src, err := os.ReadFile(srcFile)
	if err != nil {
		log.Fatalf("read input file: %v", err)
		os.Exit(1)
	}
	parser, err := parser.NewParser()
	if err != nil {
		log.Fatalf("parse error: %v", err)
		os.Exit(1)
	}
	diagram, err := parser.Parse(string(src))
	if err != nil {
		log.Fatalf("parse error: %v", err)
		os.Exit(1)
	}

	if err := gen.Generate(diagram, outFile); err != nil {
		log.Fatalf("generate error: %v", err)
		os.Exit(1)
	}
}
