package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

func main() {
	// Define flags
	var showHelp bool
	var organization string
	var scanScope string
	var detailed bool
	var outputFormat string
	var outputFile string

	flag.BoolVar(&showHelp, "help", false, "Show help information")
	flag.BoolVar(&showHelp, "h", false, "Show help information")
	flag.StringVar(&organization, "org", "", "Organization name to target")
	flag.StringVar(&organization, "o", "", "Organization name to target")
	flag.StringVar(&scanScope, "scan", "all", "Scan scope: workflows, actions, or all")
	flag.StringVar(&scanScope, "s", "all", "Scan scope: workflows, actions, or all")
	flag.BoolVar(&detailed, "detailed", false, "Detailed analysis with comprehensive action breakdown")
	flag.BoolVar(&detailed, "d", false, "Detailed analysis with comprehensive action breakdown")
	flag.StringVar(&outputFormat, "format", "default", "Output format: default, json, table, csv")
	flag.StringVar(&outputFormat, "f", "default", "Output format: default, json, table, csv")
	flag.StringVar(&outputFile, "output", "", "Write output to file instead of stdout")

	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n\ngh-action-lens - A GitHub CLI extension for exploring GitHub Actions\n\n")
		fmt.Fprintf(os.Stderr, "This extension analyzes workflow configurations and action declarations, not workflow\n")
		fmt.Fprintf(os.Stderr, "execution history or run logs. It shows you what actions are defined in your workflows\n")
		fmt.Fprintf(os.Stderr, "and how often they're used, but doesn't access runtime data or execution results.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  gh action-lens [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fmt.Fprintf(os.Stderr, "  -h, --help\n")
		fmt.Fprintf(os.Stderr, "        Show help information\n\n")
		fmt.Fprintf(os.Stderr, "  -o, --org <string>\n")
		fmt.Fprintf(os.Stderr, "        Organization name to target\n\n")
		fmt.Fprintf(os.Stderr, "  -s, --scan <string>\n")
		fmt.Fprintf(os.Stderr, "        Scan scope: workflows, actions, or all (default \"all\")\n\n")
		fmt.Fprintf(os.Stderr, "  -d, --detailed\n")
		fmt.Fprintf(os.Stderr, "        Detailed analysis with comprehensive action breakdown\n\n")
		fmt.Fprintf(os.Stderr, "  -f, --format <string>\n")
		fmt.Fprintf(os.Stderr, "        Output format: default, json, table, csv (default \"default\")\n\n")
		fmt.Fprintf(os.Stderr, "      --output <string>\n")
		fmt.Fprintf(os.Stderr, "        Write output to file instead of stdout\n")

		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Basic usage\n")
		fmt.Fprintf(os.Stderr, "  gh action-lens                                  # Run with defaults\n")
		fmt.Fprintf(os.Stderr, "  gh action-lens --help                           # Show help message\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  # Target specific organization\n")
		fmt.Fprintf(os.Stderr, "  gh action-lens -o myorg                         # Scan all workflows and actions\n")
		fmt.Fprintf(os.Stderr, "  gh action-lens -o myorg --scan workflows        # Scan workflows only\n")
		fmt.Fprintf(os.Stderr, "  gh action-lens -o myorg --scan actions          # Analyze actions only\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  # Detailed analysis\n")
		fmt.Fprintf(os.Stderr, "  gh action-lens -o myorg --scan all --detailed   # Comprehensive action breakdown\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  # Output formatting\n")
		fmt.Fprintf(os.Stderr, "  gh action-lens -o myorg --format json           # Output results as JSON\n")
		fmt.Fprintf(os.Stderr, "  gh action-lens -o myorg --format csv            # Output results as CSV\n")
		fmt.Fprintf(os.Stderr, "  gh action-lens -o myorg --output results.txt    # Write output to file\n\n\n")
	}

	// Parse command line arguments
	flag.Parse()

	// Show help if requested
	if showHelp {
		flag.Usage()
		return
	}

	// Main extension logic
	fmt.Println("Welcome to gh-action-lens!")
	fmt.Println("A GitHub CLI extension for scanning GitHub Actions workflows.")

	// Display target scope
	if organization != "" {
		fmt.Printf("🎯 Target Organization: %s\n", organization)
	} else {
		fmt.Println("📍 Scope: Current user context")
	}

	client, err := api.DefaultRESTClient()
	if err != nil {
		fmt.Printf("Error creating GitHub client: %v\n", err)
		return
	}

	response := struct{ Login string }{}
	err = client.Get("user", &response)
	if err != nil {
		fmt.Printf("Error getting user info: %v\n", err)
		return
	}

	fmt.Printf("✓ Authenticated as: %s\n", response.Login)

	// Execute workflow scanning and/or action extraction if requested
	if organization != "" {
		// Validate scan scope
		if scanScope != "workflows" && scanScope != "actions" && scanScope != "all" {
			fmt.Printf("❌ Error: Invalid scan scope '%s'. Valid options: workflows, actions, all.\n", scanScope)
			os.Exit(1)
		}

		// Validate output format
		if outputFormat != "default" && outputFormat != "json" && outputFormat != "table" && outputFormat != "csv" {
			fmt.Printf("❌ Error: Invalid output format '%s'. Valid options: default, json, table, csv.\n", outputFormat)
			os.Exit(1)
		}

		startTime := time.Now()

		switch scanScope {
		case "workflows":
			err := scanOrganizationWorkflows(organization, startTime, outputFormat, outputFile)
			if err != nil {
				fmt.Printf("❌ Error scanning workflows: %v\n", err)
				os.Exit(1)
			}

		case "actions":
			if detailed {
				if outputFormat == "default" {
					fmt.Printf("\n🔍 Detailed action analysis of organization: %s\n\n", organization)
				}
				err := comprehensiveAnalysis(organization, startTime, outputFormat, outputFile)
				if err != nil {
					fmt.Printf("❌ Error: %v\n", err)
					os.Exit(1)
				}
			} else {
				if outputFormat == "default" {
					fmt.Println("\n🔍 Extracting actions from workflows...")
				}
				err := extractActionsFromWorkflows(organization, startTime, outputFormat, outputFile)
				if err != nil {
					fmt.Printf("❌ Error extracting actions: %v\n", err)
					os.Exit(1)
				}
			}

		case "all":
			if detailed {
				if outputFormat == "default" {
					fmt.Println("\n🔍 Starting detailed analysis...")
				}
				err := comprehensiveAnalysis(organization, startTime, outputFormat, outputFile)
				if err != nil {
					fmt.Printf("❌ Error: %v\n", err)
					os.Exit(1)
				}
			} else {
				if outputFormat == "default" {
					fmt.Println("\n🔍 Starting workflow scan and action extraction...")
				}
				err := scanAndExtractActions(organization, startTime, outputFormat, outputFile)
				if err != nil {
					fmt.Printf("❌ Error: %v\n", err)
					os.Exit(1)
				}
			}
		}
		return
	}

	// Show configuration summary
	fmt.Println("\n--- Configuration ---")
	if organization != "" {
		fmt.Printf("Organization: %s\n", organization)
	}
	fmt.Println("\nUse 'gh action-lens --help' to see available options.")
	fmt.Println("\nExamples:")
	fmt.Println("  gh action-lens -o <organization>                  # Scan workflows and actions")
	fmt.Println("  gh action-lens -o <organization> --scan workflows # Scan workflows only")
	fmt.Println("  gh action-lens -o <organization> --scan actions   # Analyze actions only")
}

