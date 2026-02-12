# YDRAG - YZMA DuckDB RAG

A Retrieval-Augmented Generation (RAG) system implemented in Go using:
- **YZMA** (hybridgroup/yzma) for local embedding generation via llama.cpp
- **DuckDB** as the vector database for similarity search
- **MCP** (Model Context Protocol) for AI assistant integration

## Features

- Local embedding generation using any GGUF embedding model
- Vector similarity search using DuckDB's `array_cosine_similarity`
- Persistent storage of documents and embeddings
- PDF text extraction for document ingestion
- Simple CLI interface for document management and querying
- MCP server with configurable transport (stdio, SSE, Streamable HTTP) for integration with AI assistants (Claude, Amp, etc.)
- Flexible configuration via YAML, environment variables, and CLI flags

## Prerequisites

- **Go 1.24+** (CGo enabled — see [Build Requirements](#build-requirements))
- **C compiler** (gcc or clang) — required by DuckDB's Go bindings
- **llama.cpp library** — download or build, then set `YZMA_LIB`
- **GGUF embedding model** — e.g., embeddinggemma, nomic-embed-text, bge-small

### Installing llama.cpp

```bash
# Using yzma CLI to download
go install github.com/hybridgroup/yzma/cmd/yzma@latest
yzma lib install

# Set the library path
export YZMA_LIB=/path/to/libllama.so    # Linux
export YZMA_LIB=/path/to/libllama.dylib  # macOS
```

### Recommended Embedding Models

- [embeddinggemma-300M](https://huggingface.co/ggml-org/embeddinggemma-300M-GGUF) — see [MODEL.md](MODEL.md) for setup details
- [nomic-embed-text-v1.5](https://huggingface.co/nomic-ai/nomic-embed-text-v1.5-GGUF)
- [bge-small-en-v1.5](https://huggingface.co/BAAI/bge-small-en-v1.5-GGUF)
- [all-MiniLM-L6-v2](https://huggingface.co/leliuga/all-MiniLM-L6-v2-GGUF)

## Build Requirements

DuckDB's Go driver uses CGo, so **`CGO_ENABLED=1`** is required. A C compiler (gcc/clang) must be available.

```bash
# Verify CGo is enabled (must print "1")
go env CGO_ENABLED

# If it prints "0", enable it for the current session:
export CGO_ENABLED=1

# Or set it permanently:
go env -w CGO_ENABLED=1
```

### Building

```bash
CGO_ENABLED=1 go build -o ydrag .
```

### Running Tests

```bash
CGO_ENABLED=1 go test ./... -v
```

The test suite covers configuration loading, command registry, vector math utilities, and CLI argument validation — all without requiring a model or llama.cpp library.

## Usage

### Add Documents

```bash
./ydrag -model ./models/nomic-embed-text-v1.5.Q8_0.gguf add doc1 "The capital of France is Paris"
./ydrag -model ./models/nomic-embed-text-v1.5.Q8_0.gguf add doc2 "Python is a programming language"
./ydrag -model ./models/nomic-embed-text-v1.5.Q8_0.gguf add doc3 "Berlin is the capital of Germany"
```

### Query Documents

```bash
./ydrag -model ./models/nomic-embed-text-v1.5.Q8_0.gguf query "What is the capital of France?"
```

Output:
```
Top 5 results for: "What is the capital of France?"

1. [0.8934] doc1: The capital of France is Paris
2. [0.7521] doc3: Berlin is the capital of Germany
3. [0.3412] doc2: Python is a programming language
```

### List Documents

```bash
./ydrag -model ./models/nomic-embed-text-v1.5.Q8_0.gguf list
```

### Delete Documents

```bash
./ydrag -model ./models/nomic-embed-text-v1.5.Q8_0.gguf delete doc1
```

### MCP Server Mode

Run as an MCP (Model Context Protocol) server for integration with AI assistants.
The transport is configurable: **stdio** (default), **sse**, or **streamable-http**.

```bash
# Default: stdio transport
./ydrag -model ./models/nomic-embed-text-v1.5.Q8_0.gguf serve

# SSE transport (listens on port 8080)
YDRAG_TRANSPORT=sse ./ydrag -model ./models/nomic-embed-text-v1.5.Q8_0.gguf serve

# Streamable HTTP transport (listens on port 8080)
YDRAG_TRANSPORT=streamable-http ./ydrag -model ./models/nomic-embed-text-v1.5.Q8_0.gguf serve
```

Or set the transport in `config.yaml`:

```yaml
server:
  transport: "sse"       # "stdio" | "sse" | "streamable-http"
  port: "8080"
```

The server exposes the following tools:
- `add_document` — Add a document to the knowledge base
- `query_documents` — Search for similar documents
- `list_documents` — List all documents
- `delete_document` — Delete a document

#### MCP Client Configuration

**stdio transport** — add to your MCP client config (e.g., Claude Desktop):

```json
{
  "mcpServers": {
    "ydrag": {
      "command": "/path/to/ydrag",
      "args": ["-model", "/path/to/model.gguf", "serve"]
    }
  }
}
```

**SSE / Streamable HTTP transport** — point your MCP client at the URL:

```json
{
  "mcpServers": {
    "ydrag": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

## Configuration

YDRAG supports configuration via YAML file, environment variables, and command-line flags.

**Priority order**: flags > environment variables > config.yaml > defaults

### Configuration File

Create a `config.yaml` in the working directory:

```yaml
model: "./models/nomic-embed-text-v1.5.Q8_0.gguf"
lib_path: "/path/to/libllama.so"
db_path: "rag.db"
context_size: 512
batch_size: 512
verbose: false
server:
  transport: "stdio"
  port: "8080"
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `YDRAG_MODEL` | Path to GGUF embedding model | — |
| `YZMA_LIB` | Path to llama.cpp library | — |
| `YDRAG_DB_PATH` | Path to DuckDB database file | `rag.db` |
| `YDRAG_CONTEXT_SIZE` | Context size for embeddings | `512` |
| `YDRAG_BATCH_SIZE` | Batch size for processing | `512` |
| `YDRAG_VERBOSE` | Enable verbose logging (`true`/`1`) | `false` |
| `YDRAG_TRANSPORT` | MCP transport type (stdio, sse, streamable-http) | `stdio` |
| `YDRAG_SERVER_PORT` | MCP server port | `8080` |

### Command-Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-config` | Path to configuration file | `config.yaml` |
| `-model` | Path to GGUF embedding model | — |
| `-lib` | Path to llama.cpp library | — |
| `-db` | Path to DuckDB database file | `rag.db` |
| `-context` | Context size for embeddings | `512` |
| `-batch` | Batch size for processing | `512` |
| `-verbose` | Enable verbose logging | `false` |

## Project Structure

```
.
├── main.go          # Entry point, flag parsing, orchestration
├── config.go        # Configuration loading (YAML, env, defaults)
├── command.go       # Command registry interface
├── cmd_add.go       # "add" command
├── cmd_delete.go    # "delete" command
├── cmd_list.go      # "list" command
├── cmd_query.go     # "query" command
├── cmd_serve.go     # "serve" command (MCP server)
├── rag.go           # RAG core: embeddings, DuckDB storage, search
├── readpdf.go       # PDF text extraction
├── mcp_server.go    # MCP server tool definitions and handlers
├── config.yaml      # Default configuration file
├── MODEL.md         # Embedding model setup guide
├── config_test.go   # Config loading and env override tests
├── command_test.go  # Command registry tests
├── rag_test.go      # Vector math and utility function tests
└── cmd_test.go      # CLI command argument validation tests
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      YDRAG                              │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │     CLI      │  │  MCP Server  │  │   RAG Core   │  │
│  │              │  │(stdio/sse/http)│ │              │  │
│  │  add/query/  │  │              │  │  • Embed     │  │
│  │  list/delete │  │  4 tools     │  │  • Store     │  │
│  └──────┬───────┘  └──────┬───────┘  │  • Search    │  │
│         │                 │          └──────┬───────┘  │
│         └─────────────────┴─────────────────┘          │
│                                                         │
│  ┌─────────────────┐         ┌─────────────────────┐   │
│  │  YZMA/llama.cpp │         │      DuckDB         │   │
│  │                 │         │                     │   │
│  │  • Load GGUF    │         │  • Store documents  │   │
│  │  • Tokenize     │         │  • Store embeddings │   │
│  │  • Embed text   │         │  • Vector search    │   │
│  └────────┬────────┘         └──────────┬──────────┘   │
│           │      []float32              │              │
│           └─────────────────────────────┘              │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## How It Works

1. **Document Ingestion** — Text is tokenized and passed through the embedding model to generate a vector representation
2. **Storage** — Documents and their embeddings are stored in DuckDB using the `FLOAT[]` array type
3. **Query** — The query text is embedded, then DuckDB's `array_cosine_similarity` function finds the most similar documents
4. **Retrieval** — Results are returned sorted by similarity score

## Dependencies

| Package | Purpose |
|---------|---------|
| [hybridgroup/yzma](https://github.com/hybridgroup/yzma) | llama.cpp Go bindings for embedding generation |
| [marcboeker/go-duckdb/v2](https://github.com/marcboeker/go-duckdb) | DuckDB Go driver |
| [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) | MCP server implementation |
| [ledongthuc/pdf](https://github.com/ledongthuc/pdf) | PDF text extraction |
| [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3) | YAML configuration parsing |

## MCP Tools Reference

| Tool | Description | Parameters |
|------|-------------|------------|
| `add_document` | Add a document to the knowledge base | `id` (string, required), `content` (string, required) |
| `query_documents` | Search for similar documents | `query` (string, required), `top_k` (int, default: 5) |
| `list_documents` | List all documents | none |
| `delete_document` | Delete a document | `id` (string, required) |

## License

MIT
