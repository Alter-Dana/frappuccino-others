#!/bin/bash


usage() {
    echo "usage: ./count_code_lines dir_path"
}

# Check if a directory path argument is provided
if [ $# -eq 0 ]; then
    usage
    exit 0
fi

DIR_PATH=$1

# Check if the argument is a directory
if [ ! -d "$DIR_PATH" ]; then
    if [ -e "$DIR_PATH" ]; then
        echo "error: $DIR_PATH is not a directory"
    else
        echo "error: directory $DIR_PATH not found"
    fi
    exit 0
fi

# Count the lines of code, excluding specified directories and empty lines
CODE_LINES=$(find "$DIR_PATH" \
  -type d \( -name node_modules -o -name build -o -name dest -o -name .git \) -prune -false \
  -o -type f -name "*.go" -print \
  | xargs grep -v '^\s*$' \
  | wc -l)

echo $CODE_LINES