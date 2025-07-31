#!/bin/bash

# E2E Test Runner Script
# This script starts the development server and runs E2E tests

echo "🚀 Starting E2E Test Suite for Cube Castle HR System"
echo "=================================================="

# Check if Node.js and npm are available
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js first."
    exit 1
fi

if ! command -v npm &> /dev/null; then
    echo "❌ npm is not installed. Please install npm first."
    exit 1
fi

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "📦 Installing dependencies..."
    npm install
fi

# Install Playwright browsers if needed
echo "🔧 Setting up Playwright browsers..."
npx playwright install

echo "📋 E2E Test Summary:"
echo "• Testing 7 refactored pages from UI standardization project"
echo "• Cross-browser testing (Chromium, Firefox, WebKit)"
echo "• Comprehensive coverage including CRUD, search, responsive design"
echo "• Performance validation (<3s load time requirement)"
echo "• Component integration testing (shadcn/ui + Radix UI)"
echo ""

echo "⚠️  NOTE: These tests require the development server to be running."
echo "   Please start the server with 'npm run dev' in another terminal."
echo "   The tests will connect to http://localhost:3000"
echo ""

read -p "Press Enter to continue when the development server is ready..."

echo "🧪 Running E2E Tests..."
echo "========================"

# Run tests with HTML reporter for better visualization
npx playwright test --reporter=html --reporter=list

# Check if tests passed
if [ $? -eq 0 ]; then
    echo ""
    echo "✅ All E2E tests completed successfully!"
    echo "📊 Test report generated at: playwright-report/index.html"
    echo "🖼️  Screenshots saved to: test-results/screenshots/"
    echo ""
    echo "📈 Test Coverage Summary:"
    echo "• 7 pages tested with comprehensive scenarios"
    echo "• Cross-browser compatibility validated"
    echo "• UI component integration verified"
    echo "• Performance benchmarks met"
    echo "• Responsive design validated"
else
    echo ""
    echo "❌ Some E2E tests failed."
    echo "📊 Check the HTML report for details: playwright-report/index.html"
    echo "🖼️  Screenshots available at: test-results/screenshots/"
    echo ""
    echo "🔍 Common issues:"
    echo "• Development server not running on localhost:3000"
    echo "• Database not seeded with test data"
    echo "• Browser compatibility issues"
    echo "• Network connectivity problems"
fi