# Technical Documentation

This document contains technical information about the gh-action-lens extension, including detailed usage examples, output formats, development information, and implementation details.

## Detailed Usage Examples

### Output Format Options

The `--format` flag supports four different output formats for comprehensive analysis:

#### `default` (Tree View)

- **Best for**: Human-readable overview and exploration
- **Features**: Hierarchical tree structure with emojis and visual indicators
- **Shows**: Repository → Workflow → Actions with usage counts
- **Benefits**: Easy to scan, visually appealing, shows relationships clearly

```bash
gh action-lens -o myorg --scan all --detailed --format default
# or simply:
gh action-lens -o myorg --scan all --detailed
```

#### `table` (Tabular Format)

- **Best for**: Detailed analysis and comparison
- **Features**: Professional table with columns for Repository, Workflow, Action, Version, Count, and Total
- **Shows**: All data in structured rows with clear column headers
- **Benefits**: Easy to compare counts, export-friendly, shows total action counts per workflow

```bash
gh action-lens -o myorg --scan all --detailed --format table
```

#### `json` (JSON Output)

- **Best for**: Programmatic processing and integration
- **Features**: Structured JSON with nested objects for repositories, workflows, and actions
- **Shows**: Complete data structure with all metadata and counts
- **Benefits**: Perfect for scripts, APIs, further processing, or data storage

```bash
gh action-lens -o myorg --scan all --detailed --format json
```

#### `csv` (CSV Output)

- **Best for**: Data analysis and spreadsheet integration
- **Features**: Comma-separated values format with proper escaping for special characters
- **Shows**: Tabular data with columns for Repository, Workflow, Action, Version, Count, and Total
- **Benefits**: Perfect for Excel/Google Sheets, data analysis tools, and database imports

```bash
gh action-lens -o myorg --scan all --detailed --format csv
```

### File Output

All output formats support writing results to a file instead of displaying on the terminal:

```bash
# Write any format to file using --output
gh action-lens -o myorg --scan all --detailed --output report.txt
gh action-lens -o myorg --scan all --detailed --format table --output table-report.txt
gh action-lens -o myorg --scan all --detailed --format json --output data.json
gh action-lens -o myorg --scan all --detailed --format csv --output data.csv
```

**Benefits of file output:**

- **Save for later analysis**: Keep reports for comparison over time
- **Share with team**: Export results for team review and planning
- **Process programmatically**: Use JSON files with other tools and scripts
- **Archive documentation**: Maintain historical records of GitHub Actions usage

### Authentication

The extension supports multiple authentication methods:

1. GitHub CLI authentication (recommended): `gh auth login`
2. Environment variables: `GITHUB_TOKEN` or `GH_TOKEN`

## Example Outputs

### Basic Scan Output

```text
🔍 Scanning organization: myorg

📁 my-repo:
  └─ .github/workflows/ci.yml
  └─ .github/workflows/deploy.yaml

📁 another-repo:
  └─ .github/workflows/test.yml

✅ Scan complete!
📊 Summary: Found 2 repositories with workflows out of 15 total repositories.
⏱️  Process time: 1.234s
```

### Detailed Analysis - Default Tree View

```text
📁 my-web-app → 📄 .github/workflows/ci.yml (3 unique, 5 total actions)
📁 my-web-app → 📄 .github/workflows/deploy.yml (2 unique, 3 total actions)
📁 api-service → 📄 .github/workflows/test.yml (4 unique, 6 total actions)

🔍 Detailed Analysis Results
============================================================

📁 my-web-app (2 workflows)
   📄 .github/workflows/ci.yml (3 unique, 5 total actions)
      🔧 actions/checkout@v4 (2 times)
      🔧 actions/setup-node@v4
      🔧 actions/upload-artifact@v4 (2 times)
   📄 .github/workflows/deploy.yml (2 unique, 3 total actions)
      🔧 actions/checkout@v4
      🔧 actions/deploy-pages@v4 (2 times)

📁 api-service (1 workflows)
   📄 .github/workflows/test.yml (4 unique, 6 total actions)
      🔧 actions/checkout@v4
      🔧 actions/setup-go@v5
      🔧 actions/cache@v4 (2 times)
      🔧 codecov/codecov-action@v4 (2 times)

📊 Summary:
   • Total repositories: 10
   • Repositories with workflows: 2
   • Total workflows: 3
   • Total action usages: 14
   • Unique actions: 6
   • Actions with multiple versions: 1
   • Most used action: actions/checkout (4 usages across 2 repos, 4 workflows)
   ⏱️  Process time: 2.456s
```

### Detailed Analysis - Table Format

