#!/bin/sh
set -e

echo current version: $(gobump show -r)
read -p "input next version: " next_version

gobump set $next_version -w
ghch -w -N v$next_version

git add version.go CHANGELOG.md
git ci -m "Checking in changes prior to tagging of version v$next_version"
git tag v$next_version
git push && git push --tags