// scanOrganizationWorkflows scans an organization for repositories with workflow files
func scanOrganizationWorkflows(org string, startTime time.Time, outputFormat string, outputFile string) error {
	// Get GitHub token from environment
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		// Try to get token from gh CLI configuration
		token = os.Getenv("GH_TOKEN")
	}

	if token == "" {
		return fmt.Errorf("GitHub token not found. Please set GITHUB_TOKEN or GH_TOKEN environment variable, or authenticate with 'gh auth login'")
	}

	// Create OAuth2 client
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	// Define GraphQL query structure
	var q struct {
		Organization struct {
			Repositories struct {
				Nodes []struct {
					Name      string
					Workflows struct {
						Tree struct {
							Entries []struct {
								Name string
								Path string
								Type string
							}
						} `graphql:"... on Tree"`
					} `graphql:"workflows: object(expression: \"HEAD:.github/workflows\")"`
				}
				PageInfo struct {
					HasNextPage bool
					EndCursor   githubv4.String
				}
			} `graphql:"repositories(first: 50, after: $cursor)"`
		} `graphql:"organization(login: $org)"`
	}

	vars := map[string]interface{}{
		"org":    githubv4.String(org),
		"cursor": (*githubv4.String)(nil),
	}

	if outputFormat == "default" {
		fmt.Printf("🔍 Scanning organization: %s\n\n", org)
	}

	var repositories []RepositoryWorkflows
	totalRepos := 0
	reposWithWorkflows := 0

	for {
		err := client.Query(context.Background(), &q, vars)
		if err != nil {
			return fmt.Errorf("GraphQL query failed: %v", err)
		}

		for _, repo := range q.Organization.Repositories.Nodes {
			totalRepos++

			if len(repo.Workflows.Tree.Entries) == 0 {
				continue
			}

			var workflowFiles []string
			for _, entry := range repo.Workflows.Tree.Entries {
				if entry.Type == "blob" && (strings.HasSuffix(entry.Name, ".yml") || strings.HasSuffix(entry.Name, ".yaml")) {
					workflowFiles = append(workflowFiles, entry.Path)
				}
			}

			if len(workflowFiles) > 0 {
				reposWithWorkflows++
				repositories = append(repositories, RepositoryWorkflows{
					Name:      repo.Name,
					Workflows: workflowFiles,
				})

				// Repository data will be output later by outputScanResult
			}
		}

		if !q.Organization.Repositories.PageInfo.HasNextPage {
			break
		}
		vars["cursor"] = githubv4.NewString(q.Organization.Repositories.PageInfo.EndCursor)
	}

	duration := time.Since(startTime)

	// Output in requested format
	result := ScanResult{
		Organization:              org,
		TotalRepositories:         totalRepos,
		RepositoriesWithWorkflows: reposWithWorkflows,
		Repositories:              repositories,
		ProcessTimeSeconds:        duration.Seconds(),
	}

	// Get the appropriate writer (file or stdout)
	writer, file, err := getOutputWriter(outputFile)
	if err != nil {
		return fmt.Errorf("error opening output file: %v", err)
	}
	if file != nil {
		defer file.Close()
	}

	return outputScanResult(result, outputFormat, writer)
}

// extractActionsFromWorkflows scans workflows and extracts all actions used
func extractActionsFromWorkflows(org string, startTime time.Time, outputFormat, outputFile string) error {
	workflows, err := getWorkflowFiles(org)
	if err != nil {
		return err
	}

	actionMap := make(map[string]map[string]int) // action -> version -> count
	totalWorkflows := 0

	fmt.Printf("📊 Analyzing %d workflow files...\n\n", len(workflows))

	for _, wf := range workflows {
		totalWorkflows++
		actions, err := extractActionsFromFile(org, wf.Repo, wf.Path)
		if err != nil {
			fmt.Printf("⚠️  Warning: Could not analyze %s/%s: %v\n", wf.Repo, wf.Path, err)
			continue
		}

		for _, action := range actions {
			if actionMap[action.Name] == nil {
				actionMap[action.Name] = make(map[string]int)
			}
			actionMap[action.Name][action.Version]++
		}
	}

	// Generate report
	generateActionReport(actionMap, totalWorkflows, startTime, outputFormat, outputFile)
	return nil
}

