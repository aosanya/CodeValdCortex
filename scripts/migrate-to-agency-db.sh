#!/bin/bash

# Script to migrate data from the main database to agency-specific databases
# Usage: ./migrate-to-agency-db.sh <source_db> <agency_id>

set -e

# Configuration
ARANGO_HOST="${ARANGO_HOST:-host.docker.internal}"
ARANGO_PORT="${ARANGO_PORT:-8529}"
ARANGO_USER="${ARANGO_USER:-root}"
ARANGO_PASSWORD="${ARANGO_PASSWORD:-}"

SOURCE_DB="${1:-codevaldcortex}"
AGENCY_ID="${2:-UC-INFRA-001}"
TARGET_DB="${AGENCY_ID}"

echo "=================================="
echo "Agency Database Migration Script"
echo "=================================="
echo "Source Database: $SOURCE_DB"
echo "Target Database: $TARGET_DB"
echo "ArangoDB Host: $ARANGO_HOST:$ARANGO_PORT"
echo "=================================="
echo ""

# Function to execute ArangoDB commands
arango_exec() {
    local db="$1"
    local query="$2"
    
    curl -s -X POST \
        -u "${ARANGO_USER}:${ARANGO_PASSWORD}" \
        --data "{\"query\": \"${query}\"}" \
        "http://${ARANGO_HOST}:${ARANGO_PORT}/_db/${db}/_api/cursor" \
        -H "Content-Type: application/json"
}

# Check if source database exists
echo "Checking if source database exists..."
SOURCE_EXISTS=$(curl -s -u "${ARANGO_USER}:${ARANGO_PASSWORD}" \
    "http://${ARANGO_HOST}:${ARANGO_PORT}/_api/database" | \
    grep -o "\"${SOURCE_DB}\"" || echo "")

if [ -z "$SOURCE_EXISTS" ]; then
    echo "ERROR: Source database '${SOURCE_DB}' does not exist!"
    exit 1
fi
echo "✓ Source database exists"

# Check if target database already exists
echo "Checking if target database exists..."
TARGET_EXISTS=$(curl -s -u "${ARANGO_USER}:${ARANGO_PASSWORD}" \
    "http://${ARANGO_HOST}:${ARANGO_PORT}/_api/database" | \
    grep -o "\"${TARGET_DB}\"" || echo "")

if [ -n "$TARGET_EXISTS" ]; then
    echo "WARNING: Target database '${TARGET_DB}' already exists!"
    read -p "Do you want to drop and recreate it? (yes/no): " confirm
    if [ "$confirm" = "yes" ]; then
        echo "Dropping existing target database..."
        curl -s -X DELETE \
            -u "${ARANGO_USER}:${ARANGO_PASSWORD}" \
            "http://${ARANGO_HOST}:${ARANGO_PORT}/_api/database/${TARGET_DB}"
        echo "✓ Dropped existing database"
    else
        echo "Aborted."
        exit 1
    fi
fi

# Create target database
echo "Creating target database: ${TARGET_DB}..."
CREATE_RESULT=$(curl -s -X POST \
    -u "${ARANGO_USER}:${ARANGO_PASSWORD}" \
    --data "{\"name\": \"${TARGET_DB}\"}" \
    "http://${ARANGO_HOST}:${ARANGO_PORT}/_api/database" \
    -H "Content-Type: application/json")

if echo "$CREATE_RESULT" | grep -q '"error":false'; then
    echo "✓ Created target database"
else
    echo "ERROR creating database: $CREATE_RESULT"
    exit 1
fi

# Get list of collections from source database
echo "Getting collections from source database..."
COLLECTIONS=$(curl -s -u "${ARANGO_USER}:${ARANGO_PASSWORD}" \
    "http://${ARANGO_HOST}:${ARANGO_PORT}/_db/${SOURCE_DB}/_api/collection" | \
    grep -o '"name":"[^"]*"' | \
    grep -v '"name":"_' | \
    cut -d'"' -f4)

echo "Collections found: $COLLECTIONS"

# Copy each collection
for COLLECTION in $COLLECTIONS; do
    echo ""
    echo "Processing collection: $COLLECTION"
    
    # Create collection in target database
    echo "  Creating collection in target database..."
    curl -s -X POST \
        -u "${ARANGO_USER}:${ARANGO_PASSWORD}" \
        --data "{\"name\": \"${COLLECTION}\"}" \
        "http://${ARANGO_HOST}:${ARANGO_PORT}/_db/${TARGET_DB}/_api/collection" \
        -H "Content-Type: application/json" > /dev/null
    
    # Export documents from source
    echo "  Exporting documents..."
    DOCS=$(curl -s -X POST \
        -u "${ARANGO_USER}:${ARANGO_PASSWORD}" \
        --data "{\"query\": \"FOR doc IN ${COLLECTION} RETURN doc\"}" \
        "http://${ARANGO_HOST}:${ARANGO_PORT}/_db/${SOURCE_DB}/_api/cursor" \
        -H "Content-Type: application/json")
    
    # Count documents
    DOC_COUNT=$(echo "$DOCS" | grep -o '"_key"' | wc -l)
    echo "  Found $DOC_COUNT documents"
    
    if [ "$DOC_COUNT" -gt 0 ]; then
        # Import documents to target
        echo "  Importing documents to target database..."
        echo "$DOCS" | jq -r '.result[]' | while read -r doc; do
            curl -s -X POST \
                -u "${ARANGO_USER}:${ARANGO_PASSWORD}" \
                --data "$doc" \
                "http://${ARANGO_HOST}:${ARANGO_PORT}/_db/${TARGET_DB}/_api/document/${COLLECTION}" \
                -H "Content-Type: application/json" > /dev/null
        done
        echo "  ✓ Imported $DOC_COUNT documents"
    else
        echo "  ✓ No documents to import"
    fi
done

echo ""
echo "=================================="
echo "Migration completed successfully!"
echo "=================================="
echo ""
echo "New database '${TARGET_DB}' created with all collections and documents."
echo ""
echo "To use this database, update your configuration:"
echo "  export DB_DATABASE=${TARGET_DB}"
echo "  or update config.yaml:"
echo "  database:"
echo "    database: ${TARGET_DB}"
echo ""
