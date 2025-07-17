package mcp

import (
    "testing"
)

func TestNewSearchHandler(t *testing.T) {
    // This test doesn't require tantivy library
    handler := NewSearchHandler(nil)
    if handler == nil {
        t.Fatal("Expected handler to be created")
    }
    
    if handler.index != nil {
        t.Error("Expected index to be nil")
    }
}

func TestSetupServer(t *testing.T) {
    handler := NewSearchHandler(nil)
    server := handler.SetupServer()
    
    if server == nil {
        t.Fatal("Expected server to be created")
    }
    
    // Verify server has the expected tools
    // Note: mcp-go doesn't expose a way to check registered tools directly
    // so we just verify the server was created successfully
}