package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AddDocumentArgs contains the parameters for adding a document to the knowledge base.
type AddDocumentArgs struct {
	ID      string `json:"id" jsonschema:"required,Unique document identifier"`
	Content string `json:"content" jsonschema:"required,Document content text"`
}

// AddDocumentResult is the response returned after adding a document.
type AddDocumentResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// QueryDocumentsArgs contains the parameters for querying documents by vector similarity.
type QueryDocumentsArgs struct {
	Query string `json:"query" jsonschema:"required,Search query text"`
	TopK  int    `json:"top_k" jsonschema:"Maximum number of results to return (default: 5)"`
}

// QueryResult represents a single document match from a similarity search.
type QueryResult struct {
	ID      string  `json:"id"`
	Content string  `json:"content"`
	Score   float64 `json:"score"`
}

// QueryDocumentsResult is the response returned from a document query, containing matched results.
type QueryDocumentsResult struct {
	Results []QueryResult `json:"results"`
	Count   int           `json:"count"`
}

// ListDocumentsArgs contains the parameters for listing documents (currently empty).
type ListDocumentsArgs struct{}

// DocumentItem represents a document entry with its ID and content.
type DocumentItem struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

// ListDocumentsResult is the response returned when listing all documents in the knowledge base.
type ListDocumentsResult struct {
	Documents []DocumentItem `json:"documents"`
	Total     int            `json:"total"`
}

// DeleteDocumentArgs contains the parameters for deleting a document from the knowledge base.
type DeleteDocumentArgs struct {
	ID string `json:"id" jsonschema:"required,Document identifier to delete"`
}

// DeleteDocumentResult is the response returned after deleting a document.
type DeleteDocumentResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// MCPServer wraps a RAG system and exposes it as an MCP server with tool-based document operations.
type MCPServer struct {
	rag    *RAGSystem
	server *mcp.Server
}

// NewMCPServer creates a new MCPServer that serves the given RAG system and registers all tools.
func NewMCPServer(rag *RAGSystem) *MCPServer {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "ydrag",
		Version: "1.0.0",
	}, nil)

	m := &MCPServer{
		rag:    rag,
		server: server,
	}

	m.registerTools()

	return m
}

// registerTools registers all MCP tools (add, query, list, delete) on the server.
func (m *MCPServer) registerTools() {
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "add_document",
		Description: "Add a document to the RAG knowledge base with embeddings generated automatically",
	}, m.addDocument)

	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "query_documents",
		Description: "Search the knowledge base for documents similar to the query text using vector similarity",
	}, m.queryDocuments)

	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "list_documents",
		Description: "List all documents in the knowledge base",
	}, m.listDocuments)

	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "delete_document",
		Description: "Delete a document from the knowledge base",
	}, m.deleteDocument)
}

// addDocument handles the add_document tool call, validating inputs and storing the document with its embedding.
func (m *MCPServer) addDocument(ctx context.Context, req *mcp.CallToolRequest, args AddDocumentArgs) (*mcp.CallToolResult, AddDocumentResult, error) {
	if args.ID == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: document ID is required"}},
			IsError: true,
		}, AddDocumentResult{Success: false, Message: "document ID is required"}, nil
	}
	if args.Content == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: document content is required"}},
			IsError: true,
		}, AddDocumentResult{Success: false, Message: "document content is required"}, nil
	}

	if err := m.rag.AddDocument(args.ID, args.Content); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error adding document: %v", err)}},
			IsError: true,
		}, AddDocumentResult{Success: false, Message: err.Error()}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Document '%s' added successfully", args.ID)}},
	}, AddDocumentResult{Success: true, Message: fmt.Sprintf("Document '%s' added successfully", args.ID)}, nil
}

