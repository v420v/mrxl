# MRXL

![Screenshot of MRXL](./images/top.png)

A CLI tool that converts Mermaid diagrams into Excel (`.xlsx`) files.

[![Build Check](https://github.com/v420v/mrxl/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/v420v/mrxl/actions/workflows/build.yml) [![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go)](https://go.dev/) [![codecov](https://codecov.io/gh/v420v/mrxl/graph/badge.svg)](https://codecov.io/gh/v420v/mrxl)

## Features

✅ Full support · ⚠️ Partial support · ❌ Not yet implemented

| Name | Status |
|---------|--------|
| Sequence Diagram | ⚠️ |
| Flowchart | ❌ |
| Class Diagram | ❌ |
| State Diagram | ❌ |
| Entity Relationship Diagram | ❌ |
| User Journey Diagram | ⚠️ |
| Gantt Diagram | ⚠️ |
| Pie Chart Diagram | ✅ |
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

## Contributing

Bug reports and feature requests are welcome — please use the [issue templates](.github/ISSUE_TEMPLATE/).

Pull requests are appreciated. See [CONTRIBUTING.md](CONTRIBUTING.md) for setup instructions, project structure, and commit style guidelines.

## Authors

- [@ryutaKimu](https://github.com/ryutaKimu)
- [@v420v](https://github.com/v420v)
