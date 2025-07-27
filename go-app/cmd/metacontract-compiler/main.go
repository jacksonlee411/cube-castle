// cmd/metacontract-compiler/main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	
	"github.com/gaogu/cube-castle/go-app/internal/metacontract"
)

const (
	version = "v6.0.0"
	banner = `
╔═══════════════════════════════════════════════════════════════╗
║                Meta-Contract Compiler %s                    ║
║                  Schema-as-Code for Cube Castle               ║
╚═══════════════════════════════════════════════════════════════╝
`
)

func main() {
	// Command line flags
	var (
		inputFile   = flag.String("input", "", "Path to meta-contract YAML file")
		outputDir   = flag.String("output", "./generated", "Output directory for generated code")
		validate    = flag.Bool("validate", false, "Only validate the meta-contract without generating code")
		verbose     = flag.Bool("verbose", false, "Enable verbose output")
		showVersion = flag.Bool("version", false, "Show version information")
		help        = flag.Bool("help", false, "Show help information")
	)
	
	flag.Parse()
	
	// Handle special flags
	if *showVersion {
		fmt.Printf("Meta-Contract Compiler %s\n", version)
		fmt.Println("Built for Cube Castle Employee Model System")
		os.Exit(0)
	}
	
	if *help || *inputFile == "" {
		printUsage()
		os.Exit(0)
	}
	
	// Print banner
	if *verbose {
		fmt.Printf(banner, version)
	}
	
	// Initialize compiler
	compiler := metacontract.NewCompiler()
	
	// Validate input file exists
	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		log.Fatalf("Meta-contract file does not exist: %s", *inputFile)
	}
	
	// Parse and validate meta-contract
	if *verbose {
		fmt.Printf("📄 Parsing meta-contract: %s\n", *inputFile)
	}
	
	contract, err := compiler.ParseMetaContract(*inputFile)
	if err != nil {
		log.Fatalf("Failed to parse meta-contract: %v", err)
	}
	
	if *verbose {
		fmt.Printf("✅ Successfully parsed meta-contract for resource: %s\n", contract.ResourceName)
		fmt.Printf("   Namespace: %s\n", contract.Namespace)
		fmt.Printf("   Version: %s\n", contract.Version)
		fmt.Printf("   Fields: %d\n", len(contract.DataStructure.Fields))
		fmt.Printf("   Relationships: %d\n", len(contract.Relationships))
	}
	
	// If only validation is requested, exit here
	if *validate {
		fmt.Println("✅ Meta-contract validation passed!")
		os.Exit(0)
	}
	
	// Ensure output directory exists
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}
	
	// Generate code
	if *verbose {
		fmt.Printf("🔧 Generating code to: %s\n", *outputDir)
	}
	
	if err := compiler.Compile(*inputFile, *outputDir); err != nil {
		log.Fatalf("Code generation failed: %v", err)
	}
	
	// Success message
	fmt.Printf("🎉 Successfully generated code for %s!\n", contract.ResourceName)
	if *verbose {
		fmt.Printf("   Ent Schema: %s/schema/%s.go\n", *outputDir, contract.ResourceName)
		fmt.Printf("   API Handler: %s/api/%s_handler.go\n", *outputDir, contract.ResourceName)
	}
	
	// Print summary
	printGenerationSummary(contract, *outputDir)
}

func printUsage() {
	fmt.Printf(banner, version)
	fmt.Println("Usage: metacontract-compiler [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -input string")
	fmt.Println("        Path to meta-contract YAML file (required)")
	fmt.Println("  -output string")
	fmt.Println("        Output directory for generated code (default: ./generated)")
	fmt.Println("  -validate")
	fmt.Println("        Only validate the meta-contract without generating code")
	fmt.Println("  -verbose")
	fmt.Println("        Enable verbose output")
	fmt.Println("  -version")
	fmt.Println("        Show version information")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Validate a meta-contract")
	fmt.Println("  metacontract-compiler -input person.yaml -validate")
	fmt.Println()
	fmt.Println("  # Generate code from meta-contract")
	fmt.Println("  metacontract-compiler -input person.yaml -output ./generated")
	fmt.Println()
	fmt.Println("  # Generate with verbose output")
	fmt.Println("  metacontract-compiler -input person.yaml -output ./generated -verbose")
	fmt.Println()
	fmt.Println("Meta-Contract Specification:")
	fmt.Println("  The meta-contract YAML file should follow the Cube Castle")
	fmt.Println("  meta-contract v6.0 specification for employee model entities.")
	fmt.Println()
}

func printGenerationSummary(contract *metacontract.MetaContract, outputDir string) {
	fmt.Println()
	fmt.Println("📋 Generation Summary:")
	fmt.Printf("   Resource: %s (%s)\n", contract.ResourceName, contract.Namespace)
	fmt.Printf("   Security: %s", contract.SecurityModel.AccessControl)
	if contract.SecurityModel.DataClassification != "" {
		fmt.Printf(" (%s)", contract.SecurityModel.DataClassification)
	}
	fmt.Println()
	
	if contract.TemporalBehavior.TemporalityParadigm != "" {
		fmt.Printf("   Temporal: %s", contract.TemporalBehavior.TemporalityParadigm)
		if contract.TemporalBehavior.StateTransitionModel != "" {
			fmt.Printf(" + %s", contract.TemporalBehavior.StateTransitionModel)
		}
		fmt.Println()
	}
	
	fmt.Printf("   Generated Files:\n")
	
	// List generated files
	schemaDir := filepath.Join(outputDir, "schema")
	apiDir := filepath.Join(outputDir, "api")
	
	if _, err := os.Stat(schemaDir); err == nil {
		fmt.Printf("     📁 %s/\n", schemaDir)
		if files, err := os.ReadDir(schemaDir); err == nil {
			for _, file := range files {
				if !file.IsDir() {
					fmt.Printf("       📄 %s\n", file.Name())
				}
			}
		}
	}
	
	if _, err := os.Stat(apiDir); err == nil {
		fmt.Printf("     📁 %s/\n", apiDir)
		if files, err := os.ReadDir(apiDir); err == nil {
			for _, file := range files {
				if !file.IsDir() {
					fmt.Printf("       📄 %s\n", file.Name())
				}
			}
		}
	}
	
	fmt.Println()
	fmt.Println("🚀 Next Steps:")
	fmt.Println("   1. Run 'go generate ./...' to generate Ent client code")
	fmt.Println("   2. Update your main.go to register the new routes")
	fmt.Println("   3. Run database migrations if needed")
	fmt.Println("   4. Test the generated API endpoints")
	fmt.Println()
}