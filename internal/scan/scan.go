package scan

import (
    "bufio"
    "fmt"
    "io"
    "io/fs"
    "os"
    "path/filepath"
    "runtime"
    "strings"
    "sync"
)

// Config defines how scanning should run.
type Config struct {
    Root         string
    Workers      int
    MaxFileSize  int64
    ShowLines    bool
    SkipDirNames map[string]struct{}
    Verbose      bool
}

type Result struct {
    Category     string   `json:"category"`
    Port         string   `json:"port"`
    Path         string   `json:"path"`
    File         string   `json:"file"`
    MatchedTerms []string `json:"matched_terms"`
    Lines        []string `json:"lines,omitempty"`
}

// Run walks the tree and returns results for any port whose Makefile matches terms.
func Run(cfg Config, terms TermSet) ([]Result, error) {
    if cfg.Workers <= 0 {
        cfg.Workers = runtime.NumCPU()
    }
    if cfg.MaxFileSize <= 0 {
        cfg.MaxFileSize = 1 << 20 // 1 MiB safeguard
    }
    jobs := make(chan string, cfg.Workers*2)
    var (
        wg sync.WaitGroup
        mu sync.Mutex
        res []Result
    )

    // workers
    for i := 0; i < cfg.Workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for path := range jobs {
                if r, ok := scanMakefile(path, cfg, terms); ok {
                    mu.Lock()
                    res = append(res, r)
                    mu.Unlock()
                }
            }
        }()
    }

    err := filepath.WalkDir(cfg.Root, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        // Skip configured dir names
        if d.IsDir() {
            if _, hit := cfg.SkipDirNames[d.Name()]; hit {
                return fs.SkipDir
            }
            return nil
        }
        if strings.EqualFold(d.Name(), "Makefile") {
            jobs <- path
        }
        return nil
    })
    close(jobs)
    wg.Wait()
    return res, err
}

func scanMakefile(path string, cfg Config, terms TermSet) (Result, bool) {
    info, err := os.Stat(path)
    if err != nil {
        return Result{}, false
    }
    if info.Size() > cfg.MaxFileSize {
        if cfg.Verbose {
            fmt.Fprintf(os.Stderr, "skip large file: %s (%d bytes)\n", path, info.Size())
        }
        return Result{}, false
    }
    f, err := os.Open(path)
    if err != nil {
        return Result{}, false
    }
    defer f.Close()

    matched := make(map[string]struct{})
    var linesOut []string

    s := bufio.NewScanner(io.LimitReader(f, cfg.MaxFileSize))
    for s.Scan() {
        line := s.Text()
        if !looksRelevant(line) {
            continue
        }
        for _, rx := range terms.Patterns {
            if rx.FindStringIndex(line) != nil {
                matched[rx.String()] = struct{}{}
                if cfg.ShowLines {
                    linesOut = append(linesOut, line)
                }
            }
        }
    }
    _ = s.Err()

    if len(matched) == 0 {
        return Result{}, false
    }
    cat, port := derivePortCoords(path)
    return Result{
        Category:     cat,
        Port:         port,
        Path:         filepath.Dir(path),
        File:         filepath.Base(path),
        MatchedTerms: keys(matched),
        Lines:        linesOut,
    }, true
}

// Only skip blank and comment lines. This avoids missing valid matches.
func looksRelevant(line string) bool {
    line = strings.TrimSpace(line)
    if line == "" || strings.HasPrefix(line, "#") {
        return false
    }
    return true
}

func derivePortCoords(makefilePath string) (string, string) {
    dir := filepath.Dir(makefilePath)
    port := filepath.Base(dir)
    cat := filepath.Base(filepath.Dir(dir))
    return cat, port
}

func keys(m map[string]struct{}) []string {
    out := make([]string, 0, len(m))
    for k := range m {
        out = append(out, k)
    }
    return out
}

// ParseCSV builds a set from a comma separated string.
func ParseCSV(s string) map[string]struct{} {
    m := map[string]struct{}{}
    for _, part := range strings.Split(s, ",") {
        p := strings.TrimSpace(part)
        if p != "" {
            m[p] = struct{}{}
        }
    }
    return m
}
