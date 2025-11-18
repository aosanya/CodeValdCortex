#!/bin/bash
# Script to remove all .Debug( logging statements from Go files

set -e

echo "Removing all .Debug() logging statements from Go files..."

# Find all Go files with .Debug( calls
files_with_debug=$(grep -rl "\.Debug(" internal/ --include="*.go" || true)

if [ -z "$files_with_debug" ]; then
    echo "No .Debug() calls found."
    exit 0
fi

echo "Files with .Debug() calls:"
echo "$files_with_debug"
echo ""

# Process each file
for file in $files_with_debug; do
    echo "Processing: $file"
    
    # Create a temporary file
    temp_file="${file}.tmp"
    
    # Use awk to remove Debug logging patterns
    # This handles multi-line Debug calls with WithFields
    awk '
    BEGIN { in_debug = 0; skip_lines = 0 }
    {
        # Check if this line starts a Debug call
        if ($0 ~ /\.Debug\(/ && $0 !~ /\/\/.*\.Debug\(/) {
            # Single line Debug call - skip it
            if ($0 ~ /\)[ \t]*$/ || $0 ~ /\)[,;][ \t]*$/) {
                next
            } else {
                # Multi-line - mark for skipping
                in_debug = 1
                next
            }
        }
        
        # Check if previous line had WithField(s) leading to Debug
        if ($0 ~ /^[ \t]*\}\)\.Debug\(/) {
            # This is the .Debug( part after WithFields
            if ($0 ~ /\)[ \t]*$/ || $0 ~ /\)[,;][ \t]*$/) {
                # Single line completion
                skip_lines = 0
                next
            } else {
                # Multi-line Debug call
                in_debug = 1
                skip_lines = 0
                next
            }
        }
        
        # If we are in a Debug call, skip until we find the closing paren
        if (in_debug == 1) {
            if ($0 ~ /\)[ \t]*$/ || $0 ~ /\)[,;][ \t]*$/) {
                in_debug = 0
            }
            next
        }
        
        # Check for WithFields/WithField followed by Debug on next lines
        if ($0 ~ /\.WithFields?\(/ && $0 !~ /\/\/.*\.WithFields?\(/) {
            # Store the line
            buffer[++buf_count] = $0
            
            # Check if this starts a multi-line WithFields
            if ($0 !~ /\}\)\.Debug\(/ && $0 ~ /\{[ \t]*$/) {
                skip_lines = 1
                next
            }
        }
        
        # If we are collecting lines for potential WithFields -> Debug pattern
        if (skip_lines > 0) {
            buffer[++buf_count] = $0
            
            # Check if this completes WithFields and has Debug
            if ($0 ~ /\}\)\.Debug\(/) {
                # This is a WithFields...Debug pattern - clear buffer and skip
                buf_count = 0
                skip_lines = 0
                
                # If Debug completes on this line, done
                if ($0 ~ /\)[ \t]*$/ || $0 ~ /\)[,;][ \t]*$/) {
                    next
                } else {
                    # Multi-line Debug - continue skipping
                    in_debug = 1
                    next
                }
            }
            
            # Check if WithFields completes but no Debug
            if ($0 ~ /\}\)\./ && $0 !~ /\.Debug\(/) {
                # Not a Debug call - print buffered lines
                for (i = 1; i <= buf_count; i++) {
                    print buffer[i]
                }
                buf_count = 0
                skip_lines = 0
                next
            }
            
            # Just a field definition - continue collecting
            next
        }
        
        # Check if we have buffered lines and current line is not part of WithFields
        if (buf_count > 0 && $0 !~ /^[ \t]*".*":/) {
            # Not part of WithFields - check if it is Debug
            if ($0 ~ /^[ \t]*\}\)\.Debug\(/) {
                # Yes, it is Debug - clear buffer and skip
                buf_count = 0
                if ($0 ~ /\)[ \t]*$/ || $0 ~ /\)[,;][ \t]*$/) {
                    next
                } else {
                    in_debug = 1
                    next
                }
            } else {
                # Not Debug - print buffered lines and current line
                for (i = 1; i <= buf_count; i++) {
                    print buffer[i]
                }
                buf_count = 0
                print $0
                next
            }
        }
        
        # Normal line - print it
        if (buf_count == 0) {
            print $0
        }
    }
    END {
        # Print any remaining buffered lines
        for (i = 1; i <= buf_count; i++) {
            print buffer[i]
        }
    }
    ' "$file" > "$temp_file"
    
    # Replace original file
    mv "$temp_file" "$file"
    
    echo "  ✓ Processed $file"
done

echo ""
echo "✅ All .Debug() logging statements removed."
echo ""
echo "Files modified: $(echo "$files_with_debug" | wc -l)"
