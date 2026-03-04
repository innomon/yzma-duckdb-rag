# ydrag Skill

A Retrieval-Augmented Generation (RAG) system implemented in Go using YZMA (llama.cpp) for local embeddings and DuckDB as a vector database.

## Capabilities

- **Local Embeddings**: Generates vector embeddings for text using GGUF models via llama.cpp.
- **Vector Search**: Performs similarity search using DuckDB's `array_cosine_similarity`.
- **Document Management**: Add, list, and delete documents with persistent storage.
- **MCP Server**: Exposes RAG operations as Model Context Protocol (MCP) tools.
- **PDF Support**: Extract text from PDF files for ingestion.

## Core Components

- `rag.go`: Core RAG logic including embedding generation and DuckDB operations.
- `mcp_server.go`: MCP server implementation and tool definitions.
- `main.go`: CLI entry point and orchestration.

## Usage Guide

### CLI Commands

The `ydrag` binary supports several subcommands:

- `add <id> <content>`: Add a document to the knowledge base.
- `query <text> [--top-k <n>]`: Search for documents similar to the input text.
- `list`: List all stored documents.
- `delete <id>`: Remove a document by its ID.
- `serve`: Start the MCP server.

### MCP Tools

When running in `serve` mode, the following tools are available to AI assistants:

- `add_document(id, content)`: Ingests text into the RAG system.
- `query_documents(query, top_k)`: Searches for relevant context.
- `list_documents()`: Shows all ingested documents.
- `delete_document(id)`: Removes context from the system.

## Configuration

YDRAG can be configured via `config.yaml`, environment variables, or CLI flags.

For detailed instructions on downloading and setting up embedding models (like **EmbeddingGemma**), refer to [MODEL.md](../../../MODEL.md).

### Key Environment Variables

- `YDRAG_MODEL`: Path to the GGUF embedding model.
- `YZMA_LIB`: Path to the `libllama` shared library (required).
- `YDRAG_DB_PATH`: Path to the DuckDB database file (default: `rag.db`).
- `YDRAG_TRANSPORT`: MCP transport type (`stdio`, `sse`, or `streamable-http`).

## Development and Build

- **Requirements**: Go 1.24+, C compiler (gcc/clang).
- **Build**: Must be built with `CGO_ENABLED=1`.
  ```bash
  CGO_ENABLED=1 go build -o ydrag .
  ```
- **Tests**: 
  ```bash
  CGO_ENABLED=1 go test ./...
  ```

## Best Practices for Agents

- Use `query_documents` to find relevant information before answering user questions if a knowledge base is available.
- When adding documents, use descriptive and unique IDs to avoid collisions.
- Ensure `YZMA_LIB` is correctly set in the environment before attempting to run the system.
- For large documents, consider chunking text before calling `add_document`.
