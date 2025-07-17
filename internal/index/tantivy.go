package index

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"
    
    tantivy "github.com/anyproto/tantivy-go"
    "github.com/karrick/godirwalk"
)

type SearchResult struct {
    FilePath    string   `json:"file_path"`
    Snippet     string   `json:"snippet"`
    Score       float32  `json:"score"`
    LineNumbers []int    `json:"line_numbers"`
}

type TantivyIndex struct {
    context     *tantivy.TantivyContext
    indexPath   string
    mu          sync.RWMutex
    lastIndexed map[string]time.Time
}

func NewTantivyIndex(indexPath string) (*TantivyIndex, error) {
    // Initialize tantivy library
    err := tantivy.LibInit(false, false, "info")
    if err != nil {
        return nil, fmt.Errorf("failed to initialize tantivy: %w", err)
    }
    
    // Create schema
    builder, err := tantivy.NewSchemaBuilder()
    if err != nil {
        return nil, fmt.Errorf("failed to create schema builder: %w", err)
    }
    
    // Add fields to schema
    err = builder.AddTextField(
        "path",
        true,  // stored
        false, // indexed as text
        false, // fast
        tantivy.IndexRecordOptionBasic,
        tantivy.TokenizerRaw,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to add path field: %w", err)
    }
    
    err = builder.AddTextField(
        "content",
        false, // not stored (we'll read from file)
        true,  // indexed as text
        false, // fast
        tantivy.IndexRecordOptionWithFreqsAndPositions,
        tantivy.TokenizerSimple,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to add content field: %w", err)
    }
    
    // For now, store modified as text field since AddI64Field is not available
    err = builder.AddTextField(
        "modified",
        true,  // stored
        false, // not indexed as text
        false, // fast
        tantivy.IndexRecordOptionBasic,
        tantivy.TokenizerRaw,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to add modified field: %w", err)
    }
    
    err = builder.AddTextField(
        "title",
        true,  // stored
        true,  // indexed as text
        false, // fast
        tantivy.IndexRecordOptionWithFreqsAndPositions,
        tantivy.TokenizerSimple,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to add title field: %w", err)
    }
    
    // Build schema
    schema, err := builder.BuildSchema()
    if err != nil {
        return nil, fmt.Errorf("failed to build schema: %w", err)
    }
    
    // Create or open index
    if _, err := os.Stat(indexPath); os.IsNotExist(err) {
        os.MkdirAll(indexPath, 0755)
    }
    
    context, err := tantivy.NewTantivyContextWithSchema(indexPath, schema)
    if err != nil {
        return nil, fmt.Errorf("failed to create index context: %w", err)
    }
    
    // Register tokenizers
    err = context.RegisterTextAnalyzerSimple(tantivy.TokenizerSimple, 10000, tantivy.English)
    if err != nil {
        return nil, fmt.Errorf("failed to register simple analyzer: %w", err)
    }
    
    err = context.RegisterTextAnalyzerRaw(tantivy.TokenizerRaw)
    if err != nil {
        return nil, fmt.Errorf("failed to register raw analyzer: %w", err)
    }
    
    return &TantivyIndex{
        context:     context,
        indexPath:   indexPath,
        lastIndexed: make(map[string]time.Time),
    }, nil
}

func (ti *TantivyIndex) IndexDirectory(rootPath string, numWorkers int) error {
    ti.mu.Lock()
    defer ti.mu.Unlock()
    
    // Load saved timestamps
    ti.loadIndexTimestamps()
    
    type indexJob struct {
        path string
        info os.FileInfo
    }
    
    jobs := make(chan indexJob, 100)
    var wg sync.WaitGroup
    
    // Start workers
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                ti.indexFile(job.path, job.info)
            }
        }()
    }
    
    // Directory traversal
    err := godirwalk.Walk(rootPath, &godirwalk.Options{
        Callback: func(path string, de *godirwalk.Dirent) error {
            if !strings.HasSuffix(path, ".md") {
                return nil
            }
            
            info, err := os.Stat(path)
            if err != nil {
                return nil
            }
            
            // Only index modified files
            if lastIndexed, exists := ti.lastIndexed[path]; exists {
                if info.ModTime().Before(lastIndexed) || info.ModTime().Equal(lastIndexed) {
                    return nil
                }
            }
            
            jobs <- indexJob{path: path, info: info}
            return nil
        },
        Unsorted: true,
        ErrorCallback: func(path string, err error) godirwalk.ErrorAction {
            fmt.Printf("Error walking %s: %v\n", path, err)
            return godirwalk.SkipNode
        },
    })
    
    close(jobs)
    wg.Wait()
    
    // Save timestamps
    ti.saveIndexTimestamps()
    
    return err
}

