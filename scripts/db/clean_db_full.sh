#!/bin/bash

# Full database clean script
# This script drops the public schema and recreates it, effectively wiping all data

set -e  # Exit on any error

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Source database utilities
source "${SCRIPT_DIR}/db_utils.sh"

# Change to project root to find .env file
cd "${PROJECT_ROOT}"

echo "🧹 Cleaning DB..."

# Test connection first
if ! test_db_connection; then
    echo "❌ Cannot proceed with database cleaning due to connection failure"
    exit 1
fi

# Confirm action
echo "⚠️  WARNING: This will completely wipe all data in the database!"
echo "Database: ${DB_NAME} on ${DB_HOST}:${DB_PORT}"
read -p "Are you sure you want to continue? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Operation cancelled"
    exit 1
fi

# Execute the cleaning commands
echo "🗑️  Dropping public schema..."
execute_sql "DROP SCHEMA public CASCADE;"

echo "🔨 Recreating public schema..."
execute_sql "CREATE SCHEMA public;"

echo "✅ DB fully wiped and public schema recreated"
echo "💡 You may want to run migrations or seed scripts next"