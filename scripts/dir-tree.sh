#!/usr/bin/env bash

# A script to emulate the `tree` command with whitespace indentation,
# as a workaround for `tree` v2.2.1+ where this functionality is missing.

# --- Configuration and Argument Parsing ---

# Use the first argument as the target directory, or default to the current directory '.'
TARGET_DIR="${1:-.}"

# Use the second argument as the file extension to filter by
EXTENSION="${2}"

# Use the third argument as the indent size, or default to 2 spaces
INDENT_SIZE="${3:-2}"

# --- Input Validation ---

# Check if the target is a valid directory
if [ ! -d "$TARGET_DIR" ]; then
  echo "Error: Directory '$TARGET_DIR' not found." >&2
  exit 1
fi

# Check if the indent size is a positive integer
if ! [[ "$INDENT_SIZE" =~ ^[0-9]+$ ]]; then
    echo "Error: Indent size '$INDENT_SIZE' must be a positive integer." >&2
    exit 1
fi

# --- Main Logic ---

# Remove a potential trailing slash from the input for consistent depth calculation
TARGET_DIR_CLEAN="${TARGET_DIR%/}"

# Calculate the path depth of the starting directory to use as a baseline
BASE_DEPTH=$(echo "$TARGET_DIR_CLEAN" | awk -F'/' '{print NF}')

# Determine the paths to display
if [ -n "$EXTENSION" ]; then
  # Find all files with the given extension
  files=$(find "$TARGET_DIR_CLEAN" -type f -name "*.$EXTENSION")
  
  if [ -z "$files" ]; then
    echo "No *.$EXTENSION files found in '$TARGET_DIR_CLEAN'"
    exit 0
  fi
  
  # Get the directory of each file and build a unique, sorted list of all necessary paths
  all_paths=$( (echo "$files"; dirname -- "$files" | sed 's|/[^/]*$||' | sort -u) | sort -u)
  
  # This is getting complicated. Let's try another way.
  # We will build a list of all files and all their parent directories
  path_list=""
  for file in $files; do
    path_list="$path_list\n$file"
    dir=$(dirname -- "$file")
    while [[ "$dir" != "." && "$dir" != "/" && "$dir" != "$TARGET_DIR_CLEAN" ]]; do
      path_list="$path_list\n$dir"
      dir=$(dirname -- "$dir")
    done
  done
  # Get unique sorted list of paths
  all_paths=$(echo -e "$path_list" | grep -v '^$' | sort -u)
else
  # If no extension is provided, find all files and directories
  all_paths=$(find "$TARGET_DIR_CLEAN" -mindepth 1 | sort)
fi

# Print the root directory name first, since the loop will skip it
echo "$(basename "$TARGET_DIR_CLEAN")/"

# Loop through each path
echo "$all_paths" | while IFS= read -r path; do
  # Skip the root directory itself in the loop
  if [ "$path" == "$TARGET_DIR_CLEAN" ]; then
    continue
  fi

  # Calculate the depth of the current path by counting slashes
  CURRENT_DEPTH=$(echo "$path" | awk -F'/' '{print NF}')

  # Determine the relative depth for proper indentation
  relative_depth=$((CURRENT_DEPTH - BASE_DEPTH))

  # Create the indentation string based on the calculated depth and size
  indent=$(printf '%*s' $((relative_depth * INDENT_SIZE)) '')

  # Get just the final component of the path (the file or directory name)
  base=$(basename "$path")

  # Print the indented name, adding a '/' back if the path is a directory
  if [ -d "$path" ]; then
    echo "${indent}${base}/"
  else
    echo "${indent}${base}"
  fi
done