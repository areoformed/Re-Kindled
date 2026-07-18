# Footprint

Measured from the validated 0.1.0 release build on macOS:

| Artifact | Bytes | Lives where |
| --- | ---: | --- |
| Stripped stdio MCP | 2,819,152 | Host only |
| Stripped ARMv7 reader | 2,359,458 | Kindle |
| Locally derived reference font | 179,888 | Kindle |
| Public source before manifests | about 174,000 | Host/repository |

The incremental Kindle payload is therefore about 2.54 MB for one reader, one face, scripts, preset, and content. FBInk is a separately installed prerequisite and is not counted. The MCP process stays on the host and adds no Kindle storage.

Both Go programs are built with `CGO_ENABLED=0`, `-trimpath`, and stripped symbol/debug data (`-ldflags='-s -w'`). They use no third-party Go module and open no database or web service. The single-file Type Lab adds no runtime dependency.

Build caches are intentionally ignored and excluded from manifests; they can be much larger than the shipped artifacts. Public releases contain source only.
