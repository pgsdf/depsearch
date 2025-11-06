package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/pgsdf/depsearch/internal/scan"
)

// AsPaths prints one Makefile path per line, suitable for piping to xargs/sed.
func AsPaths(w io.Writer, results []scan.Result) {
	seen := make(map[string]struct{})
	rows := sorted(results)
	for _, r := range rows {
		p := r.Path + "/" + r.File
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		fmt.Fprintln(w, p)
	}
}

func AsJSON(w io.Writer, results []scan.Result) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(sorted(results))
}

func AsCSV(w io.Writer, results []scan.Result) {
	cw := csv.NewWriter(w)
	defer cw.Flush()
	_ = cw.Write([]string{"category", "port", "path", "file", "matched_terms"})
	for _, r := range sorted(results) {
		_ = cw.Write([]string{
			r.Category,
			r.Port,
			r.Path,
			r.File,
			strings.Join(r.MatchedTerms, "|"),
		})
	}
}

func AsTable(w io.Writer, results []scan.Result) {
	rows := sorted(results)
	if len(rows) == 0 {
		fmt.Fprintln(w, "No matches found.")
		return
	}

	// compute widths
	w1, w2 := 0, 0
	for _, r := range rows {
		l := len(r.Category + "/" + r.Port)
		if l > w1 {
			w1 = l
		}
		l2 := len(strings.Join(r.MatchedTerms, ", "))
		if l2 > w2 {
			w2 = l2
		}
	}

	fmt.Fprintf(w, "%-*s  %-*s  %s\n", w1, "CATEGORY/PORT", w2, "MATCHED_TERMS", "FILE")
	for _, r := range rows {
		fmt.Fprintf(w, "%-*s  %-*s  %s\n",
			w1, r.Category+"/"+r.Port,
			w2, strings.Join(r.MatchedTerms, ", "),
			r.Path+"/"+r.File,
		)
		if len(r.Lines) > 0 {
			for _, ln := range r.Lines {
				fmt.Fprintf(w, "    %s\n", ln)
			}
		}
	}
}

func sorted(results []scan.Result) []scan.Result {
	out := append([]scan.Result(nil), results...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Category == out[j].Category {
			return out[i].Port < out[j].Port
		}
		return out[i].Category < out[j].Category
	})
	return out
}

