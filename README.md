# gh-action-lens

A GitHub CLI extension for exploring GitHub Actions workflows and analyzing action usage across organizations.

## Overview

gh-action-lens helps you discover and analyze GitHub Actions usage across your organization's repositories. It scans workflow files, extracts action usage patterns, and provides detailed insights through multiple output formats.

> **Note:** This extension analyzes workflow configurations and action declarations, not workflow execution history or run logs. It shows you *what* actions are defined in your workflows and *how often* they're used, but doesn't access runtime data or execution results.

## Features

### Workflow Discovery
- Scan all repositories in an organization for GitHub Actions workflows
- Identify repositories with `.yml` and `.yaml` workflow files

### Action Analysis
- Extract and catalog all GitHub Actions used across workflows
- Count usage frequencies and track action versions
- Deduplicate actions by name and version

### Multiple Output Formats
- **Tree View**: Hierarchical display with visual indicators (default)
- **Table**: Professional tabular output for detailed analysis  
- **JSON**: Structured data for programmatic processing
- **CSV**: Spreadsheet-friendly format for data analysis

### Organization Ready
- Organization-wide scanning capabilities
- Authenticated access via GitHub CLI credentials
- Efficient GraphQL and REST API integration

---

## Installation

### Prerequisites

- [GitHub CLI](https://cli.github.com/) installed and authenticated
- Go 1.23.0 or later (for development)

### Install from source

1. Clone this repository:

   ```bash
   git clone https://github.com/jefeish/gh-action-lens.git
   cd gh-action-lens
   ```

2. Build and install the extension:

   ```bash
   go build -o gh-action-lens
   gh extension install .
   ```

### Install from GitHub (when published)

```bash
gh extension install jefeish/gh-action-lens
```

## Usage

Once installed, you can use the extension with:

```bash
gh action-lens [flags]
```

### Available Flags

- `-h, --help`: Show help information
- `-o, --org <string>`: Organization name to target
- `-s, --scan <string>`: Scan scope: workflows, actions, or all (default "all")
- `-d, --detailed`: Detailed analysis with comprehensive action breakdown
- `-f, --format <string>`: Output format: default, json, table, csv (default "default")
- `--output <string>`: Write output to file instead of stdout

### Examples

```bash
# Basic usage
gh action-lens                                 # Run with defaults
gh action-lens --help                          # Show help message

# Target specific organization
gh action-lens -o myorg                        # Scan all workflows and actions
gh action-lens -o myorg --scan workflows       # Scan workflows only
gh action-lens -o myorg --scan actions         # Analyze actions only

# Detailed analysis
gh action-lens -o myorg --scan all --detailed  # Comprehensive action breakdown

# Output formatting
gh action-lens -o myorg --format json          # Output results as JSON
gh action-lens -o myorg --format csv           # Output results as CSV
gh action-lens -o myorg --output results.txt   # Write output to file
```

### Authentication

The extension uses your existing GitHub CLI credentials. If not authenticated, run:

```bash
gh auth login
```

Alternatively, set environment variables: `GITHUB_TOKEN` or `GH_TOKEN`

## Technical Documentation

For detailed usage examples, output format specifications, development information, and implementation details, see [TECHNICAL.md](TECHNICAL.md).

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)  
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

If you encounter any issues or have questions, please [open an issue](https://github.com/jefeish/gh-action-lens/issues) on GitHub.