```text
╔════════════════════════════════════════════════════════════════════════════════════════════════════════╗
║                               🔍 COMPREHENSIVE ACTION RESULTS                                          ║
╚════════════════════════════════════════════════════════════════════════════════════════════════════════╝
  🏢 Organization: myorg                                                                                  
  📁 Total Repositories: 10                                                                              
  ⚙️  Repositories with Workflows: 2                                                                      
  📄 Total Workflows: 3                                                                                   
  🎯 Unique Actions: 6                                                                                    
  📈 Total Action Usages: 14                                                                              
  ⚠️  Actions with Multiple Versions: 1                                                                   
 🔝 Most Used Action: actions/checkout (4 usages, 2 repos, 4 workflows)                                 
═════════════════════════════════════════════════════════════════════════════════════════════════════════

┌─────────────────────┬──────────────────────────────────┬────────────────────┬─────────┬─────────┬───────┐
│ 📁 REPOSITORY       │ 📄 WORKFLOW                     │ 🔧 ACTION          │ VERSION │ COUNT   │ TOTAL │
├─────────────────────┼──────────────────────────────────┼────────────────────┼─────────┼─────────┼───────┤
│ my-web-app          │ .github/workflows/ci.yml         │ actions/checkout   │ @v4     │ 2       │ 5     │
│                     │                                  │ actions/setup-node │ @v4     │ 1       │       │
│                     │                                  │ actions/upload-... │ @v4     │ 2       │       │
├─────────────────────┼──────────────────────────────────┼────────────────────┼─────────┼─────────┼───────┤
│                     │ .github/workflows/deploy.yml     │ actions/checkout   │ @v4     │ 1       │ 3     │
│                     │                                  │ actions/deploy-... │ @v4     │ 2       │       │
├─────────────────────┼──────────────────────────────────┼────────────────────┼─────────┼─────────┼───────┤
│ api-service         │ .github/workflows/test.yml       │ actions/checkout   │ @v4     │ 1       │ 6     │
│                     │                                  │ actions/setup-go   │ @v5     │ 1       │       │
│                     │                                  │ actions/cache      │ @v4     │ 2       │       │
│                     │                                  │ codecov/codecov... │ @v4     │ 2       │       │
└─────────────────────┴──────────────────────────────────┴────────────────────┴─────────┴─────────┴───────┘

🎯 Summary: 2 repositories, 3 workflows, 6 unique actions, 14 total usages
```

### Detailed Analysis - JSON Format

```json
{
  "organization": "myorg",
  "scan_timestamp": "2024-01-15T10:30:00Z",
  "repositories": [
    {
      "name": "my-web-app",
      "workflow_count": 2,
      "workflows": [
        {
          "path": ".github/workflows/ci.yml",
          "action_count": 3,
          "total_action_count": 5,
          "actions": [
            {
              "name": "actions/checkout",
              "version": "v4",
              "count": 2
            },
            {
              "name": "actions/setup-node",
              "version": "v4",
              "count": 1
            },
            {
              "name": "actions/upload-artifact",
              "version": "v4",
              "count": 2
            }
          ]
        },
        {
          "path": ".github/workflows/deploy.yml",
          "action_count": 2,
          "total_action_count": 3,
          "actions": [
            {
              "name": "actions/checkout",
              "version": "v4",
              "count": 1
            },
            {
              "name": "actions/deploy-pages",
              "version": "v4",
              "count": 2
            }
          ]
        }
      ]
    },
    {
      "name": "api-service",
      "workflow_count": 1,
      "workflows": [
        {
          "path": ".github/workflows/test.yml",
          "action_count": 4,
          "total_action_count": 6,
          "actions": [
            {
              "name": "actions/checkout",
              "version": "v4",
              "count": 1
            },
            {
              "name": "actions/setup-go",
              "version": "v5",
              "count": 1
            },
            {
              "name": "actions/cache",
              "version": "v4",
              "count": 2
            },
            {
              "name": "codecov/codecov-action",
              "version": "v4",
              "count": 2
            }
          ]
        }
      ]
    }
  ],
  "summary": {
    "total_repositories": 10,
    "repositories_with_workflows": 2,
    "total_workflows": 3,
    "total_action_usages": 14,
    "unique_actions": 6,
    "actions_with_multiple_versions": 1,
    "most_used_action": {
      "name": "actions/checkout",
      "total_usages": 4,
      "repositories_using": 2,
      "workflows_using": 4
    }
  },
  "process_time_seconds": 2.456
}
```

## Development

### Building

```bash
go build -o gh-action-lens main.go
```

### Running locally

```bash
./gh-action-lens
```

### Dependencies

This project uses:

- [go-gh](https://github.com/cli/go-gh) v2.12.2 - GitHub CLI library for Go
- [githubv4](https://github.com/shurcooL/githubv4) v0.0.0-20240429030203-be2daab69064 - GitHub GraphQL API client  
- [oauth2](https://golang.org/x/oauth2) v0.23.0 - OAuth2 authentication support
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) v3.0.1 - YAML parsing for workflow files
- Go standard library (encoding/json, fmt, regexp, strings, time, etc.)

### Project Structure

```text
gh-action-lens/
├── main.go          # Main application entry point
├── go.mod           # Go module definition
├── go.sum           # Go module checksums
├── README.md        # User documentation
└── TECHNICAL.md     # Technical documentation (this file)
```

### Architecture

The extension uses a modular approach:

1. **Authentication**: Leverages GitHub CLI credentials or environment variables
2. **API Integration**: Uses GitHub GraphQL API for repository discovery and REST API for file content
3. **Data Processing**: Parses YAML workflow files and extracts action usage patterns
4. **Output Formatting**: Supports multiple output formats (default tree, table, JSON, CSV)
5. **File I/O**: Supports writing results to files for further processing

### Key Functions

- `scanOrganizationWorkflows()`: Discovers repositories with workflow files
- `comprehensiveAnalysis()`: Performs detailed analysis of workflows and actions
- `extractActionsFromFile()`: Parses individual workflow files to extract actions
- `parseActionsFromYAML()`: YAML parsing logic to find action usage patterns
- Output functions: `outputScanResult()`, `outputComprehensiveReport()` with format-specific implementations

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Resources

- [GitHub CLI Extensions Documentation](https://docs.github.com/en/github-cli/github-cli/using-github-cli-extensions)
- [go-gh Examples](https://github.com/cli/go-gh/blob/trunk/example_gh_test.go)
- [GitHub REST API Documentation](https://docs.github.com/en/rest)
- [GitHub GraphQL API Documentation](https://docs.github.com/en/graphql)
