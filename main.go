package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	modelFile   = flag.String("model", "", "path to embedding model file (GGUF format)")
	libPath     = flag.String("lib", os.Getenv("YZMA_LIB"), "path to llama.cpp library")
	dbPath      = flag.String("db", "rag.db", "path to DuckDB database file (use :memory: for in-memory)")
	contextSize = flag.Int("context", 512, "context size for embeddings")
	batchSize   = flag.Int("batch", 512, "batch size for processing")
	verbose     = flag.Bool("verbose", false, "enable verbose logging")
)

func main() {
	flag.Parse()

	if *modelFile == "" {
		showUsage()
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Please specify a command")
		PrintCommandsHelp()
		os.Exit(1)
	}

	cmd, ok := GetCommand(args[0])
	if !ok {
		fmt.Printf("Unknown command: %s\n", args[0])
		PrintCommandsHelp()
		os.Exit(1)
	}

	rag, err := NewRAGSystem(*modelFile, *libPath, *dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing RAG system: %v\n", err)
		os.Exit(1)
	}
	defer rag.Close()

	if err := cmd.Run(rag, args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Println("Usage: ydrag [options] <command> [args]")
	PrintCommandsHelp()
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nExample:")
	fmt.Println("  ydrag -model ./models/nomic-embed-text-v1.5.Q8_0.gguf add doc1 \"The capital of France is Paris\"")
	fmt.Println("  ydrag -model ./models/nomic-embed-text-v1.5.Q8_0.gguf query \"What is the capital of France?\"")
}
