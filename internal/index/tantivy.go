package index

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"
    
    "github.com/anyproto/tantivy-go"
    "github.com/karrick/godirwalk"
)

type SearchResult struct {
    FilePath    string   `json:"file_path"`
    Snippet     string   `json:"snippet"`
    Score       float32  `json:"score"`
    LineNumbers []int    `json:"line_numbers"`
}

type TantivyIndex struct {
    schema      *tantivy.SchemaBuilder
    index       *tantivy.Index
    writer      *tantivy.IndexWriter
    searcher    *tantivy.Searcher
    indexPath   string
    mu          sync.RWMutex
    lastIndexed map[string]time.Time
}

func NewTantivyIndex(indexPath string) (*TantivyIndex, error) {
    // Schema erstellen
    schema := tantivy.NewSchemaBuilder()
    schema.AddTextField("path", tantivy.Stored)
    schema.AddTextField("content", tantivy.IndexRecordOption)
    schema.AddI64Field("modified", tantivy.Stored)
    schema.AddTextField("title", tantivy.Stored)
    
    builtSchema := schema.Build()
    
    // Index erstellen oder öffnen
    var index *tantivy.Index
    var err error
    
    if _, err := os.Stat(indexPath); os.IsNotExist(err) {
        os.MkdirAll(indexPath, 0755)
        index, err = tantivy.NewIndex(builtSchema, indexPath)
    } else {
        index, err = tantivy.OpenIndex(indexPath)
    }
    
    if err != nil {
        return nil, fmt.Errorf("failed to open index: %w", err)
    }
    
    writer, err := index.Writer(50_000_000) // 50MB buffer
    if err != nil {
        return nil, fmt.Errorf("failed to create writer: %w", err)
    }
    
    searcher := index.Searcher()
    
    return &TantivyIndex{
        schema:      schema,
        index:       index,
        writer:      writer,
        searcher:    searcher,
        indexPath:   indexPath,
        lastIndexed: make(map[string]time.Time),
    }, nil
}

func (ti *TantivyIndex) IndexDirectory(rootPath string, numWorkers int) error {
    ti.mu.Lock()
    defer ti.mu.Unlock()
    
    // Lade gespeicherte Index-Zeiten
    ti.loadIndexTimestamps()
    
    type indexJob struct {
        path string
        info os.FileInfo
    }
    
    jobs := make(chan indexJob, 100)
    var wg sync.WaitGroup
    
    // Worker starten
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
            
            // Nur neu modifizierte Dateien indexieren
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
    
    // Commit changes
    ti.writer.Commit()
    ti.saveIndexTimestamps()
    
    return err
}

func (ti *TantivyIndex) indexFile(path string, info os.FileInfo) error {
    content, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    
    // Dokument erstellen
    doc := tantivy.NewDocument()
    doc.AddTextField("path", path)
    doc.AddTextField("content", string(content))
    doc.AddI64Field("modified", info.ModTime().Unix())
    
    // Titel aus erster Zeile extrahieren
    lines := strings.Split(string(content), "\n")
    title := filepath.Base(path)
    if len(lines) > 0 && strings.HasPrefix(lines[0], "#") {
        title = strings.TrimSpace(strings.TrimPrefix(lines[0], "#"))
    }
    doc.AddTextField("title", title)
    
    // Altes Dokument löschen
    ti.writer.DeleteTerm(tantivy.NewTerm("path", path))
    
    // Neues Dokument hinzufügen
    ti.writer.AddDocument(doc)
    
    // Timestamp aktualisieren
    ti.lastIndexed[path] = info.ModTime()
    
    return nil
}

func (ti *TantivyIndex) Search(query string, limit int) ([]SearchResult, error) {
    ti.mu.RLock()
    defer ti.mu.RUnlock()
    
    // Query parser
    queryParser := tantivy.NewQueryParser(ti.index, []string{"content", "title"})
    tantivyQuery, err := queryParser.ParseQuery(query)
    if err != nil {
        return nil, fmt.Errorf("failed to parse query: %w", err)
    }
    
    // Search
    topDocs, err := ti.searcher.Search(tantivyQuery, limit)
    if err != nil {
        return nil, fmt.Errorf("search failed: %w", err)
    }
    
    results := make([]SearchResult, 0, len(topDocs))
    
    for _, scoreDoc := range topDocs {
        doc, err := ti.searcher.Doc(scoreDoc.DocId)
        if err != nil {
            continue
        }
        
        path := doc.GetFirstTextField("path")
        content := doc.GetFirstTextField("content")
        
        // Snippet erstellen
        snippet, lineNumbers := ti.createSnippet(content, query, 150)
        
        results = append(results, SearchResult{
            FilePath:    path,
            Snippet:     snippet,
            Score:       scoreDoc.Score,
            LineNumbers: lineNumbers,
        })
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
            if len(snippetParts) < 3 { // Max 3 Zeilen im Snippet
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
    
    ti.writer.DeleteTerm(tantivy.NewTerm("path", path))
    ti.writer.Commit()
    delete(ti.lastIndexed, path)
    
    return nil
}

func (ti *TantivyIndex) Close() error {
    ti.writer.Commit()
    ti.saveIndexTimestamps()
    return nil
}