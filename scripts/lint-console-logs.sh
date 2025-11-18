#!/bin/bash
# Script to find console.log statements in JavaScript and Templ files

echo "🔍 Checking for console statements in JS and Templ files..."
echo ""

# Color codes
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Counters
total_count=0
file_count=0

# Search for console statements in JavaScript files (excluding minified files)
# Using awk to handle multi-line console statements with balanced parentheses
while IFS= read -r file; do
    # Use awk to find console statements and handle multi-line (tracks parentheses)
    file_results=$(awk '
        BEGIN { in_console = 0; start_line = 0; content = ""; paren_count = 0; }
        
        /console\.(log|warn|error|debug|info)/ {
            if (!in_console) {
                in_console = 1
                start_line = NR
                content = $0
                # Count parentheses on this line
                paren_count = gsub(/\(/, "(", $0) - gsub(/\)/, ")", $0)
            }
            next
        }
        
        in_console {
            content = content " " $0
            paren_count += gsub(/\(/, "(", $0) - gsub(/\)/, ")", $0)
            
            # Check if statement is complete (balanced parens and ends with semicolon)
            if (paren_count <= 0 && ($0 ~ /\);$/ || $0 ~ /;$/)) {
                # Print the complete statement
                gsub(/^[[:space:]]+|[[:space:]]+$/, "", content)
                # Replace multiple spaces with single space
                gsub(/[[:space:]]+/, " ", content)
                # Truncate long lines
                if (length(content) > 200) {
                    content = substr(content, 1, 200) "..."
                }
                print start_line ":" content
                in_console = 0
                content = ""
                paren_count = 0
            }
        }
        
        END {
            # Handle unclosed statements
            if (in_console && content != "") {
                gsub(/^[[:space:]]+|[[:space:]]+$/, "", content)
                gsub(/[[:space:]]+/, " ", content)
                if (length(content) > 200) {
                    content = substr(content, 1, 200) "..."
                }
                print start_line ":" content
            }
        }
    ' "$file" 2>/dev/null || true)
    
    if [ -n "$file_results" ]; then
        count=$(echo "$file_results" | wc -l)
        total_count=$((total_count + count))
        file_count=$((file_count + 1))
        
        echo -e "${YELLOW}$file${NC} - ${RED}$count${NC} statement(s)"
        
        # Show line numbers
        echo "$file_results" | while IFS=: read -r line_num content; do
            # Trim whitespace
            content=$(echo "$content" | sed 's/^[[:space:]]*//')
            echo -e "  ${CYAN}Line $line_num:${NC} ${content:0:100}$([ ${#content} -gt 100 ] && echo '...')"
        done
        echo ""
    fi
done < <(find static/js -type f -name "*.js" ! -name "*.min.js" ! -path "*/vendor/*" 2>/dev/null)

# Also check templ files
while IFS= read -r file; do
    # Use awk to find console statements in templ files
    file_results=$(awk '
        BEGIN { in_console = 0; start_line = 0; content = ""; paren_count = 0; }
        
        /console\.(log|warn|error|debug|info)/ {
            if (!in_console) {
                in_console = 1
                start_line = NR
                content = $0
                # Count parentheses on this line
                paren_count = gsub(/\(/, "(", $0) - gsub(/\)/, ")", $0)
            }
            next
        }
        
        in_console {
            content = content " " $0
            paren_count += gsub(/\(/, "(", $0) - gsub(/\)/, ")", $0)
            
            # Check if statement is complete (balanced parens and ends with semicolon or quote)
            if (paren_count <= 0 && ($0 ~ /\);$/ || $0 ~ /;$/ || $0 ~ /"$/)) {
                # Print the complete statement
                gsub(/^[[:space:]]+|[[:space:]]+$/, "", content)
                gsub(/[[:space:]]+/, " ", content)
                if (length(content) > 200) {
                    content = substr(content, 1, 200) "..."
                }
                print start_line ":" content
                in_console = 0
                content = ""
                paren_count = 0
            }
        }
        
        END {
            if (in_console && content != "") {
                gsub(/^[[:space:]]+|[[:space:]]+$/, "", content)
                gsub(/[[:space:]]+/, " ", content)
                if (length(content) > 200) {
                    content = substr(content, 1, 200) "..."
                }
                print start_line ":" content
            }
        }
    ' "$file" 2>/dev/null || true)
    
    if [ -n "$file_results" ]; then
        count=$(echo "$file_results" | wc -l)
        total_count=$((total_count + count))
        file_count=$((file_count + 1))
        
        echo -e "${YELLOW}$file${NC} - ${RED}$count${NC} statement(s)"
        
        # Show line numbers
        echo "$file_results" | while IFS=: read -r line_num content; do
            content=$(echo "$content" | sed 's/^[[:space:]]*//')
            echo -e "  ${CYAN}Line $line_num:${NC} ${content:0:100}$([ ${#content} -gt 100 ] && echo '...')"
        done
        echo ""
    fi
done < <(find internal/web -type f -name "*.templ" 2>/dev/null)

echo ""

# Summary
if [ $total_count -eq 0 ]; then
    echo -e "${GREEN}✅ No console statements found!${NC}"
    exit 0
else
    echo -e "${BLUE}═══════════════════════════════════════${NC}"
    echo -e "${RED}⚠️  Summary: Found $total_count console statement(s) in $file_count file(s)${NC}"
    echo -e "${BLUE}═══════════════════════════════════════${NC}"
    echo ""
    echo "💡 Tip: Remove debug console statements before production."
    echo "   JS files: grep -rn 'console\.' static/js --include='*.js' --exclude='*.min.js'"
    echo "   Templ files: grep -rn 'console\.' internal/web --include='*.templ'"
    exit 1
fi
