#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <new-name>" >&2
  exit 1
fi

OLD="lib-config"
NEW="$1"

if [[ -z "$NEW" ]]; then
  echo "Error: new name cannot be empty" >&2
  exit 1
fi

if [[ ! "$NEW" =~ ^[a-z0-9-]+$ ]]; then
  echo "Error: name must match ^[a-z0-9-]+$" >&2
  exit 1
fi

if [[ "$NEW" == "$OLD" ]]; then
  echo "Error: new name is the same as old name" >&2
  exit 1
fi

echo "Renaming '$OLD' → '$NEW'"

echo "→ Replacing content in files..."

find . -depth -not -path './.git/*' -type f -print0 \
  | xargs -0 grep -rlI "$OLD" \
  | while IFS= read -r file; do
      sed -i "s|$OLD|$NEW|g" "$file"
      echo "  patched: $file"
    done

echo "→ Renaming files and directories..."

find . -depth -not -path './.git/*' -name "*${OLD}*" \
  | while IFS= read -r path; do
      new_path="${path//$OLD/$NEW}"
      mv "$path" "$new_path"
      echo "  renamed: $path → $new_path"
    done

echo "Done."