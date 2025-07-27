package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Output defines the strategy interface for writing edges
type Output interface {
	Write(edges []string) error
}

// MermaidOutput writes edges to diagram.md in Mermaid format
type MermaidOutput struct{}

func (m MermaidOutput) Write(edges []string) error {
	f, err := os.Create("diagram.md")
	if err != nil {
		fmt.Println("Error creating diagram.md:", err)
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "```mermaid")
	fmt.Fprintln(f, "graph TD")
	for _, edge := range edges {
		fmt.Fprintln(f, edge)
	}
	fmt.Fprintln(f, "```")
	return nil
}

// StdoutOutput writes edges to stdout
type StdoutOutput struct{}

func (s StdoutOutput) Write(edges []string) error {
	fmt.Println("graph TD")
	for _, edge := range edges {
		fmt.Println(edge)
	}
	return nil
}

// Contract represents a contract in the TOML configuration
type Contract struct {
	Name          string   `toml:"name"`
	IgnoreImports []string `toml:"ignore_imports"`
}

// ImportLinter represents the import linter configuration in the TOML file
type ImportLinter struct {
	RootPackages []string   `toml:"root_packages"`
	Contracts    []Contract `toml:"contracts"`
}

// Tool represents the tool configuration in the TOML file
type Tool struct {
	ImportLinter ImportLinter `toml:"importlinter"`
}

// Config represents the entire configuration structure for tools
type Config struct {
	Tool Tool `toml:"tool"`
}

// Flags holds the command line flags
var Flags = struct {
	ConfigFilePath  string
	DestinationType string
	OnlySource      string
	Renderer        string
	Part            int
}{
	ConfigFilePath:  "pyproject.toml",
	DestinationType: "",
	OnlySource:      "",
	Renderer:        "stdout",
	Part:            -1,
}

// Main runs the application
func main() {
	handleFlags()

	config, err := readConfig()
	if err != nil {
		fmt.Println("Error reading configuration:", err)
		os.Exit(1)
	}

	// Collect edges
	var edges []string
	for _, contract := range config.Tool.ImportLinter.Contracts {
		for _, mapping := range contract.IgnoreImports {
			parts := strings.Split(mapping, "->")
			if len(parts) == 2 {
				source := getPartValue(parts[0])
				destination := getPartValue(parts[1])

				if Flags.OnlySource != "" && !strings.Contains(source, Flags.OnlySource) {
					continue
				}

				if Flags.DestinationType != "" && !strings.Contains(destination, Flags.DestinationType) {
					continue
				}

				edge := fmt.Sprintf("%s --> %s", source, destination)

				// Avoid duplicates
				if !strings.Contains(strings.Join(edges, "\n"), edge) {
					edges = append(edges, edge)
				}
			}
		}
	}

	err = handleRendering(edges)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

// handleFlags initializes command line flags
func handleFlags() {
	flag.IntVar(&Flags.Part, "part", -1, "The part we want to focus on e.g. web.x.y.z, if we pick 1, we will show x")
	flag.StringVar(&Flags.ConfigFilePath, "config", "pyproject.toml", "Path to the TOML configuration file")
	flag.StringVar(&Flags.OnlySource, "only-source", "", "Only show edges from this source package")
	flag.StringVar(&Flags.DestinationType, "destination-type", "", "Filter edges by destination type (e.g., 'package', 'module')")
	flag.StringVar(&Flags.Renderer, "renderer", "stdout", "Output renderer: 'stdout' or 'mermaid'")
	flag.Parse()
}

// readConfig reads the TOML configuration file and returns a Config struct
func readConfig() (*Config, error) {
	// Read TOML file
	data, err := os.ReadFile(Flags.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading TOML file: %w", err)
	}

	// Parse Config
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing TOML: %w", err)
	}

	return &config, nil
}

// handleRendering processes the edges and writes them using the specified renderer
func handleRendering(edges []string) error {
	var output Output

	switch Flags.Renderer {
	case "mermaid":
		output = MermaidOutput{}
	case "stdout":
		output = StdoutOutput{}
	default:
		return fmt.Errorf("invalid renderer specified. Use 'stdout' or 'mermaid'")
	}

	return output.Write(edges)
}

// getPartValue extracts the relevant part of the import path based on the
// Flags.Part value
func getPartValue(importPath string) string {
	if Flags.Part < 0 {
		return importPath
	}

	explodedPath := strings.Split(importPath, ".")
	if Flags.Part >= len(explodedPath) {
		return importPath
	}

	return explodedPath[Flags.Part]
}
