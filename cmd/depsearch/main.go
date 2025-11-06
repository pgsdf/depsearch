package main

import (
    "flag"
    "fmt"
    "os"
    "runtime"

    "github.com/pgsdf/depsearch/internal/output"
    "github.com/pgsdf/depsearch/internal/scan"
)

func main() {
    termsPath := flag.String("terms", "", "path to terms file (one regex per line); falls back to $DEPSEARCH_TERMS if empty")
    format := flag.String("format", "table", "output format: table, json, csv, paths")
    j := flag.Int("j", runtime.NumCPU(), "number of worker threads")
    maxSize := flag.Int("max-size", 1<<20, "max file size in bytes for Makefile reads")
    showLines := flag.Bool("show-lines", false, "include matching line text in results")
    skipDirs := flag.String("skip-dirs", "distfiles,.git,.idea,.vscode,packages,work", "comma separated dir names to skip")
    verbose := flag.Bool("v", false, "verbose logging to stderr")
    flag.Parse()

    root := "."
    if flag.NArg() > 0 {
        root = flag.Arg(0)
    }

    path := *termsPath
    if path == "" {
        path = os.Getenv("DEPSEARCH_TERMS")
    }
    if path == "" {
        fmt.Fprintln(os.Stderr, "depsearch: no -terms file provided and DEPSEARCH_TERMS not set")
        os.Exit(2)
    }

    termSet, err := scan.LoadTermsFromFile(path)
    if err != nil {
        fmt.Fprintf(os.Stderr, "depsearch: loading terms: %v\n", err)
        os.Exit(2)
    }

    cfg := scan.Config{
        Root:         root,
        Workers:      *j,
        MaxFileSize:  int64(*maxSize),
        ShowLines:    *showLines,
        SkipDirNames: scan.ParseCSV(*skipDirs),
        Verbose:      *verbose,
    }

    results, err := scan.Run(cfg, termSet)
    if err != nil {
        fmt.Fprintf(os.Stderr, "depsearch: %v\n", err)
        os.Exit(1)
    }

    switch *format {
    case "json":
        output.AsJSON(os.Stdout, results)
    case "csv":
        output.AsCSV(os.Stdout, results)
    case "paths":
        output.AsPaths(os.Stdout, results)
    default:
        output.AsTable(os.Stdout, results)
    }
}
