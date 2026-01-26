package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	configPath  = flag.String("config", "config.yaml", "path to configuration file")
	modelFile   = flag.String("model", "", "path to embedding model file (GGUF format)")
	libPath     = flag.String("lib", "", "path to llama.cpp library")
	dbPath      = flag.String("db", "", "path to DuckDB database file (use :memory: for in-memory)")
	contextSize = flag.Int("context", 0, "context size for embeddings")
	batchSize   = flag.Int("batch", 0, "batch size for processing")
	verbose     = flag.Bool("verbose", false, "enable verbose logging")
)

var cfg *Config

func main() {
	flag.Parse()

	var err error
	cfg, err = LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	applyFlagOverrides(cfg)

	if cfg.Model == "" {
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

	rag, err := NewRAGSystem(cfg.Model, cfg.LibPath, cfg.DBPath)
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

func applyFlagOverrides(cfg *Config) {
	if *modelFile != "" {
		cfg.Model = *modelFile
	}
	if *libPath != "" {
		cfg.LibPath = *libPath
	}
	if *dbPath != "" {
		cfg.DBPath = *dbPath
	}
	if *contextSize != 0 {
		cfg.ContextSize = *contextSize
	}
	if *batchSize != 0 {
		cfg.BatchSize = *batchSize
	}
	if *verbose {
		cfg.Verbose = true
	}
}

func showUsage() {
	fmt.Println("Usage: ydrag [options] <command> [args]")
	PrintCommandsHelp()
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nConfiguration priority: flags > env vars > config.yaml > defaults")
	fmt.Println("\nExample:")
	fmt.Println("  ydrag -model ./models/nomic-embed-text-v1.5.Q8_0.gguf add doc1 \"The capital of France is Paris\"")
	fmt.Println("  ydrag -model ./models/nomic-embed-text-v1.5.Q8_0.gguf query \"What is the capital of France?\"")
}
