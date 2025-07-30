#!/bin/bash
# Documentation Maintenance Script
# Purpose: Automated documentation cleanup and organization

set -e

echo "🔍 Starting documentation maintenance..."

PROJECT_ROOT="/home/shangmeilin/cube-castle"
DOCS_DIR="$PROJECT_ROOT/docs"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Function to check directory structure compliance
check_directory_structure() {
    echo "📁 Checking directory structure compliance..."
    
    expected_dirs=("api" "architecture" "deployment" "reports" "troubleshooting")
    
    for dir in "${expected_dirs[@]}"; do
        if [ ! -d "$DOCS_DIR/$dir" ]; then
            echo -e "${RED}❌ Missing directory: docs/$dir${NC}"
        else
            echo -e "${GREEN}✅ Found directory: docs/$dir${NC}"
        fi
    done
}

# Function to count files by category
generate_statistics() {
    echo "📊 Generating documentation statistics..."
    
    total_md_files=$(find "$PROJECT_ROOT" -name "*.md" ! -path "*/node_modules/*" | wc -l)
    docs_dir_files=$(find "$DOCS_DIR" -name "*.md" | wc -l)
    
    echo "📈 Documentation Statistics:"
    echo "  Total Markdown files: $total_md_files"
    echo "  Files in docs/ directory: $docs_dir_files"
    echo "  Files by category:"
    
    for dir in api architecture deployment reports troubleshooting; do
        if [ -d "$DOCS_DIR/$dir" ]; then
            count=$(find "$DOCS_DIR/$dir" -name "*.md" | wc -l)
            echo "    $dir: $count files"
        fi
    done
}

# Main execution
main() {
    echo "🚀 Documentation Maintenance Report - $(date)"
    echo "=================================================="
    
    check_directory_structure
    echo ""
    
    generate_statistics
    echo ""
    
    echo -e "${GREEN}✅ Documentation maintenance check completed!${NC}"
}

# Run main function
main