#!/bin/bash

# Extract the direct dependencies from go.mod
dependencies=$(grep -E '^\s*[^#]*\s(v[0-9]+\.[0-9]+\.[0-9]+|v[0-9]+\.[0-9]+\.[0-9]+-[0-9]{14}-[a-f0-9]+)$' go.mod | awk '{print $1}')

# Update each dependency
for dep in $dependencies; do
  echo "Updating $dep..."
  go get -u "$dep"
done

echo "Tidying up the go.mod and go.sum..."
go mod tidy -compat=1.18

read -p "Do you want to commit and push the changes? (y/n): " confirm
if [[ $confirm == [yY] ]]; then
  git add go.mod go.sum
  commit_message="chore: update direct dependencies and tidy go.mod"
  git commit -m "$commit_message"
  git push
  echo "Changes have been committed and pushed."
else
  echo "Changes were not committed or pushed."
fi
