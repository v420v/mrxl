# MRXL

A CLI tool that converts Mermaid diagrams into Excel (`.xlsx`) files, perfect for sharing in presentations and reviews.

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go)](https://go.dev/)

## Features
- [x] Sequence Diagram
- [ ] Flowchart
- [ ] Class Diagram
- [ ] State Diagram
- [ ] Entity Relationship Diagram
- [ ] User Journey Diagram
- [ ] Gantt diagrams
- [ ] Pie chart diagrams
- [ ] Quadrant Chart
- [ ] Requirement Diagram
- [ ] GitGraph Diagrams
- [ ] C4 Diagrams
- [ ] Mindmap
- [ ] Timeline Diagram

## Quick Start

```bash
go run . -src /path/to/mermaid.mmd -out /path/to/diagram.xlsx
```

## CLI Usage

```bash
go run . -src INPUT_FILE -out OUTPUT_FILE
```

- `-src`: Input file (required)
- `-out`: Output file (optional, default: `mermaid.out.xlsx`)

## Authors

- [@ryutaKimu](https://github.com/ryutaKimu)
- [@v420v](https://github.com/v420v)
