package index

import (
    "os"
    "path/filepath"
    "testing"
)

func TestNewTantivyIndex(t *testing.T) {
    // Skip if tantivy library is not available
    if os.Getenv("CI") == "" {
        t.Skip("Skipping tantivy tests outside CI environment")
    }
    
    tmpDir := t.TempDir()
    indexPath := filepath.Join(tmpDir, "test-index")
    
    index, err := NewTantivyIndex(indexPath)
    if err != nil {
        t.Fatalf("Failed to create index: %v", err)
    }
    defer index.Close()
    
    if index.GetIndexedFilesCount() != 0 {
        t.Errorf("Expected 0 indexed files, got %d", index.GetIndexedFilesCount())
    }
}

func TestIndexAndSearch(t *testing.T) {
    // Skip if tantivy library is not available
    if os.Getenv("CI") == "" {
        t.Skip("Skipping tantivy tests outside CI environment")
    }
    
    tmpDir := t.TempDir()
    indexPath := filepath.Join(tmpDir, "test-index")
    vaultPath := filepath.Join(tmpDir, "vault")
    
    // Create test markdown file
    os.MkdirAll(vaultPath, 0755)
    testFile := filepath.Join(vaultPath, "test.md")
    content := `# Test Document

This is a test document for searching.
It contains some keywords like golang and tantivy.
`
    os.WriteFile(testFile, []byte(content), 0644)
    
    // Create and populate index
    index, err := NewTantivyIndex(indexPath)
    if err != nil {
        t.Fatalf("Failed to create index: %v", err)
    }
    defer index.Close()
    
    err = index.IndexDirectory(vaultPath, 1)
    if err != nil {
        t.Fatalf("Failed to index directory: %v", err)
    }
    
    // Test search
    results, err := index.Search("golang", 10)
    if err != nil {
        t.Fatalf("Search failed: %v", err)
    }
    
    if len(results) != 1 {
        t.Errorf("Expected 1 result, got %d", len(results))
    }
    
    if len(results) > 0 && results[0].FilePath != testFile {
        t.Errorf("Expected file path %s, got %s", testFile, results[0].FilePath)
    }
}