# MRXL

![Screenshot of MRXL](./images/top.png)

A CLI tool that converts Mermaid diagrams into Excel (`.xlsx`) files.

[![Build Check](https://github.com/v420v/mrxl/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/v420v/mrxl/actions/workflows/build.yml) [![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go)](https://go.dev/)

## Features
| Name | Status |
|---------|--------|
| Sequence Diagram | ⚠️ |
| Flowchart | ❌ |
| Class Diagram | ❌ |
| State Diagram | ❌ |
| Entity Relationship Diagram | ❌ |
| User Journey Diagram | ⚠️ |
| Gantt Diagram | ❌ |
| Pie Chart Diagram | ⚠️ |
| Quadrant Chart | ⚠️ |
| Requirement Diagram | ❌ |
| GitGraph Diagram | ❌ |
| C4 Diagram | ❌ |
| Mindmap | ❌ |
| Timeline Diagram | ⚠️ |

## Quick Start

```bash
go build cmd/mrxl.go 
./mrxl -src examples/SequenceDiagram.mmd
```

## CLI Usage

```bash
mrxl -src INPUT_FILE -out OUTPUT_FILE
```

- `-src`: Input file (required)
- `-out`: Output file (optional, default: `mermaid.out.xlsx`)

## Authors

- [@ryutaKimu](https://github.com/ryutaKimu)
- [@v420v](https://github.com/v420v)
