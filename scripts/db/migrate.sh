#!/bin/bash

# Database migration script using goose
# Usage: ./migrate.sh [command] [args...]

set -e  # Exit on any error

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
MIGRATIONS_DIR="${PROJECT_ROOT}/migrations"

# Source database utilities
source "${SCRIPT_DIR}/db_utils.sh"

# Change to project root to find .env file
cd "${PROJECT_ROOT}"

# Function to get database URL for goose
get_database_url() {
    load_env
    echo "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"
}

# Function to run goose command
run_goose() {
    local command="$1"
    shift  # Remove first argument
    local db_url
    db_url=$(get_database_url)
    
    echo "🦆 Running goose $command..."
    goose -dir "$MIGRATIONS_DIR" postgres "$db_url" "$command" "$@"
}

# Function to show migration status
show_status() {
    echo "📊 Migration Status:"
    echo "Database: ${DB_NAME} on ${DB_HOST}:${DB_PORT}"
    echo "Migrations directory: $MIGRATIONS_DIR"
    echo ""
    run_goose status
}

# Function to apply migrations
migrate_up() {
    local count="${1:-all}"
    echo "⬆️  Applying migrations..."
    
    if [ "$count" = "all" ]; then
        run_goose up
    else
        run_goose up-by-one
    fi
    
    echo "✅ Migrations applied successfully"
    show_status
}

# Function to rollback migrations
migrate_down() {
    local count="${1:-1}"
    echo "⚠️  WARNING: This will rollback database changes!"
    echo "Database: ${DB_NAME} on ${DB_HOST}:${DB_PORT}"
    read -p "Are you sure you want to rollback $count migration(s)? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Rollback cancelled"
        exit 1
    fi
    
    echo "⬇️  Rolling back migrations..."
    if [ "$count" = "1" ]; then
        run_goose down
    else
        for ((i=1; i<=count; i++)); do
            run_goose down
        done
    fi
    
    echo "✅ Rollback completed"
    show_status
}

# Function to create new migration
create_migration() {
    local name="$1"
    local type="${2:-sql}"
    
    if [ -z "$name" ]; then
        echo "❌ Error: Migration name is required"
        echo "Usage: $0 create <migration_name> [sql|go]"
        exit 1
    fi
    
    echo "📝 Creating new migration: $name"
    run_goose create "$name" "$type"
    
    echo "✅ Migration created successfully"
    echo "📁 Check the $MIGRATIONS_DIR directory for the new migration file"
}

# Function to reset database (down all + up all)
reset_migrations() {
    echo "🔄 WARNING: This will reset all migrations (down all + up all)!"
    echo "Database: ${DB_NAME} on ${DB_HOST}:${DB_PORT}"
    read -p "Are you sure you want to reset all migrations? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Reset cancelled"
        exit 1
    fi
    
    echo "⬇️  Rolling back all migrations..."
    run_goose reset
    
    echo "⬆️  Applying all migrations..."
    run_goose up
    
    echo "✅ Database reset completed"
    show_status
}

# Function to validate migrations
validate_migrations() {
    echo "🔍 Validating migrations..."
    run_goose validate
    echo "✅ Migrations are valid"
}

# Function to show usage
show_usage() {
    echo "Database Migration Tool (using goose)"
    echo ""
    echo "Usage: $0 <command> [args...]"
    echo ""
    echo "Commands:"
    echo "  up [N]           Apply pending migrations (or N migrations)"
    echo "  down [N]         Rollback N migrations (default: 1)"
    echo "  status           Show migration status"
    echo "  create <name>    Create new migration file"
    echo "  reset            Reset all migrations (down all + up all)"
    echo "  validate         Validate migration files"
    echo "  version          Show current migration version"
    echo ""
    echo "Examples:"
    echo "  $0 status                    # Show current status"
    echo "  $0 up                        # Apply all pending migrations"
    echo "  $0 down                      # Rollback last migration"
    echo "  $0 down 3                    # Rollback last 3 migrations"
    echo "  $0 create add_user_table     # Create new migration"
    echo "  $0 reset                     # Reset all migrations"
}

# Main logic
main() {
    # Test connection first (except for create command)
    if [ "$1" != "create" ] && [ "$1" != "--help" ] && [ "$1" != "-h" ]; then
        if ! test_db_connection; then
            echo "❌ Cannot proceed due to database connection failure"
            exit 1
        fi
    fi
    
    # Handle commands
    case "${1:-}" in
        "up"|"apply")
            migrate_up "${2:-all}"
            ;;
        "down"|"rollback")
            migrate_down "${2:-1}"
            ;;
        "status"|"st")
            show_status
            ;;
        "create"|"new")
            create_migration "$2" "$3"
            ;;
        "reset")
            reset_migrations
            ;;
        "validate"|"check")
            validate_migrations
            ;;
        "version"|"ver")
            run_goose version
            ;;
        "--help"|"-h"|"help")
            show_usage
            ;;
        "")
            echo "❌ Error: Command is required"
            echo ""
            show_usage
            exit 1
            ;;
        *)
            echo "❌ Error: Unknown command '$1'"
            echo ""
            show_usage
            exit 1
            ;;
    esac
}

# Run main function
main "$@"