// comprehensiveAnalysis performs comprehensive analysis of repositories, workflows, and actions
func comprehensiveAnalysis(org string, startTime time.Time, outputFormat string, outputFile string) error {
	// Get GitHub token from environment
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}

	if token == "" {
		return fmt.Errorf("GitHub token not found. Please set GITHUB_TOKEN or GH_TOKEN environment variable, or authenticate with 'gh auth login'")
	}

	// Create OAuth2 client
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	// Define GraphQL query structure
	var q struct {
		Organization struct {
			Repositories struct {
				Nodes []struct {
					Name      string
					Workflows struct {
						Tree struct {
							Entries []struct {
								Name string
								Path string
								Type string
							}
						} `graphql:"... on Tree"`
					} `graphql:"workflows: object(expression: \"HEAD:.github/workflows\")"`
				}
				PageInfo struct {
					HasNextPage bool
					EndCursor   githubv4.String
				}
			} `graphql:"repositories(first: 50, after: $cursor)"`
		} `graphql:"organization(login: $org)"`
	}

	vars := map[string]interface{}{
		"org":    githubv4.String(org),
		"cursor": (*githubv4.String)(nil),
	}

	var repositories []ComprehensiveRepository
	totalRepos := 0
	reposWithWorkflows := 0
	totalWorkflows := 0
	actionUsageMap := make(map[string]map[string]int) // action -> version -> count
	actionRepoMap := make(map[string]map[string]bool) // action -> repo -> true
	actionWorkflowMap := make(map[string]int)         // action -> workflow count

	// Scan repositories
	for {
		err := client.Query(context.Background(), &q, vars)
		if err != nil {
			return fmt.Errorf("GraphQL query failed: %v", err)
		}

		for _, repo := range q.Organization.Repositories.Nodes {
			totalRepos++

			if len(repo.Workflows.Tree.Entries) == 0 {
				continue
			}

			var workflowFiles []string
			for _, entry := range repo.Workflows.Tree.Entries {
				if entry.Type == "blob" && (strings.HasSuffix(entry.Name, ".yml") || strings.HasSuffix(entry.Name, ".yaml")) {
					workflowFiles = append(workflowFiles, entry.Path)
				}
			}

			if len(workflowFiles) == 0 {
				continue
			}

			reposWithWorkflows++
			totalWorkflows += len(workflowFiles)

			// Analyze workflows in this repository
			var workflows []ComprehensiveWorkflow
			for _, workflowPath := range workflowFiles {
				actions, err := extractActionsFromFile(org, repo.Name, workflowPath)
				if err != nil {
					if outputFormat == "default" {
						fmt.Printf("⚠️  Warning: Could not analyze %s/%s: %v\n", repo.Name, workflowPath, err)
					}
					continue
				}

				// Deduplicate actions within this workflow and count occurrences
				actionCounts := make(map[string]map[string]int) // action -> version -> count
				for _, action := range actions {
					if actionCounts[action.Name] == nil {
						actionCounts[action.Name] = make(map[string]int)
					}
					actionCounts[action.Name][action.Version]++
				}

				// Convert to comprehensive actions with counts
				var comprehensiveActions []ComprehensiveAction
				totalUniqueActions := 0
				for actionName, versions := range actionCounts {
					for version, count := range versions {
						comprehensiveActions = append(comprehensiveActions, ComprehensiveAction{
							Name:    actionName,
							Version: version,
							Count:   count,
						})
						totalUniqueActions += count

						// Track usage statistics
						if actionUsageMap[actionName] == nil {
							actionUsageMap[actionName] = make(map[string]int)
							actionRepoMap[actionName] = make(map[string]bool)
						}
						actionUsageMap[actionName][version] += count
						actionRepoMap[actionName][repo.Name] = true
						actionWorkflowMap[actionName] += count
					}
				}

				workflows = append(workflows, ComprehensiveWorkflow{
					Path:             workflowPath,
					ActionCount:      len(comprehensiveActions),
					TotalActionCount: totalUniqueActions,
					Actions:          comprehensiveActions,
				})

				if outputFormat == "default" {
					if len(comprehensiveActions) == totalUniqueActions {
						fmt.Printf("📁 %s → 📄 %s (%d actions)\n", repo.Name, workflowPath, len(comprehensiveActions))
					} else {
						fmt.Printf("📁 %s → 📄 %s (%d unique, %d total actions)\n", repo.Name, workflowPath, len(comprehensiveActions), totalUniqueActions)
					}
				}
			}

			repositories = append(repositories, ComprehensiveRepository{
				Name:          repo.Name,
				WorkflowCount: len(workflowFiles),
				Workflows:     workflows,
			})
		}

		if !q.Organization.Repositories.PageInfo.HasNextPage {
			break
		}
		vars["cursor"] = githubv4.NewString(q.Organization.Repositories.PageInfo.EndCursor)
	}

	// Generate comprehensive summary
	uniqueActions := len(actionUsageMap)
	totalActionUsages := 0
	actionsWithMultipleVersions := 0
	var mostUsedAction ComprehensiveMostUsedAction

	for actionName, versions := range actionUsageMap {
		actionTotal := 0
		for _, count := range versions {
			actionTotal += count
		}
		totalActionUsages += actionTotal

		if len(versions) > 1 {
			actionsWithMultipleVersions++
		}

		// Track most used action
		if actionTotal > mostUsedAction.TotalUsages {
			mostUsedAction = ComprehensiveMostUsedAction{
				Name:              actionName,
				TotalUsages:       actionTotal,
				RepositoriesUsing: len(actionRepoMap[actionName]),
				WorkflowsUsing:    actionWorkflowMap[actionName],
			}
		}
	}

	duration := time.Since(startTime)

	// Create comprehensive report
	report := ComprehensiveReport{
		Organization:  org,
		ScanTimestamp: startTime.Format(time.RFC3339),
		Repositories:  repositories,
		Summary: ComprehensiveSummary{
			TotalRepositories:           totalRepos,
			RepositoriesWithWorkflows:   reposWithWorkflows,
			TotalWorkflows:              totalWorkflows,
			TotalActionUsages:           totalActionUsages,
			UniqueActions:               uniqueActions,
			ActionsWithMultipleVersions: actionsWithMultipleVersions,
			MostUsedAction:              mostUsedAction,
		},
		ProcessTimeSeconds: duration.Seconds(),
	}

	// Get the appropriate writer (file or stdout)
	writer, file, err := getOutputWriter(outputFile)
	if err != nil {
		return fmt.Errorf("error opening output file: %v", err)
	}
	if file != nil {
		defer file.Close()
	}

	return outputComprehensiveReport(report, outputFormat, writer)
}

