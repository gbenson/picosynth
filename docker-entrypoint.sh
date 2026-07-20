#!/bin/sh
git config --global --add safe.directory $(pwd)
"$@"
