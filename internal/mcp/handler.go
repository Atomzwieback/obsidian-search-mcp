package mcp

import (
    "encoding/json"
    "fmt"
    
    "github.com/mark3labs/mcp-go"
    "github.com/mark3labs/mcp-go/server"
    "github.com/Atomzwieback/obsidian-search-mcp/internal/index"
)

type SearchHandler struct {
    index *index.TantivyIndex
}

type SearchParams struct {
    Query string `json:"query"`
    Limit int    `json:"limit,omitempty"`
}

type SearchResponse struct {
    Results []index.SearchResult `json:"results"`
    Total   int                  `json:"total"`
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
        server.WithPromptSupport(),
    )
    
    // Search Tool
    searchTool := mcp.NewTool("search_obsidian",
        mcp.WithDescription("Search for content in Obsidian markdown files"),
        mcp.WithString("query", 
            mcp.Required(), 
            mcp.Description("Search query text")),
        mcp.WithNumber("limit", 
            mcp.Default(10.0),
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
        mcp.WithDescription("Current index status and statistics"),
    )
    
    s.AddResource(statusResource, h.handleStatus)
    
    return s
}

func (h *SearchHandler) handleSearch(arguments json.RawMessage) (*mcp.CallToolResult, error) {
    var params SearchParams
    if err := json.Unmarshal(arguments, &params); err != nil {
        return mcp.NewCallToolResult(nil, false), fmt.Errorf("invalid parameters: %w", err)
    }
    
    if params.Limit == 0 {
        params.Limit = 10
    }
    
    results, err := h.index.Search(params.Query, params.Limit)
    if err != nil {
        return mcp.NewCallToolResult(nil, false), fmt.Errorf("search failed: %w", err)
    }
    
    response := SearchResponse{
        Results: results,
        Total:   len(results),
    }
    
    // Format response for better readability
    var formattedResponse string
    formattedResponse = fmt.Sprintf("Found %d results for query '%s':\n\n", len(results), params.Query)
    
    for i, result := range results {
        formattedResponse += fmt.Sprintf("%d. %s (Score: %.2f)\n", i+1, result.FilePath, result.Score)
        formattedResponse += fmt.Sprintf("   Lines: %v\n", result.LineNumbers)
        formattedResponse += fmt.Sprintf("   Snippet:\n%s\n\n", result.Snippet)
    }
    
    return mcp.NewCallToolResult(
        []mcp.TextContent{
            {
                Type: "text",
                Text: formattedResponse,
            },
        },
        false,
    ), nil
}

func (h *SearchHandler) handleReindex(arguments json.RawMessage) (*mcp.CallToolResult, error) {
    // Reindexing logic hier implementieren
    return mcp.NewCallToolResult(
        []mcp.TextContent{
            {
                Type: "text", 
                Text: "Reindexing started...",
            },
        },
        false,
    ), nil
}

func (h *SearchHandler) handleStatus(arguments json.RawMessage) (any, error) {
    // Status Informationen sammeln
    status := map[string]interface{}{
        "status": "operational",
        "indexed_files": h.index.GetIndexedFilesCount(),
    }
    
    return status, nil
}