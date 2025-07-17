package mcp

import (
    "context"
    "fmt"
    
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
    "github.com/Atomzwieback/obsidian-search-mcp/internal/index"
)

type SearchHandler struct {
    index *index.TantivyIndex
}

func NewSearchHandler(tantivyIndex *index.TantivyIndex) *SearchHandler {
    return &SearchHandler{
        index: tantivyIndex,
    }
}

func (h *SearchHandler) SetupServer() *server.MCPServer {
    s := server.NewMCPServer(
        "Obsidian Search Server",
        "1.0.0",
        server.WithToolCapabilities(false),
    )
    
    // Search Tool
    searchTool := mcp.NewTool("search_vault",
        mcp.WithDescription("Search for content in Obsidian vault markdown files"),
        mcp.WithString("query", 
            mcp.Required(), 
            mcp.Description("Search query text")),
        mcp.WithNumber("limit", 
            mcp.Description("Maximum number of results to return")),
    )
    
    s.AddTool(searchTool, h.handleSearch)
    
    // Reindex Tool
    reindexTool := mcp.NewTool("reindex_vault",
        mcp.WithDescription("Force reindex of the entire Obsidian vault"),
    )
    
    s.AddTool(reindexTool, h.handleReindex)
    
    // Status Resource
    statusResource := mcp.NewResource(
        "index_status",
        "text/plain",
    )
    
    s.AddResource(statusResource, h.handleStatus)
    
    return s
}

func (h *SearchHandler) handleSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Get parameters
    query, err := request.RequireString("query")
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("Invalid query parameter: %v", err)), nil
    }
    
    limit := 10
    // Try to get limit parameter
    args := request.GetArguments()
    if limitArg, exists := args["limit"]; exists {
        if limitFloat, ok := limitArg.(float64); ok && limitFloat > 0 {
            limit = int(limitFloat)
        }
    }
    
    // Perform search
    results, err := h.index.Search(query, limit)
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
    }
    
    // Format response
    var formattedResponse string
    formattedResponse = fmt.Sprintf("Found %d results for query '%s':\n\n", len(results), query)
    
    for i, result := range results {
        formattedResponse += fmt.Sprintf("%d. %s (Score: %.2f)\n", i+1, result.FilePath, result.Score)
        formattedResponse += fmt.Sprintf("   Lines: %v\n", result.LineNumbers)
        formattedResponse += fmt.Sprintf("   Snippet:\n%s\n\n", result.Snippet)
    }
    
    return mcp.NewToolResultText(formattedResponse), nil
}

func (h *SearchHandler) handleReindex(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // TODO: Implement reindexing logic
    return mcp.NewToolResultText("Reindexing started..."), nil
}

func (h *SearchHandler) handleStatus(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
    // Get status information
    indexedFiles := h.index.GetIndexedFilesCount()
    
    statusText := fmt.Sprintf("Index Status:\n"+
        "- Status: operational\n"+
        "- Indexed files: %d\n", indexedFiles)
    
    return []mcp.ResourceContents{
        &mcp.TextResourceContents{
            URI:      "index_status",
            MIMEType: "text/plain",
            Text:     statusText,
        },
    }, nil
}