// getOutputWriter returns the appropriate writer based on the output file flag
func getOutputWriter(outputFile string) (io.Writer, *os.File, error) {
	if outputFile == "" {
		return os.Stdout, nil, nil
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return nil, nil, err
	}

	return file, file, nil
}

// scanAndExtractActions combines scanning and action extraction
func scanAndExtractActions(org string, startTime time.Time, outputFormat, outputFile string) error {
	if outputFormat == "default" {
		fmt.Println("Phase 1: Scanning for workflow files...")
	}
	err := scanOrganizationWorkflows(org, startTime, outputFormat, "")
	if err != nil {
		return fmt.Errorf("scanning failed: %v", err)
	}

	if outputFormat == "default" {
		fmt.Println("\nPhase 2: Extracting actions from workflows...")
	}
	err = extractActionsFromWorkflows(org, startTime, outputFormat, outputFile)
	if err != nil {
		return fmt.Errorf("action extraction failed: %v", err)
	}

	return nil
}

// WorkflowFile represents a workflow file in a repository
type WorkflowFile struct {
	Repo string `json:"repo"`
	Path string `json:"path"`
}

// Action represents a GitHub Action usage
type Action struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ScanResult represents the output of a workflow scan
type ScanResult struct {
	Organization              string                `json:"organization"`
	TotalRepositories         int                   `json:"total_repositories"`
	RepositoriesWithWorkflows int                   `json:"repositories_with_workflows"`
	Repositories              []RepositoryWorkflows `json:"repositories"`
	ProcessTimeSeconds        float64               `json:"process_time_seconds"`
}

// RepositoryWorkflows represents a repository and its workflow files
type RepositoryWorkflows struct {
	Name      string   `json:"name"`
	Workflows []string `json:"workflows"`
}

// ActionReport represents the output of action extraction
type ActionReport struct {
	Organization       string          `json:"organization"`
	TotalWorkflows     int             `json:"total_workflows"`
	UniqueActions      int             `json:"unique_actions"`
	TotalUsages        int             `json:"total_usages"`
	Actions            []ActionSummary `json:"actions"`
	ProcessTimeSeconds float64         `json:"process_time_seconds"`
}

// ActionSummary represents an action and its usage statistics
type ActionSummary struct {
	Name     string         `json:"name"`
	Total    int            `json:"total_usages"`
	Versions []VersionUsage `json:"versions"`
}

// VersionUsage represents version usage statistics
type VersionUsage struct {
	Version string `json:"version"`
	Count   int    `json:"count"`
}

// ComprehensiveReport represents the comprehensive analysis output
type ComprehensiveReport struct {
	Organization       string                    `json:"organization"`
	ScanTimestamp      string                    `json:"scan_timestamp"`
	Repositories       []ComprehensiveRepository `json:"repositories"`
	Summary            ComprehensiveSummary      `json:"summary"`
	ProcessTimeSeconds float64                   `json:"process_time_seconds"`
}

// ComprehensiveRepository represents a repository with its workflows and actions
type ComprehensiveRepository struct {
	Name          string                  `json:"name"`
	WorkflowCount int                     `json:"workflow_count"`
	Workflows     []ComprehensiveWorkflow `json:"workflows"`
}

// ComprehensiveWorkflow represents a workflow file with its actions
type ComprehensiveWorkflow struct {
	Path             string                `json:"path"`
	ActionCount      int                   `json:"action_count"`       // Number of unique actions
	TotalActionCount int                   `json:"total_action_count"` // Total action occurrences
	Actions          []ComprehensiveAction `json:"actions"`
}

// ComprehensiveAction represents an action usage with metadata
type ComprehensiveAction struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Count   int    `json:"count"`
}

// ComprehensiveSummary represents summary statistics for comprehensive analysis
type ComprehensiveSummary struct {
	TotalRepositories           int                         `json:"total_repositories"`
	RepositoriesWithWorkflows   int                         `json:"repositories_with_workflows"`
	TotalWorkflows              int                         `json:"total_workflows"`
	TotalActionUsages           int                         `json:"total_action_usages"`
	UniqueActions               int                         `json:"unique_actions"`
	ActionsWithMultipleVersions int                         `json:"actions_with_multiple_versions"`
	MostUsedAction              ComprehensiveMostUsedAction `json:"most_used_action"`
}

// ComprehensiveMostUsedAction represents the most frequently used action
type ComprehensiveMostUsedAction struct {
	Name              string `json:"name"`
	TotalUsages       int    `json:"total_usages"`
	RepositoriesUsing int    `json:"repositories_using"`
	WorkflowsUsing    int    `json:"workflows_using"`
}

