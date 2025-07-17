package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/mark3labs/mcp-go/server"
    "github.com/Atomzwieback/obsidian-mcp-search/internal/config"
    "github.com/Atomzwieback/obsidian-mcp-search/internal/index"
    "github.com/Atomzwieback/obsidian-mcp-search/internal/mcp"
)

func main() {
    // Konfiguration laden
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    if cfg.VaultPath == "" {
        log.Fatal("OBSIDIAN_VAULT_PATH environment variable must be set")
    }
    
    fmt.Printf("Starting Obsidian MCP Search Server...\n")
    fmt.Printf("Vault Path: %s\n", cfg.VaultPath)
    fmt.Printf("Index Path: %s\n", cfg.IndexPath)
    
    // Tantivy Index initialisieren
    tantivyIndex, err := index.NewTantivyIndex(cfg.IndexPath)
    if err != nil {
        log.Fatalf("Failed to initialize index: %v", err)
    }
    defer tantivyIndex.Close()
    
    // Initial indexing
    fmt.Println("Starting initial indexing...")
    if err := tantivyIndex.IndexDirectory(cfg.VaultPath, cfg.MaxWorkers); err != nil {
        log.Printf("Warning: Initial indexing error: %v", err)
    }
    
    // File Watcher starten
    if cfg.WatchFiles {
        watcher, err := index.NewFileWatcher(tantivyIndex, cfg.VaultPath)
        if err != nil {
            log.Printf("Failed to start file watcher: %v", err)
        } else {
            go watcher.Start()
            defer watcher.Stop()
        }
    }
    
    // MCP Server setup
    handler := mcp.NewSearchHandler(tantivyIndex)
    mcpServer := handler.SetupServer()
    
    // Graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        fmt.Println("\nShutting down server...")
        tantivyIndex.Close()
        os.Exit(0)
    }()
    
    // Server starten
    fmt.Println("MCP Server ready. Listening on stdio...")
    if err := server.ServeStdio(mcpServer); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}