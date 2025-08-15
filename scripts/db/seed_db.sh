#!/bin/bash

# Database seeding script - loads default data into tables
# Usage: ./seed_db.sh [table_name] or ./seed_db.sh all

set -e  # Exit on any error

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
SEED_DATA_DIR="${SCRIPT_DIR}/seed_data"

# Source database utilities
source "${SCRIPT_DIR}/db_utils.sh"

# Change to project root to find .env file
cd "${PROJECT_ROOT}"

# Available tables for seeding
AVAILABLE_TABLES=("currencies" "account_types" "languages" "default_categories")

# Function to seed a specific table
seed_table() {
    local table_name="$1"
    local sql_file="${SEED_DATA_DIR}/${table_name}.sql"
    
    if [ ! -f "$sql_file" ]; then
        echo "❌ Error: SQL file for table '$table_name' not found: $sql_file"
        return 1
    fi
    
    echo "🌱 Seeding table: $table_name"
    if execute_sql_file "$sql_file"; then
        echo "✅ Successfully seeded table: $table_name"
    else
        echo "❌ Failed to seed table: $table_name"
        return 1
    fi
}

# Function to seed all tables
seed_all_tables() {
    echo "🌱 Seeding all tables..."
    
    for table in "${AVAILABLE_TABLES[@]}"; do
        seed_table "$table"
    done
    
    echo "✅ All tables seeded successfully!"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [table_name|all]"
    echo ""
    echo "Available tables:"
    for table in "${AVAILABLE_TABLES[@]}"; do
        echo "  - $table"
    done
    echo "  - all (seeds all tables)"
    echo ""
    echo "Examples:"
    echo "  $0 currencies          # Seed only currencies table"
    echo "  $0 all                 # Seed all tables"
    echo "  $0                     # Interactive mode"
}

# Main logic
main() {
    # Test connection first
    if ! test_db_connection; then
        echo "❌ Cannot proceed with seeding due to database connection failure"
        exit 1
    fi
    
    # Handle command line arguments
    if [ $# -eq 0 ]; then
        # Interactive mode
        echo "🌱 Database Seeding Tool"
        echo ""
        echo "Available options:"
        echo "1. Seed all tables"
        for i in "${!AVAILABLE_TABLES[@]}"; do
            echo "$((i+2)). Seed ${AVAILABLE_TABLES[i]} table"
        done
        echo ""
        read -p "Select an option (1-$((${#AVAILABLE_TABLES[@]}+1))): " choice
        
        case $choice in
            1)
                seed_all_tables
                ;;
            *)
                table_index=$((choice-2))
                if [ $table_index -ge 0 ] && [ $table_index -lt ${#AVAILABLE_TABLES[@]} ]; then
                    seed_table "${AVAILABLE_TABLES[table_index]}"
                else
                    echo "❌ Invalid choice"
                    exit 1
                fi
                ;;
        esac
    elif [ "$1" = "all" ]; then
        seed_all_tables
    elif [[ " ${AVAILABLE_TABLES[*]} " =~ " $1 " ]]; then
        seed_table "$1"
    elif [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
        show_usage
    else
        echo "❌ Error: Unknown table '$1'"
        echo ""
        show_usage
        exit 1
    fi
}

# Run main function
main "$@"