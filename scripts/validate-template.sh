#!/bin/bash
set -e

# InfraTales Template Validation Script
# Run this before committing to ensure template integrity

echo "🔍 InfraTales Template Validation"
echo "=================================="

ERRORS=0

# Check required files
echo "📁 Checking required files..."
required_files=(
  "README.md"
  "LICENSE"
  "CONTRIBUTING.md"
  "CODE_OF_CONDUCT.md"
  "SECURITY.md"
  "CHANGELOG.md"
  "QUICK_START.md"
  "docs/architecture.md"
  "docs/cost.md"
  "docs/security.md"
  "docs/runbook.md"
  "docs/troubleshooting.md"
  "diagrams/architecture.mmd"
  ".github/workflows/ci.yml"
)

for file in "${required_files[@]}"; do
  if [ ! -f "$file" ]; then
    echo "  ❌ Missing: $file"
    ((ERRORS++))
  else
    echo "  ✅ $file"
  fi
done

# Check for unreplaced placeholders
echo ""
echo "🔎 Checking for unreplaced placeholders..."
if grep -r "{{.*}}" --include="*.md" --include="*.yml" . 2>/dev/null; then
  echo "  ⚠️  Found unreplaced placeholders (this is OK for templates)"
else
  echo "  ✅ No placeholders found"
fi

# Validate Mermaid diagrams (basic syntax check)
echo ""
echo "📊 Validating Mermaid diagrams..."
for diagram in diagrams/*.mmd; do
  if [ -f "$diagram" ]; then
    if grep -q "flowchart\|sequenceDiagram\|graph" "$diagram"; then
      echo "  ✅ $(basename $diagram)"
    else
      echo "  ❌ $(basename $diagram) - Invalid syntax"
      ((ERRORS++))
    fi
  fi
done

# Check markdown formatting
echo ""
echo "📝 Checking markdown files..."
if command -v markdownlint &> /dev/null; then
  markdownlint -c .markdownlint.json . || ERRORS=$((ERRORS + 1))
else
  echo "  ⚠️  markdownlint not installed, skipping"
fi

# Check for broken links (basic check)
echo ""
echo "🔗 Checking for broken internal links..."
if grep -r "\](file://" --include="*.md" . 2>/dev/null | grep -v "node_modules"; then
  echo "  ⚠️  Found absolute file links (use relative paths)"
fi

# Summary
echo ""
echo "=================================="
if [ $ERRORS -eq 0 ]; then
  echo "✅ Validation passed! Template is ready."
  exit 0
else
  echo "❌ Validation failed with $ERRORS error(s)."
  exit 1
fi
