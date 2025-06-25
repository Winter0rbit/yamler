#!/bin/bash

# Yamler Examples Runner
# This script runs all example programs to demonstrate Yamler functionality

echo "🚀 Running all Yamler examples..."
echo "=================================="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counter for tracking results
total_examples=0
successful_examples=0
failed_examples=0

# Function to run an example
run_example() {
    local example_dir=$1
    local example_name=$2
    local description=$3
    
    echo -e "${BLUE}📂 Running: ${example_name}${NC}"
    echo -e "${YELLOW}   ${description}${NC}"
    echo "   Directory: ${example_dir}/"
    echo
    
    if [ -d "${example_dir}" ] && [ -f "${example_dir}/main.go" ]; then
        cd "${example_dir}" || exit 1
        
        # Run go mod tidy to ensure dependencies
        go mod tidy > /dev/null 2>&1
        
        # Run the example
        if go run main.go; then
            echo -e "${GREEN}✅ ${example_name} completed successfully${NC}"
            ((successful_examples++))
        else
            echo -e "${RED}❌ ${example_name} failed${NC}"
            ((failed_examples++))
        fi
        
        cd - > /dev/null || exit 1
    else
        echo -e "${RED}❌ ${example_name} not found or missing main.go${NC}"
        ((failed_examples++))
    fi
    
    ((total_examples++))
    echo
    echo "----------------------------------------"
    echo
}

# Run all examples in recommended learning order
echo "Running examples in recommended learning order:"
echo

# Beginner Examples
echo -e "${YELLOW}🎯 BEGINNER EXAMPLES${NC}"
echo

run_example "basic_usage" "Basic Usage" "Fundamental operations and type-safe getters"
run_example "comment_alignment" "Comment Alignment" "Flexible comment formatting control"
run_example "file_operations" "File Operations" "File system integration and merging"

# Intermediate Examples
echo -e "${YELLOW}🚀 INTERMEDIATE EXAMPLES${NC}"
echo

run_example "docker_compose" "Docker Compose" "Real-world container orchestration"
run_example "kubernetes" "Kubernetes" "Manifest manipulation and scaling"
run_example "wildcard_patterns" "Wildcard Patterns" "Bulk operations and pattern matching"

# Advanced Examples
echo -e "${YELLOW}🔥 ADVANCED EXAMPLES${NC}"
echo

run_example "ansible" "Ansible" "Playbook management (array-root documents)"
run_example "advanced_performance" "Advanced Performance" "Performance optimization features"
run_example "real_world_use_cases" "Real-World Use Cases" "Production-ready scenarios"

# Summary
echo "=========================================="
echo -e "${BLUE}📊 EXECUTION SUMMARY${NC}"
echo "=========================================="
echo
echo -e "Total examples: ${total_examples}"
echo -e "${GREEN}Successful: ${successful_examples}${NC}"
echo -e "${RED}Failed: ${failed_examples}${NC}"
echo

if [ $failed_examples -eq 0 ]; then
    echo -e "${GREEN}🎉 All examples completed successfully!${NC}"
    echo
    echo "🎯 What you've learned:"
    echo "  ✅ Format preservation with 100% fidelity"
    echo "  ✅ Type-safe operations for all data types"
    echo "  ✅ Comment alignment and formatting control"
    echo "  ✅ Wildcard patterns for bulk operations"
    echo "  ✅ Real-world DevOps configuration management"
    echo "  ✅ Performance optimization techniques"
    echo "  ✅ Complex flow object and array handling"
    echo "  ✅ Production-ready use cases"
    echo
    echo "🚀 Ready to use Yamler in your projects!"
    exit 0
else
    echo -e "${RED}⚠️  Some examples failed. Please check the output above.${NC}"
    exit 1
fi 