func (ti *TantivyIndex) indexFile(path string, info os.FileInfo) error {
    content, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    
    // Extract title from first line
    lines := strings.Split(string(content), "\n")
    title := filepath.Base(path)
    if len(lines) > 0 && strings.HasPrefix(lines[0], "#") {
        title = strings.TrimSpace(strings.TrimPrefix(lines[0], "#"))
    }
    
    // Delete old document if exists
    err = ti.context.DeleteDocuments("path", path)
    if err != nil {
        // Ignore error - document might not exist
    }
    
    // Create new document
    doc := tantivy.NewDocument()
    if doc == nil {
        return fmt.Errorf("failed to create document")
    }
    
    // Add fields
    err = doc.AddField(path, ti.context, "path")
    if err != nil {
        return fmt.Errorf("failed to add path field: %w", err)
    }
    
    err = doc.AddField(string(content), ti.context, "content")
    if err != nil {
        return fmt.Errorf("failed to add content field: %w", err)
    }
    
    err = doc.AddField(fmt.Sprintf("%d", info.ModTime().Unix()), ti.context, "modified")
    if err != nil {
        return fmt.Errorf("failed to add modified field: %w", err)
    }
    
    err = doc.AddField(title, ti.context, "title")
    if err != nil {
        return fmt.Errorf("failed to add title field: %w", err)
    }
    
    // Add document
    err = ti.context.AddAndConsumeDocuments(doc)
    if err != nil {
        return fmt.Errorf("failed to add document: %w", err)
    }
    
    // Update timestamp
    ti.lastIndexed[path] = info.ModTime()
    
    return nil
}

func (ti *TantivyIndex) Search(query string, limit int) ([]SearchResult, error) {
    ti.mu.RLock()
    defer ti.mu.RUnlock()
    
    // Build search context
    searchCtx := tantivy.NewSearchContextBuilder().
        SetQuery(query).
        SetDocsLimit(uintptr(limit)).
        SetWithHighlights(true).
        AddFieldDefaultWeight("content").
        AddFieldDefaultWeight("title").
        Build()
    
    // Search
    searchResult, err := ti.context.Search(searchCtx)
    if err != nil {
        return nil, fmt.Errorf("search failed: %w", err)
    }
    defer searchResult.Free()
    
    results := make([]SearchResult, 0)
    
    size, err := searchResult.GetSize()
    if err != nil {
        return nil, fmt.Errorf("failed to get result size: %w", err)
    }
    
    for i := uint64(0); i < size; i++ {
        doc, err := searchResult.Get(i)
        if err != nil {
            continue
        }
        
        // Get JSON representation
        // Get fields manually since GetSchema is not available
        jsonStr, err := doc.ToJson(ti.context, "path", "title")
        if err != nil {
            doc.Free()
            continue
        }
        
        // Parse path from JSON (simple extraction)
        pathStart := strings.Index(jsonStr, `"path":"`) + 8
        pathEnd := strings.Index(jsonStr[pathStart:], `"`)
        path := jsonStr[pathStart : pathStart+pathEnd]
        
        // Get snippet from highlights
        snippet := ""
        lineNumbers := []int{}
        
        // For now, create a simple snippet
        if content, err := os.ReadFile(path); err == nil {
            snippet, lineNumbers = ti.createSnippet(string(content), query, 150)
        }
        
        results = append(results, SearchResult{
            FilePath:    path,
            Snippet:     snippet,
            Score:       1.0, // tantivy-go doesn't expose scores directly
            LineNumbers: lineNumbers,
        })
        
        doc.Free()
    }
    
    return results, nil
}

func (ti *TantivyIndex) createSnippet(content, query string, maxLength int) (string, []int) {
    lines := strings.Split(content, "\n")
    queryLower := strings.ToLower(query)
    var matchedLines []int
    var snippetParts []string
    
    for i, line := range lines {
        if strings.Contains(strings.ToLower(line), queryLower) {
            matchedLines = append(matchedLines, i+1)
            if len(snippetParts) < 3 { // Max 3 lines in snippet
                snippetParts = append(snippetParts, fmt.Sprintf("L%d: %s", i+1, line))
            }
        }
    }
    
    snippet := strings.Join(snippetParts, "\n")
    if len(snippet) > maxLength {
        snippet = snippet[:maxLength] + "..."
    }
    
    return snippet, matchedLines
}

func (ti *TantivyIndex) loadIndexTimestamps() {
    timestampFile := filepath.Join(ti.indexPath, ".timestamps")
    data, err := os.ReadFile(timestampFile)
    if err != nil {
        return
    }
    
    lines := strings.Split(string(data), "\n")
    for _, line := range lines {
        parts := strings.Split(line, "|")
        if len(parts) == 2 {
            timestamp, _ := time.Parse(time.RFC3339, parts[1])
            ti.lastIndexed[parts[0]] = timestamp
        }
    }
}

func (ti *TantivyIndex) saveIndexTimestamps() {
    timestampFile := filepath.Join(ti.indexPath, ".timestamps")
    var lines []string
    
    for path, timestamp := range ti.lastIndexed {
        lines = append(lines, fmt.Sprintf("%s|%s", path, timestamp.Format(time.RFC3339)))
    }
    
    os.WriteFile(timestampFile, []byte(strings.Join(lines, "\n")), 0644)
}

func (ti *TantivyIndex) UpdateFile(path string) error {
    ti.mu.Lock()
    defer ti.mu.Unlock()
    
    info, err := os.Stat(path)
    if err != nil {
        return err
    }
    
    return ti.indexFile(path, info)
}

func (ti *TantivyIndex) RemoveFile(path string) error {
    ti.mu.Lock()
    defer ti.mu.Unlock()
    
    err := ti.context.DeleteDocuments("path", path)
    if err != nil {
        return err
    }
    
    delete(ti.lastIndexed, path)
    
    return nil
}

func (ti *TantivyIndex) Close() error {
    ti.saveIndexTimestamps()
    ti.context.Free()
    return nil
}

func (ti *TantivyIndex) GetIndexedFilesCount() int {
    ti.mu.RLock()
    defer ti.mu.RUnlock()
    return len(ti.lastIndexed)
}