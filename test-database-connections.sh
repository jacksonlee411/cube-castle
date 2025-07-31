#\!/bin/bash

# Database Connection Test Script for Cube Castle Project
# Tests PostgreSQL and Neo4j connections with various configurations

echo "🔗 Cube Castle Database Connection Test"
echo "======================================"

# Test 1: Check PostgreSQL Connection
echo "1. Testing PostgreSQL Connection..."

# Try default PostgreSQL connection
DB_CONFIGS=(
    "postgresql://postgres:postgres@localhost:5432/postgres"
    "postgresql://user:password@localhost:5432/cubecastle"  
    "postgresql://postgres@localhost:5432/cubecastle"
    "postgresql://cubecastle:cubecastle@localhost:5432/cubecastle"
)

POSTGRES_SUCCESS=false

for config in "${DB_CONFIGS[@]}"; do
    echo "   Trying: $config"
    if psql "$config" -c "SELECT version();" >/dev/null 2>&1; then
        echo "   ✅ Connection successful: $config"
        echo "   Testing database schema..."
        
        # Test if cubecastle database exists, create if not
        if [[ "$config" == *"/cubecastle"* ]]; then
            DB_NAME="cubecastle"
        else
            DB_NAME="postgres"
        fi
        
        # Check if our tables exist
        TABLE_COUNT=$(psql "$config" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema NOT IN ('information_schema', 'pg_catalog');" 2>/dev/null | tr -d ' ')
        echo "   Database has $TABLE_COUNT user tables"
        
        POSTGRES_SUCCESS=true
        WORKING_DB_URL="$config"
        break
    else
        echo "   ❌ Connection failed"
    fi
done

if [ "$POSTGRES_SUCCESS" = false ]; then
    echo "   🚨 No PostgreSQL connections succeeded. Checking service..."
    pg_isready -h localhost -p 5432
    
    echo "   Attempting to create database and user..."
    # Try to connect as postgres user and setup
    if psql "postgresql://postgres@localhost:5432/postgres" -c "CREATE DATABASE cubecastle;" 2>/dev/null; then
        echo "   ✅ Created cubecastle database"
    fi
    
    if psql "postgresql://postgres@localhost:5432/postgres" -c "CREATE USER cubecastle WITH PASSWORD 'cubecastle';" 2>/dev/null; then
        echo "   ✅ Created cubecastle user"
    fi
    
    if psql "postgresql://postgres@localhost:5432/postgres" -c "GRANT ALL PRIVILEGES ON DATABASE cubecastle TO cubecastle;" 2>/dev/null; then
        echo "   ✅ Granted privileges"
    fi
    
    # Retry connection
    TEST_CONFIG="postgresql://cubecastle:cubecastle@localhost:5432/cubecastle"
    if psql "$TEST_CONFIG" -c "SELECT version();" >/dev/null 2>&1; then
        echo "   ✅ Database setup successful: $TEST_CONFIG"
        POSTGRES_SUCCESS=true
        WORKING_DB_URL="$TEST_CONFIG"
    fi
fi

echo ""

# Generate Summary Report
echo "📋 Summary Report"
echo "================="

if [ "$POSTGRES_SUCCESS" = true ]; then
    echo "✅ PostgreSQL: Connected ($WORKING_DB_URL)"
    echo "   Recommended DATABASE_URL: $WORKING_DB_URL"
else
    echo "❌ PostgreSQL: Not available"
    echo "   Please install PostgreSQL or configure connection"
fi

echo ""
echo "🚀 Next Steps:"
if [ "$POSTGRES_SUCCESS" = true ]; then
    echo "1. Set environment variable: export DATABASE_URL='$WORKING_DB_URL'"
    echo "2. Restart your Go application to use database mode"
    echo "3. Run database migrations if needed"
else
    echo "1. Install and configure PostgreSQL"
    echo "2. Create cubecastle database and user"
    echo "3. Set DATABASE_URL environment variable"
fi

echo ""
echo "📝 For immediate testing, run:"
echo "   export DATABASE_URL='${WORKING_DB_URL:-postgresql://cubecastle:cubecastle@localhost:5432/cubecastle}'"
echo "   go run cmd/server/main.go"
