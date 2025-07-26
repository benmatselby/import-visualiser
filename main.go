package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

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
	Part           int
	ConfigFilePath string
	OnlySource     string
}{
	Part:           -1,
	ConfigFilePath: "pyproject.toml",
	OnlySource:     "",
}

// Main runs the application
func main() {
	flag.IntVar(&Flags.Part, "part", -1, "The part we want to focus on e.g. web.x.y.z, if we pick 1, we will show x")
	flag.StringVar(&Flags.ConfigFilePath, "config", "pyproject.toml", "Path to the TOML configuration file")
	flag.StringVar(&Flags.OnlySource, "only-source", "", "Only show edges from this source package")
	flag.Parse()

	// Read TOML file
	data, err := os.ReadFile(Flags.ConfigFilePath)
	if err != nil {
		fmt.Println("Error reading TOML file:", err)
		os.Exit(1)
	}

	// Parse Config
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		fmt.Println("Error parsing TOML:", err)
		os.Exit(2)
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

				edge := fmt.Sprintf("%s --> %s", source, destination)

				// Avoid duplicates
				if !strings.Contains(strings.Join(edges, "\n"), edge) {
					edges = append(edges, edge)
				}
			}
		}
	}

	// Render Mermaid
	f, err := os.Create("diagram.md")
	if err != nil {
		fmt.Println("Error creating diagram.md:", err)
		os.Exit(3)
	}
	defer f.Close()

	fmt.Fprintln(f, "```mermaid")
	fmt.Fprintln(f, "graph TD")
	for _, edge := range edges {
		fmt.Fprintln(f, edge)
	}
	fmt.Fprintln(f, "```")
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
