package index

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"
    
    "github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
    watcher  *fsnotify.Watcher
    index    *TantivyIndex
    rootPath string
    mu       sync.Mutex
    events   map[string]time.Time
}

func NewFileWatcher(index *TantivyIndex, rootPath string) (*FileWatcher, error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }
    
    fw := &FileWatcher{
        watcher:  watcher,
        index:    index,
        rootPath: rootPath,
        events:   make(map[string]time.Time),
    }
    
    // Rekursiv alle Directories hinzufügen
    err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() && !strings.Contains(path, ".git") {
            return watcher.Add(path)
        }
        return nil
    })
    
    return fw, err
}

func (fw *FileWatcher) Start() {
    debounceTimer := time.NewTicker(500 * time.Millisecond)
    defer debounceTimer.Stop()
    
    for {
        select {
        case event, ok := <-fw.watcher.Events:
            if !ok {
                return
            }
            fw.handleEvent(event)
            
        case <-debounceTimer.C:
            fw.processPendingEvents()
            
        case err, ok := <-fw.watcher.Errors:
            if !ok {
                return
            }
            fmt.Printf("Watcher error: %v\n", err)
        }
    }
}

func (fw *FileWatcher) handleEvent(event fsnotify.Event) {
    if !strings.HasSuffix(event.Name, ".md") {
        return
    }
    
    fw.mu.Lock()
    defer fw.mu.Unlock()
    
    fw.events[event.Name] = time.Now()
}

func (fw *FileWatcher) processPendingEvents() {
    fw.mu.Lock()
    events := make(map[string]time.Time)
    for k, v := range fw.events {
        events[k] = v
    }
    fw.events = make(map[string]time.Time)
    fw.mu.Unlock()
    
    for path := range events {
        if _, err := os.Stat(path); os.IsNotExist(err) {
            fw.index.RemoveFile(path)
        } else {
            fw.index.UpdateFile(path)
        }
    }
    
    if len(events) > 0 {
        fw.index.writer.Commit()
    }
}

func (fw *FileWatcher) Stop() error {
    return fw.watcher.Close()
}