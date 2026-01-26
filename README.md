# YDRAG - YZMA DuckDB RAG

A Retrieval-Augmented Generation (RAG) system implemented in Go using:
- **YZMA** (hybridgroup/yzma) for local embedding generation via llama.cpp
- **DuckDB** as the vector database for similarity search

## Features

- Local embedding generation using any GGUF embedding model
- Vector similarity search using DuckDB's `array_cosine_similarity`
- Persistent storage of documents and embeddings
- Simple CLI interface for document management and querying
- **MCP Server** for integration with AI assistants (Claude, etc.)

## Prerequisites

1. **llama.cpp library**: Download or build llama.cpp and set the `YZMA_LIB` environment variable to the library path
2. **Embedding model**: Download a GGUF embedding model (e.g., nomic-embed-text, bge-small, all-MiniLM)

### Installing llama.cpp

```bash
# Using yzma CLI to download
go install github.com/hybridgroup/yzma/cmd/yzma@latest
yzma lib install

# Set the library path
export YZMA_LIB=/path/to/libllama.so  # Linux
export YZMA_LIB=/path/to/libllama.dylib  # macOS
```

### Recommended Embedding Models

- [nomic-embed-text-v1.5](https://huggingface.co/nomic-ai/nomic-embed-text-v1.5-GGUF)
- [bge-small-en-v1.5](https://huggingface.co/BAAI/bge-small-en-v1.5-GGUF)
- [all-MiniLM-L6-v2](https://huggingface.co/leliuga/all-MiniLM-L6-v2-GGUF)

## Building

```bash
go build -o ydrag .
```

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

Run as an MCP (Model Context Protocol) server for integration with AI assistants:

```bash
./ydrag -model ./models/nomic-embed-text-v1.5.Q8_0.gguf serve
```

The server exposes the following tools via stdio transport:
- `add_document` - Add a document to the knowledge base
- `query_documents` - Search for similar documents
- `list_documents` - List all documents
- `delete_document` - Delete a document

#### MCP Client Configuration

Add to your MCP client configuration (e.g., Claude Desktop):

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

## Options

| Flag | Description | Default |
|------|-------------|---------|
| `-model` | Path to GGUF embedding model (required) | - |
| `-lib` | Path to llama.cpp library | `$YZMA_LIB` |
| `-db` | Path to DuckDB database file | `rag.db` |
| `-context` | Context size for embeddings | 512 |
| `-batch` | Batch size for processing | 512 |
| `-verbose` | Enable verbose logging | false |

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      YDRAG                              │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │     CLI      │  │  MCP Server  │  │   RAG Core   │  │
│  │              │  │   (stdio)    │  │              │  │
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

1. **Document Ingestion**: Text is tokenized and passed through the embedding model to generate a vector representation
2. **Storage**: Documents and their embeddings are stored in DuckDB using the `FLOAT[]` array type
3. **Query**: The query text is embedded, then DuckDB's `array_cosine_similarity` function finds the most similar documents
4. **Retrieval**: Results are returned sorted by similarity score

## MCP Tools Reference

| Tool | Description | Parameters |
|------|-------------|------------|
| `add_document` | Add a document to the knowledge base | `id` (string), `content` (string) |
| `query_documents` | Search for similar documents | `query` (string), `top_k` (int, optional) |
| `list_documents` | List all documents | none |
| `delete_document` | Delete a document | `id` (string) |

## License

MIT
