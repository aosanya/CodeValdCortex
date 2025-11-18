#!/usr/bin/env python3
"""
Script to remove all .Debug() logging statements from Go files.
"""

import os
import re
import sys
from pathlib import Path

def remove_debug_logs(content):
    """Remove Debug logging statements from Go code."""
    lines = content.split('\n')
    result = []
    i = 0
    
    while i < len(lines):
        line = lines[i]
        stripped = line.strip()
        
        # Check if this is a standalone .Debug( line
        if '.Debug(' in line and not stripped.startswith('//'):
            # Check if it is a single-line Debug call
            if line.rstrip().endswith(')'):
                # Skip this line entirely
                i += 1
                continue
            else:
                # Multi-line Debug call - skip until we find the closing paren
                i += 1
                while i < len(lines):
                    if lines[i].rstrip().endswith(')'):
                        i += 1
                        break
                    i += 1
                continue
        
        # Check if this line starts with .WithFields( or .WithField(
        if (('.WithFields(log.Fields{' in line or '.WithField(' in line) and 
            not stripped.startswith('//')):
            # Look ahead to see if this leads to .Debug(
            lookahead = []
            lookahead.append(line)
            j = i + 1
            has_debug = False
            
            # Collect lines until we find Debug or a complete statement
            while j < len(lines):
                lookahead.append(lines[j])
                if '.Debug(' in lines[j]:
                    has_debug = True
                    # Continue to find the end of Debug call
                    if lines[j].rstrip().endswith(')'):
                        # Single line end
                        i = j + 1
                        break
                    else:
                        # Multi-line - find end
                        j += 1
                        while j < len(lines):
                            if lines[j].rstrip().endswith(')'):
                                i = j + 1
                                break
                            j += 1
                        break
                # Check if we have completed the WithFields/WithField call
                if lines[j].strip().startswith('})') and '.Debug(' not in lines[j]:
                    # This is not leading to Debug - keep these lines
                    break
                if lines[j].strip() == '})':
                    j += 1
                    continue
                if lines[j].strip().startswith('"') and not '.Debug(' in lines[j]:
                    j += 1
                    continue
                # Some other statement - not Debug related
                if not (lines[j].strip().startswith('"') or 
                       lines[j].strip() == '}' or 
                       lines[j].strip().startswith('})')):
                    break
                j += 1
            
            if has_debug:
                # Skip all these lines
                continue
            else:
                # Keep the original line
                result.append(line)
                i += 1
                continue
        
        # Normal line - keep it
        result.append(line)
        i += 1
    
    return '\n'.join(result)

def process_file(filepath):
    """Process a single Go file to remove Debug logs."""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # Check if file has .Debug( calls
        if '.Debug(' not in content:
            return False
        
        # Remove debug logs
        new_content = remove_debug_logs(content)
        
        # Write back
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(new_content)
        
        return True
    except Exception as e:
        print(f"Error processing {filepath}: {e}", file=sys.stderr)
        return False

def main():
    """Main function."""
    base_dir = Path('/workspaces/CodeValdCortex/internal')
    
    if not base_dir.exists():
        print(f"Directory {base_dir} does not exist", file=sys.stderr)
        return 1
    
    print("Removing all .Debug() logging statements from Go files...")
    print()
    
    # Find all Go files
    go_files = list(base_dir.rglob('*.go'))
    
    if not go_files:
        print("No Go files found.")
        return 0
    
    modified_count = 0
    
    for go_file in go_files:
        if process_file(go_file):
            print(f"  ✓ {go_file.relative_to(base_dir.parent)}")
            modified_count += 1
    
    print()
    print(f"✅ All .Debug() logging statements removed.")
    print(f"Files modified: {modified_count}")
    
    return 0

if __name__ == '__main__':
    sys.exit(main())
