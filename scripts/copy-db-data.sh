#!/bin/bash

# Simple script to copy all data from one database to another
# Usage: ./copy-db-data.sh <source_db> <target_db>

set -e

SOURCE_DB="${1:-water_distribution_network}"
TARGET_DB="${2:-UC-INFRA-001}"
ARANGO_HOST="${ARANGO_HOST:-host.docker.internal}"
ARANGO_PORT="${ARANGO_PORT:-8529}"
ARANGO_USER="${ARANGO_USER:-root}"
ARANGO_PASSWORD="${ARANGO_PASSWORD:-rootpassword}"

echo "Copying data from $SOURCE_DB to $TARGET_DB..."

# Get all collections from source
COLLECTIONS=$(curl -s -u "$ARANGO_USER:$ARANGO_PASSWORD" \
    "http://$ARANGO_HOST:$ARANGO_PORT/_db/$SOURCE_DB/_api/collection" | \
    jq -r '.result[] | select(.isSystem == false) | .name')

for COLLECTION in $COLLECTIONS; do
    echo "Processing collection: $COLLECTION"
    
    # Get all documents from source
    DOCS=$(curl -s -u "$ARANGO_USER:$ARANGO_PASSWORD" \
        -X POST "http://$ARANGO_HOST:$ARANGO_PORT/_db/$SOURCE_DB/_api/cursor" \
        -H "Content-Type: application/json" \
        -d "{\"query\":\"FOR doc IN $COLLECTION RETURN doc\"}" | \
        jq -c '.result[]')
    
    if [ -z "$DOCS" ]; then
        echo "  No documents found"
        continue
    fi
    
    # Create collection in target if it doesn't exist
    curl -s -u "$ARANGO_USER:$ARANGO_PASSWORD" \
        -X POST "http://$ARANGO_HOST:$ARANGO_PORT/_db/$TARGET_DB/_api/collection" \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"$COLLECTION\"}" > /dev/null 2>&1 || true
    
    # Insert documents one by one
    COUNT=0
    while IFS= read -r DOC; do
        if [ ! -z "$DOC" ]; then
            curl -s -u "$ARANGO_USER:$ARANGO_PASSWORD" \
                -X POST "http://$ARANGO_HOST:$ARANGO_PORT/_db/$TARGET_DB/_api/document/$COLLECTION" \
                -H "Content-Type: application/json" \
                -d "$DOC" > /dev/null
            COUNT=$((COUNT + 1))
        fi
    done <<< "$DOCS"
    
    echo "  Copied $COUNT documents"
done

echo "Done!"
