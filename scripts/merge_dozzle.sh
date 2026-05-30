#!/usr/bin/env bash
# Script to merge the Dozzle repository into this repo.
# Assumes a local clone of dozzle exists at ./dozzle

set -euo pipefail

# 1. Verify we are in the root of this repository
if [ ! -f "go.mod" ] && [ ! -d ".git" ]; then
  echo "Error: Not in a Git repository root."
  exit 1
fi

# 2. Check that dozzle folder exists and is not empty
if [ ! -d "dozzle" ]; then
  echo "Error: dozzle directory not found."
  exit 1
fi

# 3. Copy all files except .git into current repo, guard if source==dest
if [ "$PWD" = "$(cd dozzle && pwd)" ]; then
  echo "Source and destination directories are the same. Skipping copy."
else
  rsync -a --exclude='.git' dozzle/ .
fi

# 4. Stage and commit changes
if git diff --quiet; then
  echo "No changes to commit."
else
  git add .
  git commit -m "Merge Dozzle repository into main project"
fi

# 5. Remove the old dozzle directory
rm -rf dozzle

echo "Dozzle merged successfully."