// getWorkflowFiles retrieves all workflow files from an organization
func getWorkflowFiles(org string) ([]WorkflowFile, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}

	if token == "" {
		return nil, fmt.Errorf("GitHub token not found")
	}

	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	var q struct {
		Organization struct {
			Repositories struct {
				Nodes []struct {
					Name      string
					Workflows struct {
						Tree struct {
							Entries []struct {
								Name string
								Path string
								Type string
							}
						} `graphql:"... on Tree"`
					} `graphql:"workflows: object(expression: \"HEAD:.github/workflows\")"`
				}
				PageInfo struct {
					HasNextPage bool
					EndCursor   githubv4.String
				}
			} `graphql:"repositories(first: 50, after: $cursor)"`
		} `graphql:"organization(login: $org)"`
	}

	vars := map[string]interface{}{
		"org":    githubv4.String(org),
		"cursor": (*githubv4.String)(nil),
	}

	var workflows []WorkflowFile

	for {
		err := client.Query(context.Background(), &q, vars)
		if err != nil {
			return nil, fmt.Errorf("GraphQL query failed: %v", err)
		}

		for _, repo := range q.Organization.Repositories.Nodes {
			for _, entry := range repo.Workflows.Tree.Entries {
				if entry.Type == "blob" && (strings.HasSuffix(entry.Name, ".yml") || strings.HasSuffix(entry.Name, ".yaml")) {
					workflows = append(workflows, WorkflowFile{
						Repo: repo.Name,
						Path: entry.Path,
					})
				}
			}
		}

		if !q.Organization.Repositories.PageInfo.HasNextPage {
			break
		}
		vars["cursor"] = githubv4.NewString(q.Organization.Repositories.PageInfo.EndCursor)
	}

	return workflows, nil
}

// extractActionsFromFile fetches and parses a workflow file to extract actions
func extractActionsFromFile(org, repo, path string) ([]Action, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}

	// Use GitHub REST API to get file content
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", org, repo, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var fileData struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}

	err = json.NewDecoder(resp.Body).Decode(&fileData)
	if err != nil {
		return nil, err
	}

	// Decode base64 content
	var yamlContent string
	if fileData.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(fileData.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 content: %v", err)
		}
		yamlContent = string(decoded)
	} else {
		yamlContent = fileData.Content
	}

	// Parse YAML and extract actions
	return parseActionsFromYAML(yamlContent)
}

// parseActionsFromYAML parses YAML content and extracts GitHub Actions
func parseActionsFromYAML(yamlContent string) ([]Action, error) {
	var workflow map[string]interface{}
	err := yaml.Unmarshal([]byte(yamlContent), &workflow)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	var actions []Action
	usesPattern := regexp.MustCompile(`^([^@]+)@(.+)$`)

	// Recursively search for "uses" fields
	var extractUses func(interface{})
	extractUses = func(obj interface{}) {
		switch v := obj.(type) {
		case map[string]interface{}:
			for key, value := range v {
				if key == "uses" {
					if usesStr, ok := value.(string); ok {
						matches := usesPattern.FindStringSubmatch(usesStr)
						if len(matches) == 3 {
							actions = append(actions, Action{
								Name:    matches[1],
								Version: matches[2],
							})
						}
					}
				} else {
					extractUses(value)
				}
			}
		case []interface{}:
			for _, item := range v {
				extractUses(item)
			}
		}
	}

	extractUses(workflow)
	return actions, nil
}

// generateActionReport creates a summary report of all actions found
func generateActionReport(actionMap map[string]map[string]int, totalWorkflows int, startTime time.Time, outputFormat, outputFile string) {
	// Sort actions by name
	var actionNames []string
	for name := range actionMap {
		actionNames = append(actionNames, name)
	}
	sort.Strings(actionNames)

	// Calculate totals and build action summaries
	totalActions := 0
	var actions []ActionSummary

	for _, name := range actionNames {
		versions := actionMap[name]
		actionTotal := 0
		var versionUsages []VersionUsage

		// Sort versions
		var versionList []string
		for version := range versions {
			versionList = append(versionList, version)
		}
		sort.Strings(versionList)

		for _, version := range versionList {
			count := versions[version]
			actionTotal += count
			versionUsages = append(versionUsages, VersionUsage{
				Version: version,
				Count:   count,
			})
		}

		totalActions += actionTotal
		actions = append(actions, ActionSummary{
			Name:     name,
			Total:    actionTotal,
			Versions: versionUsages,
		})
	}

	duration := time.Since(startTime)

	// Create report data
	report := ActionReport{
		TotalWorkflows:     totalWorkflows,
		UniqueActions:      len(actionNames),
		TotalUsages:        totalActions,
		Actions:            actions,
		ProcessTimeSeconds: duration.Seconds(),
	}

	outputActionReport(report, outputFormat, outputFile)
}

// outputScanResult outputs scan results in the specified format
func outputScanResult(result ScanResult, format string, writer io.Writer) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)

	case "table":
		return outputScanTable(result, writer)

	case "csv":
		return outputScanCSV(result, writer)

	default: // "default"
		// Output repository listing with workflows
		for _, repo := range result.Repositories {
			fmt.Fprintf(writer, "📁 %s:\n", repo.Name)
			for _, wf := range repo.Workflows {
				fmt.Fprintf(writer, "  └─ %s\n", wf)
			}
			fmt.Fprintf(writer, "\n")
		}

		fmt.Fprintf(writer, "✅ Scan complete!\n")
		fmt.Fprintf(writer, "📊 Summary: Found %d repositories with workflows out of %d total repositories.\n",
			result.RepositoriesWithWorkflows, result.TotalRepositories)
		fmt.Fprintf(writer, "⏱️  Process time: %.3fs\n", result.ProcessTimeSeconds)

		return nil
	}
}

