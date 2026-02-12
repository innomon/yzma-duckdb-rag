// Package main implements ydrag, a Retrieval-Augmented Generation (RAG) system
// that uses local embedding models via YZMA/llama.cpp and stores documents with
// their vector embeddings in DuckDB.
//
// # Architecture
//
// YDRAG has three main components:
//
//   - CLI — a set of subcommands (add, query, list, delete, serve) for managing
//     documents and running the server.
//   - RAG core — handles embedding generation through YZMA/llama.cpp, document
//     storage in DuckDB, and cosine-similarity vector search.
//   - MCP server — exposes the RAG system as a Model Context Protocol server
//     with configurable transports (stdio, SSE, Streamable HTTP) for integration
//     with AI assistants such as Claude and Amp.
//
// # Configuration
//
// Configuration is loaded with the following priority (highest first):
//
//   - Command-line flags (-model, -db, -lib, etc.)
//   - Environment variables (YDRAG_MODEL, YDRAG_DB_PATH, YZMA_LIB, etc.)
//   - YAML configuration file (config.yaml by default)
//   - Built-in defaults
//
// # Usage
//
// Build and run:
//
//	CGO_ENABLED=1 go build -o ydrag .
//	./ydrag -model ./models/model.gguf add doc1 "The capital of France is Paris"
//	./ydrag -model ./models/model.gguf query "What is the capital of France?"
//	./ydrag -model ./models/model.gguf serve
//
// See the README for full documentation on embedding models, transports, and
// MCP client configuration.
package main
