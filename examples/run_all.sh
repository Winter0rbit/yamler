#!/bin/bash

echo "Running all Yamler examples..."
echo "================================"

# Array of example directories
examples=(
    "basic_usage"
    "comment_alignment"
    "docker_compose"
    "kubernetes"
    "ansible"
    "wildcard_patterns"
    "file_operations"
)

# Counter for tracking results
success_count=0
total_count=${#examples[@]}

# Run each example
for example in "${examples[@]}"; do
    echo
    echo "🚀 Running example: $example"
    echo "----------------------------------------"
    
    if [ -d "$example" ] && [ -f "$example/main.go" ]; then
        cd "$example"
        
        # Run the example and capture output
        if go run main.go 2>&1; then
            echo "✅ $example completed successfully"
            ((success_count++))
        else
            echo "❌ $example failed"
        fi
        
        cd ..
    else
        echo "❌ $example directory or main.go not found"
    fi
    
    echo "----------------------------------------"
done

echo
echo "📊 Summary:"
echo "================================"
echo "Total examples: $total_count"
echo "Successful: $success_count"
echo "Failed: $((total_count - success_count))"

if [ $success_count -eq $total_count ]; then
    echo "🎉 All examples ran successfully!"
    exit 0
else
    echo "⚠️  Some examples failed"
    exit 1
fi 