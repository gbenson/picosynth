#!/bin/bash
set -euo pipefail

git config --global --add safe.directory $(pwd)

uid=$(stat -c '%u' "$0")
gid=$(stat -c '%g' "$0")

groupadd --gid "$gid" group
useradd --uid "$uid" --gid "$gid" user

exec su user -c "$(echo $@)"
