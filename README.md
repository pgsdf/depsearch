# depsearch

`depsearch` scans a GhostBSD or FreeBSD ports tree and lists ports whose `Makefile` declares dependencies that match **your** patterns.  
Patterns are provided in a text file you control.

- One pattern per line  
- Blank lines and lines starting with `#` are ignored  
- Each line is a Go regular expression (RE2). For literal words, use `\bword\b`

---

## Install

```sh
make
sudo make install
````

---

## Usage

```sh
depsearch [flags] [PORTS_ROOT]

Flags:
  -terms string   path to terms file (one regex per line); if empty, tries $DEPSEARCH_TERMS
  -format string  output format: table, json, csv, paths (default "table")
  -j int          number of worker threads (default: number of CPUs)
  -max-size int   max Makefile size to read in bytes (default 1048576)
  -show-lines     include matching line text in results
  -skip-dirs      comma separated dir names to skip (default "distfiles,.git,.idea,.vscode,packages,work")
  -v              verbose logging to stderr
```

---

### Example

Create `wayland.terms`:

```text
(?i)\bwayland\b
(?i)\bwayland-protocols\b
(?i)\bwlroots\b
(?i)\blibwayland(-client|-server)?\b
(?i)\bseatd\b
(?i)\bsway\b
```

Run:

```sh
depsearch -terms ./samples/wayland.terms -format table -show-lines /usr/ports

# Produce Makefile paths only (easy to pipe into sed/xargs)
depsearch -terms ./samples/wayland.terms -format paths /usr/ports
```

---

### Editing matched ports

Comment out matching dependency lines conservatively:

```sh
depsearch -terms ./samples/wayland.terms -format paths /usr/ports \
| sudo xargs -I{} sed -i '' -E \
  '/^(USES|LIB_DEPENDS|BUILD_DEPENDS|RUN_DEPENDS|USE_.*)=/I {/wayland|wlroots|libwayland|wayland-protocols/I s/^/# depsearch: disabled /}' {}
```

Or surgically remove only matched tokens (example shows Wayland subset; adjust to your `terms`):

```sh
depsearch -terms ./samples/wayland.terms -format paths /usr/ports \
| sudo xargs -I{} sed -i '' -E \
  '/^(USES|LIB_DEPENDS|BUILD_DEPENDS|RUN_DEPENDS|USE_.*)=/I {
    s/(^|[[:space:],])(wayland|wlroots|libwayland(-client|-server)?|wayland-protocols)([[:space:],]|$)/\1\4/Ig;
    s/,[[:space:]]*,/,/g; s/[[:space:]]+,/,/g; s/=[[:space:]]*,/=/g; s/,[[:space:]]*$//;
  }' {}
```

---

## Notes

* The parser is heuristic. It does not evaluate bmake includes; it scans raw text for matches.
* To widen coverage, edit the `terms` file.
* If you want full make evaluation later, an optional mode can be added that shells out to `make -V` in each port.

---

## License

BSD 2-Clause. See [`LICENSE`](LICENSE).

```
