#!/bin/bash

# Test Runner Script with Environment Variables
# This script loads .env.test before running tests

set -e  # Exit on error

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Check if .env.test exists
if [ ! -f ".env.test" ]; then
    echo "❌ Error: .env.test not found in $PROJECT_ROOT"
    echo "Please create .env.test with DATABASE_URL and other test environment variables"
    exit 1
fi

echo "📝 Loading test environment variables from .env.test..."

# Export environment variables from .env.test
set -a
while IFS='=' read -r line || [ -n "$line" ]; do
    # Skip comments and empty lines
    [[ "$line" =~ ^#.*$ ]] && continue
    [ -z "$line" ] && continue
    export "$line"
done < .env.test

echo "✅ Environment variables loaded"

# Check if Docker test database is running
if ! docker ps | grep -q postgres-test; then
    echo "🐳 Starting PostgreSQL test database..."
    docker-compose -f docker-compose.test.yml up -d

    # Wait for database to be ready
    echo "⏳ Waiting for database to be ready..."
    max_attempts=30
    attempt=0
    while [ $attempt -lt $max_attempts ]; do
        if docker exec nodepulse-test-db pg_isready -U testuser -d nodepulse_test > /dev/null 2>&1; then
            echo "✅ Database is ready!"
            break
        fi
        sleep 1
        attempt=$((attempt + 1))
    done

    if [ $attempt -eq $max_attempts ]; then
        echo "❌ Error: Database did not become ready within $max_attempts seconds"
        exit 1
    fi
else
    echo "✅ PostgreSQL test database is already running"
fi

# Run tests
echo ""
echo "🧪 Running tests..."
echo ""

# Check if test path is provided
if [ -n "$1" ]; then
    # Run specific test or package
    go test -v "$1"
else
    # Run all tests with coverage
    go test -v -coverprofile=coverage.out -covermode=atomic ./...
fi

echo ""
echo "✅ Tests completed"
echo "💡 To stop the test database: docker-compose -f docker-compose.test.yml down"