// outputScanTable outputs scan results in table format
func outputScanTable(result ScanResult, writer io.Writer) error {
	// Header section with enhanced styling
	fmt.Fprintln(writer, "╔════════════════════════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Fprintf(writer, "║                                       📊 WORKFLOW SCAN RESULTS                                     ║\n")
	fmt.Fprintln(writer, "╚════════════════════════════════════════════════════════════════════════════════════════════════════╝")
	fmt.Fprintf(writer, "  🏢 Organization: %-59s \n", result.Organization)
	fmt.Fprintf(writer, "  📁 Total Repositories: %-53d \n", result.TotalRepositories)
	fmt.Fprintf(writer, "  ⚙️  Repositories with Workflows: %-44d \n", result.RepositoriesWithWorkflows)
	summaryStr := fmt.Sprintf("%d/%d repositories have GitHub Actions workflows (%.1f%%)",
		result.RepositoriesWithWorkflows, result.TotalRepositories,
		float64(result.RepositoriesWithWorkflows)/float64(result.TotalRepositories)*100)
	fmt.Fprintf(writer, "  🎯 Summary: %-64s  \n", summaryStr)
	// processTimeStr := fmt.Sprintf("%.3fs", result.ProcessTimeSeconds)
	// fmt.Fprintf(writer, "  ⏱️  Process Time: %-59s  \n", processTimeStr)
	fmt.Fprintln(writer, " ═════════════════════════════════════════════════════════════════════════════════════════════════════")
	fmt.Fprintln(writer)

	if len(result.Repositories) == 0 {
		fmt.Fprintln(writer, "┌─────────────────────────────────────────┐")
		fmt.Fprintln(writer, "│   No repositories with workflows found │")
		fmt.Fprintln(writer, "└─────────────────────────────────────────┘")
		return nil
	}

	// Table header with borders
	fmt.Fprintln(writer, "┌─────────────────────────────┬─────────────────────────────────────────────────────────────┬─────────┐")
	fmt.Fprintf(writer, "│ %-26s │ %-58s │ %-7s │\n", "📁 REPOSITORY", "📄 WORKFLOW FILES", "COUNT")
	fmt.Fprintln(writer, "├─────────────────────────────┼─────────────────────────────────────────────────────────────┼─────────┤")

	// Table rows
	for i, repo := range result.Repositories {
		workflowList := strings.Join(repo.Workflows, ", ")
		if len(workflowList) > 59 {
			workflowList = workflowList[:56] + "..."
		}

		fmt.Fprintf(writer, "│ %-27s │ %-59s │ %-7d │\n", repo.Name, workflowList, len(repo.Workflows))

		// Add separator between rows (not after last row)
		if i < len(result.Repositories)-1 {
			fmt.Fprintln(writer, "├─────────────────────────────┼─────────────────────────────────────────────────────────────┼─────────┤")
		}
	}

	// Table footer
	fmt.Fprintln(writer, "└─────────────────────────────┴─────────────────────────────────────────────────────────────┴─────────┘")
	fmt.Fprintln(writer)

	return nil
}

// outputScanCSV outputs scan results in CSV format
func outputScanCSV(result ScanResult, writer io.Writer) error {
	// CSV Header
	fmt.Fprintf(writer, "Repository,Workflow Count,Workflow Files\n")

	// CSV Data rows
	for _, repo := range result.Repositories {
		workflowList := strings.Join(repo.Workflows, "; ")
		// Escape quotes in CSV by doubling them
		workflowList = strings.ReplaceAll(workflowList, "\"", "\"\"")

		fmt.Fprintf(writer, "\"%s\",%d,\"%s\"\n", repo.Name, len(repo.Workflows), workflowList)
	}

	return nil
}

// outputActionReport outputs action report in the specified format
func outputActionReport(report ActionReport, format, outputFile string) error {
	// Determine output destination
	var writer io.Writer = os.Stdout
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer file.Close()
		writer = file
	}

	switch format {
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)

	case "table":
		return outputActionTable(report, writer)

	case "csv":
		return outputActionCSV(report, writer)

	default: // "default"
		fmt.Fprintln(writer, "📋 Action Reference Report")
		fmt.Fprintln(writer, "="+strings.Repeat("=", 50))

		for _, action := range report.Actions {
			fmt.Fprintf(writer, "\n🔧 %s (used %d times)\n", action.Name, action.Total)
			for _, version := range action.Versions {
				fmt.Fprintf(writer, "   └─ @%s (%d times)\n", version.Version, version.Count)
			}
		}

		fmt.Fprintln(writer, "\n📊 Summary:")
		fmt.Fprintf(writer, "   • Total workflows analyzed: %d\n", report.TotalWorkflows)
		fmt.Fprintf(writer, "   • Unique actions found: %d\n", report.UniqueActions)
		fmt.Fprintf(writer, "   • Total action usages: %d\n", report.TotalUsages)
		fmt.Fprintf(writer, "   ⏱️  Process time: %.3fs\n", report.ProcessTimeSeconds)

		return nil
	}
}