// queryDocuments handles the query_documents tool call, performing vector similarity search and returning ranked results.
func (m *MCPServer) queryDocuments(ctx context.Context, req *mcp.CallToolRequest, args QueryDocumentsArgs) (*mcp.CallToolResult, QueryDocumentsResult, error) {
	if args.Query == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: query text is required"}},
			IsError: true,
		}, QueryDocumentsResult{}, nil
	}

	topK := args.TopK
	if topK <= 0 {
		topK = 5
	}

	results, err := m.rag.Query(args.Query, topK)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error querying documents: %v", err)}},
			IsError: true,
		}, QueryDocumentsResult{}, nil
	}

	queryResults := make([]QueryResult, len(results))
	for i, r := range results {
		queryResults[i] = QueryResult{
			ID:      r.ID,
			Content: r.Content,
			Score:   r.Score,
		}
	}

	if len(results) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No matching documents found"}},
		}, QueryDocumentsResult{Results: queryResults, Count: 0}, nil
	}

	var text string
	for i, r := range results {
		text += fmt.Sprintf("%d. [%.4f] %s: %s\n", i+1, r.Score, r.ID, truncate(r.Content, 100))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, QueryDocumentsResult{Results: queryResults, Count: len(queryResults)}, nil
}

// listDocuments handles the list_documents tool call, returning all documents in the knowledge base.
func (m *MCPServer) listDocuments(ctx context.Context, req *mcp.CallToolRequest, args ListDocumentsArgs) (*mcp.CallToolResult, ListDocumentsResult, error) {
	docs, err := m.rag.ListDocuments()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error listing documents: %v", err)}},
			IsError: true,
		}, ListDocumentsResult{}, nil
	}

	items := make([]DocumentItem, len(docs))
	for i, d := range docs {
		items[i] = DocumentItem{
			ID:      d.ID,
			Content: d.Content,
		}
	}

	if len(docs) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No documents in knowledge base"}},
		}, ListDocumentsResult{Documents: items, Total: 0}, nil
	}

	var text string
	for _, d := range docs {
		text += fmt.Sprintf("  %s: %s\n", d.ID, truncate(d.Content, 80))
	}
	text = fmt.Sprintf("Documents in knowledge base (%d total):\n%s", len(docs), text)

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, ListDocumentsResult{Documents: items, Total: len(items)}, nil
}

// deleteDocument handles the delete_document tool call, removing a document by ID from the knowledge base.
func (m *MCPServer) deleteDocument(ctx context.Context, req *mcp.CallToolRequest, args DeleteDocumentArgs) (*mcp.CallToolResult, DeleteDocumentResult, error) {
	if args.ID == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Error: document ID is required"}},
			IsError: true,
		}, DeleteDocumentResult{Success: false, Message: "document ID is required"}, nil
	}

	if err := m.rag.DeleteDocument(args.ID); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error deleting document: %v", err)}},
			IsError: true,
		}, DeleteDocumentResult{Success: false, Message: err.Error()}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Document '%s' deleted successfully", args.ID)}},
	}, DeleteDocumentResult{Success: true, Message: fmt.Sprintf("Document '%s' deleted successfully", args.ID)}, nil
}

// Run starts the MCP server using the specified transport ("stdio", "sse", or "streamable-http") and address.
func (m *MCPServer) Run(ctx context.Context, transport, addr string) error {
	switch transport {
	case "stdio", "":
		return m.server.Run(ctx, &mcp.StdioTransport{})
	case "sse":
		handler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
			return m.server
		}, nil)
		fmt.Fprintf(os.Stderr, "MCP SSE server listening on %s\n", addr)
		srv := &http.Server{Addr: addr, Handler: handler}
		go func() {
			<-ctx.Done()
			srv.Close()
		}()
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	case "streamable-http":
		handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
			return m.server
		}, nil)
		fmt.Fprintf(os.Stderr, "MCP Streamable HTTP server listening on %s\n", addr)
		srv := &http.Server{Addr: addr, Handler: handler}
		go func() {
			<-ctx.Done()
			srv.Close()
		}()
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported transport: %q (use stdio, sse, or streamable-http)", transport)
	}
}
