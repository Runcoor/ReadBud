#!/usr/bin/env bash
# Prepend AGPL-3.0 short header to all source files that don't already have one.
# Idempotent: re-running is a no-op for files already tagged.
#
# Skips:
#   - auto-generated files (*.d.ts, components.d.ts, auto-imports.d.ts)
#   - vendor / node_modules / dist
#   - migration .sql files (data, not code)
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

SLASH_HEADER='// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

'

HTML_HEADER='<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
'

has_header() {
  head -n 6 "$1" | grep -q 'SPDX-License-Identifier: AGPL-3.0'
}

prepend() {
  local file="$1"
  local header="$2"
  local tmp
  tmp="$(mktemp)"
  printf '%s' "$header" > "$tmp"
  cat "$file" >> "$tmp"
  mv "$tmp" "$file"
}

count_added=0
count_skipped=0

# --- Go files ---
while IFS= read -r f; do
  if has_header "$f"; then
    count_skipped=$((count_skipped + 1))
  else
    prepend "$f" "$SLASH_HEADER"
    count_added=$((count_added + 1))
  fi
done < <(find backend -name '*.go' -not -path '*/vendor/*')

# --- TypeScript files (skip generated) ---
while IFS= read -r f; do
  if has_header "$f"; then
    count_skipped=$((count_skipped + 1))
  else
    prepend "$f" "$SLASH_HEADER"
    count_added=$((count_added + 1))
  fi
done < <(find frontend/src -name '*.ts' \
            -not -name '*.d.ts' \
            -not -path '*/node_modules/*')

# --- JS files (extension) ---
while IFS= read -r f; do
  if has_header "$f"; then
    count_skipped=$((count_skipped + 1))
  else
    prepend "$f" "$SLASH_HEADER"
    count_added=$((count_added + 1))
  fi
done < <(find wechat-extension -name '*.js' -not -path '*/node_modules/*')

# --- Vue files (HTML-comment header) ---
while IFS= read -r f; do
  if head -n 6 "$f" | grep -q 'SPDX-License-Identifier: AGPL-3.0'; then
    count_skipped=$((count_skipped + 1))
  else
    prepend "$f" "$HTML_HEADER"
    count_added=$((count_added + 1))
  fi
done < <(find frontend/src -name '*.vue' -not -path '*/node_modules/*')

echo "added=${count_added} skipped=${count_skipped}"
