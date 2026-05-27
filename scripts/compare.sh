#!/usr/bin/env bash
# Side-by-side comparison harness for GopherCram and repomix.
#
# This script:
#   1. Builds the gophercram binary.
#   2. Runs gophercram against the bundled fixture in every output style.
#   3. Runs `npx --yes repomix` against the same fixture in matching styles.
#   4. Reports file count, size, and a list of present / missing files.
#
# Requirements:
#   - Go (to build gophercram)
#   - npx (to invoke repomix)
#   - jq (optional; used for prettier comparison; falls back to grep otherwise)
#
# Usage: ./scripts/compare.sh
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
FIXTURE="$ROOT_DIR/test/fixtures/dummy-ts-repo"
WORK="$(mktemp -d -t gophercram-compare.XXXXXX)"
trap 'rm -rf "$WORK"' EXIT

echo "--> building gophercram"
go -C "$ROOT_DIR" build -o "$WORK/gophercram" ./cmd/gophercram

run_gc() {
  local style="$1" out="$2"
  "$WORK/gophercram" --style "$style" --output "$out" --quiet "$FIXTURE"
}

run_rmx() {
  local style="$1" out="$2"
  # The first invocation downloads repomix into the npx cache; subsequent runs
  # are fast. We don't install it globally.
  npx --yes repomix --style "$style" --output "$out" --quiet "$FIXTURE" 2>/dev/null
}

# ---- XML
echo "--> packing in XML"
run_gc xml "$WORK/gc.xml"
run_rmx xml "$WORK/rmx.xml" || echo "repomix XML pack failed"
echo "  gc.xml  $(wc -c < "$WORK/gc.xml")  bytes"
[[ -f "$WORK/rmx.xml" ]] && echo "  rmx.xml $(wc -c < "$WORK/rmx.xml") bytes"

# ---- Markdown
echo "--> packing in Markdown"
run_gc markdown "$WORK/gc.md"
run_rmx markdown "$WORK/rmx.md" || echo "repomix markdown pack failed"
echo "  gc.md  $(wc -c < "$WORK/gc.md")  bytes"
[[ -f "$WORK/rmx.md" ]] && echo "  rmx.md $(wc -c < "$WORK/rmx.md") bytes"

# ---- JSON
echo "--> packing in JSON"
run_gc json "$WORK/gc.json"
run_rmx json "$WORK/rmx.json" || echo "repomix json pack failed"

# ---- Inclusion / exclusion check
echo "--> verifying expected files"
EXPECTED=(
  "README.md"
  "package.json"
  "tsconfig.json"
  ".env.example"
  "src/index.ts"
  "src/api/users.ts"
  "src/api/products.ts"
  "src/utils/format.ts"
  "src/utils/validate.ts"
  "src/components/Button.tsx"
  "src/components/Modal.tsx"
  "tests/api.test.ts"
  "tests/utils.test.ts"
  "docs/README.md"
)
EXCLUDED=(
  ".env"
  "build.log"
  "node_modules/fake-dep/index.js"
)

for f in "${EXPECTED[@]}"; do
  if grep -q "$f" "$WORK/gc.md"; then
    printf "  [ok ] gc  includes %s\n" "$f"
  else
    printf "  [BAD] gc  missing  %s\n" "$f"
  fi
  if [[ -f "$WORK/rmx.md" ]]; then
    if grep -q "$f" "$WORK/rmx.md"; then
      printf "  [ok ] rmx includes %s\n" "$f"
    else
      printf "  [BAD] rmx missing  %s\n" "$f"
    fi
  fi
done

echo "--> verifying excluded files (gitignore + default patterns)"
for f in "${EXCLUDED[@]}"; do
  if grep -q "\"$f\"\|\b$f\b" "$WORK/gc.md"; then
    printf "  [BAD] gc  contains %s\n" "$f"
  else
    printf "  [ok ] gc  excludes %s\n" "$f"
  fi
  if [[ -f "$WORK/rmx.md" ]]; then
    if grep -q "\"$f\"\|\b$f\b" "$WORK/rmx.md"; then
      printf "  [BAD] rmx contains %s\n" "$f"
    else
      printf "  [ok ] rmx excludes %s\n" "$f"
    fi
  fi
done

echo "--> security scan check (src/config.ts should be flagged)"
if grep -q "config.ts" "$WORK/gc.md"; then
  if grep -q "security" "$WORK/gc.md"; then
    echo "  [ok ] gc  reported security findings"
  else
    echo "  [info] gc  reported config.ts but no security mention in output (check stderr)"
  fi
fi

echo "Outputs left in $WORK (will be cleaned on exit)"
