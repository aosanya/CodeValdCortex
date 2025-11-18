#!/bin/bash
# Check for common violations of .github/instructions/rules.instructions.md

set -e

ERRORS=0

echo "🔍 Checking code structure rules..."

# Rule: No HTML strings in Go handlers
echo "  ❌ Checking for HTML strings in Go handlers..."
if grep -rn 'fmt.Sprintf.*<div\|<span\|<p\|<button' internal/web/handlers/ --include="*.go" 2>/dev/null; then
    echo "  ❌ VIOLATION: HTML generation in Go handlers (use .templ files instead)"
    ERRORS=$((ERRORS + 1))
fi

# Rule: Check file sizes (max 700 lines)
echo "  📏 Checking file sizes..."
find internal/web/handlers -name "*.go" -type f | while read -r file; do
    lines=$(wc -l < "$file")
    if [ "$lines" -gt 700 ]; then
        echo "  ❌ VIOLATION: $file is $lines lines (max 700)"
        ERRORS=$((ERRORS + 1))
    fi
done

# Rule: Check function sizes (max 50 lines - approximate)
echo "  🔧 Checking for large functions..."
if command -v gocyclo &> /dev/null; then
    if gocyclo -over 15 internal/ 2>/dev/null | grep -v "^$"; then
        echo "  ⚠️  WARNING: Complex functions found (max cyclomatic complexity: 15)"
    fi
fi

# Rule: Check for duplicate types
echo "  🔄 Checking for duplicate type definitions..."
if grep -rn "type WorkflowStatus" internal/ --include="*.go" | wc -l | grep -v "^1$" &>/dev/null; then
    echo "  ⚠️  WARNING: Possible duplicate WorkflowStatus type definitions"
fi

# Rule: Check for inline JavaScript in .templ files
echo "  🚫 Checking for inline JavaScript in .templ files..."
if grep -rn '<script>' internal/web/pages/ internal/web/templates/ --include="*.templ" 2>/dev/null | grep -v "src="; then
    echo "  ❌ VIOLATION: Inline JavaScript in .templ files (use .js files instead)"
    ERRORS=$((ERRORS + 1))
fi

# Rule: Check for HTML generation in JavaScript
echo "  🌐 Checking for HTML generation in JavaScript..."
if grep -rn 'innerHTML.*<div\|innerHTML.*<span' static/js/ --include="*.js" 2>/dev/null | grep -v "streaming"; then
    echo "  ⚠️  WARNING: Possible HTML generation in JavaScript (use .templ or server rendering)"
fi

if [ $ERRORS -eq 0 ]; then
    echo "✅ All rule checks passed!"
    exit 0
else
    echo "❌ $ERRORS rule violations found"
    exit 1
fi