// outputActionTable outputs action report in table format
func outputActionTable(report ActionReport, writer io.Writer) error {
	// Calculate actions with multiple versions
	multiVersionCount := 0
	for _, action := range report.Actions {
		if len(action.Versions) > 1 {
			multiVersionCount++
		}
	}

	// Header section with enhanced styling
	fmt.Fprintln(writer, "╔════════════════════════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Fprintf(writer, "║                                   🔧 GITHUB ACTIONS SCAN RESULTS                                   ║\n")
	fmt.Fprintln(writer, "╚════════════════════════════════════════════════════════════════════════════════════════════════════╝")
	fmt.Fprintf(writer, "  📊 Total Workflows Analyzed: %-75d \n", report.TotalWorkflows)
	fmt.Fprintf(writer, "  🎯 Unique Actions Found: %-79d \n", report.UniqueActions)
	fmt.Fprintf(writer, "  📈 Total Action Usages: %-80d \n", report.TotalUsages)
	avgUsagePerAction := float64(report.TotalUsages) / float64(report.UniqueActions)
	avgUsageStr := fmt.Sprintf("%.1f", avgUsagePerAction)
	fmt.Fprintf(writer, "  📊 Average usages per action: %-74s \n", avgUsageStr)
	mostUsedStr := fmt.Sprintf("%s (%d usages)", report.Actions[0].Name, report.Actions[0].Total)
	fmt.Fprintf(writer, "  🔝 Most used action: %-83s \n", mostUsedStr)
	fmt.Fprintf(writer, "  ⚠️  Actions with multiple versions: %-69d \n", multiVersionCount)
	// processTimeStr := fmt.Sprintf("%.3fs", report.ProcessTimeSeconds)
	// fmt.Fprintf(writer, "║ ⏱️  Process Time: %-87s ║\n", processTimeStr)
	fmt.Fprintln(writer, " ════════════════════════════════════════════════════════════════════════════════════════════════════")
	fmt.Fprintln(writer)

	if len(report.Actions) == 0 {
		fmt.Fprintln(writer, "┌─────────────────────────┐")
		fmt.Fprintln(writer, "│   No actions found.     │")
		fmt.Fprintln(writer, "└─────────────────────────┘")
		return nil
	}

	// Table header with borders
	fmt.Fprintln(writer, "┌─────────────────────────────────────────────────────────────────────┬─────────────┬─────────┬───────┐")
	fmt.Fprintf(writer, "│ %-66s │ %-10s │ %-7s │ %-5s │\n", "🔧 ACTION NAME", "📦 VERSION", "USAGES", "TOTAL")
	fmt.Fprintln(writer, "├─────────────────────────────────────────────────────────────────────┼─────────────┼─────────┼───────┤")

	// Table rows
	for actionIdx, action := range report.Actions {
		for versionIdx, version := range action.Versions {
			var actionName string
			var totalStr string

			if versionIdx == 0 {
				// First version row shows action name and total
				actionName = action.Name
				if len(actionName) > 67 {
					actionName = actionName[:67] + "..."
				}
				totalStr = fmt.Sprintf("%d", action.Total)
			} else {
				// Subsequent version rows are indented
				actionName = "  └─ " + strings.Repeat(" ", len(action.Name)-6)
				if len(actionName) > 67 {
					actionName = actionName[:67]
				}
				totalStr = ""
			}

			fmt.Fprintf(writer, "│ %-67s │ @%-10s │ %-7d │ %-5s │\n",
				actionName, version.Version, version.Count, totalStr)
		}

		// Add separator between actions (not after last action)
		if actionIdx < len(report.Actions)-1 {
			fmt.Fprintln(writer, "├─────────────────────────────────────────────────────────────────────┼─────────────┼─────────┼───────┤")
		}
	}

	// Table footer
	fmt.Fprintln(writer, "└─────────────────────────────────────────────────────────────────────┴─────────────┴─────────┴───────┘")
	fmt.Fprintln(writer)
	return nil
}

// outputActionCSV outputs action report in CSV format
func outputActionCSV(report ActionReport, writer io.Writer) error {
	fmt.Fprintln(writer, "Action,Version,Usages,Total")

	for _, action := range report.Actions {
		for versionIdx, version := range action.Versions {
			if versionIdx == 0 {
				// First version row includes total
				fmt.Fprintf(writer, "%s,@%s,%d,%d\n", action.Name, version.Version, version.Count, action.Total)
			} else {
				// Subsequent version rows don't repeat total
				fmt.Fprintf(writer, "%s,@%s,%d,\n", action.Name, version.Version, version.Count)
			}
		}
	}
	return nil
}

// outputComprehensiveReport outputs comprehensive report in the specified format
func outputComprehensiveReport(report ComprehensiveReport, format string, writer io.Writer) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)

	case "table":
		return outputComprehensiveTable(report, writer)

	case "csv":
		return outputComprehensiveCSV(report, writer)

	default: // "default"
		fmt.Fprintln(writer, "\n🔍 Detailed Analysis Results")
		fmt.Fprintln(writer, "="+strings.Repeat("=", 60))

		for _, repo := range report.Repositories {
			fmt.Fprintf(writer, "\n📁 %s (%d workflows)\n", repo.Name, repo.WorkflowCount)
			for _, workflow := range repo.Workflows {
				if workflow.ActionCount == workflow.TotalActionCount {
					fmt.Fprintf(writer, "   📄 %s (%d actions)\n", workflow.Path, workflow.ActionCount)
				} else {
					fmt.Fprintf(writer, "   📄 %s (%d unique, %d total actions)\n", workflow.Path, workflow.ActionCount, workflow.TotalActionCount)
				}
				for _, action := range workflow.Actions {
					if action.Count > 1 {
						fmt.Fprintf(writer, "      🔧 %s@%s (%d times)\n", action.Name, action.Version, action.Count)
					} else {
						fmt.Fprintf(writer, "      🔧 %s@%s\n", action.Name, action.Version)
					}
				}
			}
		}

		fmt.Fprintf(writer, "\n📊 Summary:\n")
		fmt.Fprintf(writer, "   • Total repositories: %d\n", report.Summary.TotalRepositories)
		fmt.Fprintf(writer, "   • Repositories with workflows: %d\n", report.Summary.RepositoriesWithWorkflows)
		fmt.Fprintf(writer, "   • Total workflows: %d\n", report.Summary.TotalWorkflows)
		fmt.Fprintf(writer, "   • Total action usages: %d\n", report.Summary.TotalActionUsages)
		fmt.Fprintf(writer, "   • Unique actions: %d\n", report.Summary.UniqueActions)
		fmt.Fprintf(writer, "   • Actions with multiple versions: %d\n", report.Summary.ActionsWithMultipleVersions)
		fmt.Fprintf(writer, "   • Most used action: %s (%d usages across %d repos, %d workflows)\n",
			report.Summary.MostUsedAction.Name,
			report.Summary.MostUsedAction.TotalUsages,
			report.Summary.MostUsedAction.RepositoriesUsing,
			report.Summary.MostUsedAction.WorkflowsUsing)
		fmt.Fprintf(writer, "   ⏱️  Process time: %.3fs\n", report.ProcessTimeSeconds)

		return nil
	}
}

