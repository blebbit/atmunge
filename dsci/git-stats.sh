#!/bin/bash
#
# Calculate git stats per day, comparing human vs gemini commits.
#
# Usage:
#   ./dsci/git-stats.sh [--since "YYYY-MM-DD"]
#
# If --since is not provided, it defaults to 1 week ago.

# --- Configuration ---

# Set the default since date if not provided
SINCE_DEFAULT="1 week ago"

# --- Argument Parsing ---

SINCE_ARG=""
while [[ $# -gt 0 ]]; do
  case $1 in
    --since)
      SINCE_ARG="$2"
      shift # past argument
      shift # past value
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# --- Main Logic ---

# Determine the since date
if [ -n "$SINCE_ARG" ]; then
  SINCE="$SINCE_ARG"
else
  SINCE="$SINCE_DEFAULT"
fi

START_DATE=$(date -d "$SINCE" +%F)
END_DATE=$(date +%F)

echo "Calculating git stats from $START_DATE to $END_DATE"
echo "Date,Human Commits,Human Lines,Gemini Commits,Gemini Lines"

# Loop through each day from START_DATE to END_DATE
current_date=$START_DATE
while [ "$current_date" != "$(date -d "$END_DATE + 1 day" +%F)" ]; do
    # --- Human Stats ---
    human_commits=$(git log --since="$current_date 00:00:00" --until="$current_date 23:59:59" --grep="^(human)" --oneline | wc -l)
    human_stats=$(git log --since="$current_date 00:00:00" --until="$current_date 23:59:59" --grep="^(human)" --pretty=tformat: --shortstat)
    human_lines=0
    if [ -n "$human_stats" ]; then
        insertions=$(echo "$human_stats" | awk '{s+=$4} END {print s}')
        deletions=$(echo "$human_stats" | awk '{s+=$6} END {print s}')
        human_lines=$((insertions + deletions))
    fi

    # --- Gemini Stats ---
    gemini_commits=$(git log --since="$current_date 00:00:00" --until="$current_date 23:59:59" --grep="^(gemini)" --oneline | wc -l)
    gemini_stats=$(git log --since="$current_date 00:00:00" --until="$current_date 23:59:59" --grep="^(gemini)" --pretty=tformat: --shortstat)
    gemini_lines=0
    if [ -n "$gemini_stats" ]; then
        insertions=$(echo "$gemini_stats" | awk '{s+=$4} END {print s}')
        deletions=$(echo "$gemini_stats" | awk '{s+=$6} END {print s}')
        gemini_lines=$((insertions + deletions))
    fi

    echo "$current_date,$human_commits,$human_lines,$gemini_commits,$gemini_lines"

    current_date=$(date -d "$current_date + 1 day" +%F)
done
