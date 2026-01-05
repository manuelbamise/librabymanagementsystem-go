#!/bin/bash

# Library Management System - Quick Start Script

echo "🚀 Library Management System - Quick Start"
echo "=========================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Check if templ is installed
if ! command -v templ &> /dev/null; then
    echo "📦 Installing templ CLI..."
    go install github.com/a-h/templ/cmd/templ@latest
fi

echo "📦 Installing dependencies..."
go mod tidy

echo "🔨 Generating templates..."
templ generate

echo "🏗️  Building application..."
go build .

echo "🎉 Ready to run!"
echo ""
echo "To start the server:"
echo "  ./librarymanagementsystem"
echo ""
echo "Or run directly with:"
echo "  go run ."
echo ""
echo "Then open http://localhost:8080 in your browser"
echo "=========================================="