// outputComprehensiveTable outputs comprehensive report in table format
func outputComprehensiveTable(report ComprehensiveReport, writer io.Writer) error {
	// Header section with enhanced styling
	fmt.Fprintln(writer, " ╔════════════════════════════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Fprintf(writer, " ║                               🔍 COMPREHENSIVE ACTION RESULTS                                          ║\n")
	fmt.Fprintln(writer, " ╚════════════════════════════════════════════════════════════════════════════════════════════════════════╝")
	fmt.Fprintf(writer, "  🏢 Organization: %-83s \n", report.Organization)
	fmt.Fprintf(writer, "  📁 Total Repositories: %-77d \n", report.Summary.TotalRepositories)
	fmt.Fprintf(writer, "  ⚙️  Repositories with Workflows: %-69d \n", report.Summary.RepositoriesWithWorkflows)
	fmt.Fprintf(writer, "  📄 Total Workflows: %-80d \n", report.Summary.TotalWorkflows)
	fmt.Fprintf(writer, "  🎯 Unique Actions: %-81d \n", report.Summary.UniqueActions)
	fmt.Fprintf(writer, "  📈 Total Action Usages: %-76d \n", report.Summary.TotalActionUsages)
	fmt.Fprintf(writer, "  ⚠️  Actions with Multiple Versions: %-66d \n", report.Summary.ActionsWithMultipleVersions)
	mostUsedStr := fmt.Sprintf("%s (%d usages, %d repos, %d workflows)",
		report.Summary.MostUsedAction.Name,
		report.Summary.MostUsedAction.TotalUsages,
		report.Summary.MostUsedAction.RepositoriesUsing,
		report.Summary.MostUsedAction.WorkflowsUsing)
	fmt.Fprintf(writer, " 🔝 Most Used Action: %-76s \n", mostUsedStr)
	// processTimeStr := fmt.Sprintf("%.3fs", report.ProcessTimeSeconds)
	// fmt.Fprintf(writer, "  ⏱️  Process Time: %-83s \n", processTimeStr)
	fmt.Fprintln(writer, " ═════════════════════════════════════════════════════════════════════════════════════════════════════════")
	fmt.Fprintln(writer)

	if len(report.Repositories) == 0 {
		fmt.Fprintln(writer, "┌─────────────────────────────────────────┐")
		fmt.Fprintln(writer, "│   No repositories with workflows found  │")
		fmt.Fprintln(writer, "└─────────────────────────────────────────┘")
		return nil
	}

	// Hierarchical table showing repositories → workflows → actions
	fmt.Fprintln(writer, "┌─────────────────────┬──────────────────────────────────┬────────────────────┬─────────┬─────────┬───────┐")
	fmt.Fprintf(writer, "│ %-18s │ %-31s │ %-17s │ %-7s │ %-7s │ %-5s │\n", "📁 REPOSITORY", "📄 WORKFLOW", "🔧 ACTION", "VERSION", "COUNT", "TOTAL")
	fmt.Fprintln(writer, "├─────────────────────┼──────────────────────────────────┼────────────────────┼─────────┼─────────┼───────┤")

	totalRows := 0
	for _, repo := range report.Repositories {
		repoDisplayed := false

		for _, workflow := range repo.Workflows {
			workflowDisplayed := false

			for _, action := range workflow.Actions {
				var repoName, workflowName string

				if !repoDisplayed {
					repoName = repo.Name
					if len(repoName) > 19 {
						repoName = repoName[:16] + "..."
					}
					repoDisplayed = true
				} else {
					repoName = ""
				}

				if !workflowDisplayed {
					workflowName = workflow.Path
					if len(workflowName) > 32 {
						workflowName = workflowName[:29] + "..."
					}
					workflowDisplayed = true
				} else {
					workflowName = ""
				}

				actionName := action.Name
				if len(actionName) > 18 {
					actionName = actionName[:15] + "..."
				}

				var totalCount string
				if workflowName != "" {
					totalCount = fmt.Sprintf("%d", workflow.TotalActionCount)
				} else {
					totalCount = ""
				}

				fmt.Fprintf(writer, "│ %-19s │ %-32s │ %-18s │ @%-6s │ %-7d │ %-5s │\n",
					repoName, workflowName, actionName, action.Version, action.Count, totalCount)

				totalRows++

				// Add separator between actions (not after last action)
				if totalRows < getTotalActionCount(report) {
					fmt.Fprintln(writer, "├─────────────────────┼──────────────────────────────────┼────────────────────┼─────────┼─────────┼───────┤")
				}
			}
		}
	}

	// Table footer
	fmt.Fprintln(writer, "└─────────────────────┴──────────────────────────────────┴────────────────────┴─────────┴─────────┴───────┘")
	fmt.Fprintf(writer, "\n🎯 Summary: %d repositories, %d workflows, %d unique actions, %d total usages\n",
		report.Summary.RepositoriesWithWorkflows, report.Summary.TotalWorkflows,
		report.Summary.UniqueActions, report.Summary.TotalActionUsages)
	fmt.Fprintln(writer)

	return nil
}

// outputComprehensiveCSV outputs comprehensive report in CSV format
func outputComprehensiveCSV(report ComprehensiveReport, writer io.Writer) error {
	// CSV Header
	fmt.Fprintf(writer, "Repository,Workflow,Action,Version,Count,Total\n")

	// CSV Data rows
	for _, repo := range report.Repositories {
		for _, workflow := range repo.Workflows {
			for _, action := range workflow.Actions {
				// Escape quotes in CSV by doubling them
				repoName := strings.ReplaceAll(repo.Name, "\"", "\"\"")
				workflowPath := strings.ReplaceAll(workflow.Path, "\"", "\"\"")
				actionName := strings.ReplaceAll(action.Name, "\"", "\"\"")

				fmt.Fprintf(writer, "\"%s\",\"%s\",\"%s\",\"%s\",%d,%d\n",
					repoName, workflowPath, actionName, action.Version, action.Count, workflow.TotalActionCount)
			}
		}
	}

	return nil
}

// getTotalActionCount calculates the total number of action entries for table formatting
func getTotalActionCount(report ComprehensiveReport) int {
	count := 0
	for _, repo := range report.Repositories {
		for _, workflow := range repo.Workflows {
			count += len(workflow.Actions)
		}
	}
	return count
}

// For more examples of using go-gh, see:
// https://github.com/cli/go-gh/blob/trunk/example_gh_test.go
