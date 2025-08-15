#!/bin/bash

# Database utility functions for connecting to PostgreSQL

# Load environment variables from .env file
load_env() {
    local env_file=""
    if [ -f ".env" ]; then
        env_file=".env"
    elif [ -f "../../.env" ]; then
        env_file="../../.env"
    else
        echo "❌ Error: .env file not found"
        exit 1
    fi
    
    # Export variables while handling values with spaces
    while IFS='=' read -r key value; do
        # Skip comments and empty lines
        [[ $key =~ ^[[:space:]]*# ]] && continue
        [[ -z "$key" ]] && continue
        
        # Remove leading/trailing whitespace from key
        key=$(echo "$key" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
        
        # Export the variable (value may contain spaces)
        export "$key"="$value"
    done < "$env_file"
}

# Get PostgreSQL connection string
get_pg_connection() {
    load_env
    echo "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
}

# Get psql connection parameters
get_psql_params() {
    load_env
    echo "-h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME}"
}

# Test database connection
test_db_connection() {
    load_env
    echo "🔍 Testing database connection..."
    
    PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -c "SELECT 1;" > /dev/null 2>&1
    
    if [ "$?" -eq 0 ]; then
        echo "✅ Database connection successful"
        return 0
    else
        echo "❌ Database connection failed"
        return 1
    fi
}

# Execute SQL command
execute_sql() {
    local sql_command="$1"
    load_env
    
    PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -c "${sql_command}"
}

# Execute SQL file
execute_sql_file() {
    local sql_file="$1"
    load_env
    
    if [ ! -f "$sql_file" ]; then
        echo "❌ Error: SQL file '$sql_file' not found"
        return 1
    fi
    
    PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -f "${sql_file}"
}