#!/bin/bash
# Restore custom documentation after Speakeasy generation
# Run this after: speakeasy run

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "Restoring custom documentation..."

# Restore docs/index.md from template
cp "$SCRIPT_DIR/templates/docs-index.md" "$PROJECT_DIR/docs/index.md"

# Remove auto-generated examples README (we use external examples repo)
rm -f "$PROJECT_DIR/examples/README.md"

echo "Done. Custom docs restored."
