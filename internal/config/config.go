package config

import (
    "os"
    "path/filepath"
)

type Config struct {
    VaultPath   string
    IndexPath   string
    MaxWorkers  int
    WatchFiles  bool
}

func LoadConfig() (*Config, error) {
    homeDir, _ := os.UserHomeDir()
    defaultIndexPath := filepath.Join(homeDir, ".obsidian-mcp", "index")
    
    return &Config{
        VaultPath:  os.Getenv("OBSIDIAN_VAULT_PATH"),
        IndexPath:  getEnvOrDefault("MCP_INDEX_PATH", defaultIndexPath),
        MaxWorkers: 4,
        WatchFiles: true,
    }, nil
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}