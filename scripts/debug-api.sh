#!/bin/bash

# Debug API Helper Script
# Provides easy access to claude-permissions debug server endpoints
# Designed for Claude Code/AI usage - not intended for humans

set -euo pipefail

# Configuration
DEFAULT_PORT=8080
DEFAULT_HOST="localhost"

# Parse command line arguments
COMMAND=""
PORT="$DEFAULT_PORT"
HOST="$DEFAULT_HOST"
KEY=""
COLOR=false
USER_FILE=""
REPO_FILE=""
LOCAL_FILE=""

usage() {
    cat << EOF
Usage: $0 <command> [options]

Commands:
  health                    - Check debug server health
  state                     - Get application state (UI, data, files)
  layout                    - Get layout diagnostics
  snapshot                  - Capture screen content
  logs                      - Get debug event logs
  input <key>               - Send key input to application
  reset                     - Reset application state
  launch-confirm-changes    - Launch confirmation screen with mock changes
  load-settings             - Load settings from specified file paths

Options:
  --port <port>     - Debug server port (default: $DEFAULT_PORT)
  --host <host>     - Debug server host (default: $DEFAULT_HOST)
  --color           - For snapshot: include ANSI color codes (default: stripped)
  --user-file <path>   - For load-settings: path to user settings file
  --repo-file <path>   - For load-settings: path to repo settings file
  --local-file <path>  - For load-settings: path to local settings file

Key Input Examples:
  tab, enter, escape, up, down, left, right, space
  a, u, r, l, e, c, q, /, 1, 2, 3

Examples:
  $0 health
  $0 state
  $0 layout
  $0 snapshot --color
  $0 logs
  $0 input tab
  $0 input enter
  $0 reset
  $0 launch-confirm-changes
  $0 load-settings --user-file testdata/user-no-duplicates.json --repo-file testdata/repo-no-duplicates.json
EOF
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        health|state|layout|snapshot|logs|input|reset|launch-confirm-changes|load-settings)
            COMMAND="$1"
            shift
            ;;
        --port)
            PORT="$2"
            shift 2
            ;;
        --host)
            HOST="$2"
            shift 2
            ;;
        --color)
            COLOR=true
            shift
            ;;
        --user-file)
            USER_FILE="$2"
            shift 2
            ;;
        --repo-file)
            REPO_FILE="$2"
            shift 2
            ;;
        --local-file)
            LOCAL_FILE="$2"
            shift 2
            ;;
        --help|-h)
            usage
            exit 0
            ;;
        *)
            if [[ "$COMMAND" == "input" && -z "$KEY" ]]; then
                KEY="$1"
                shift
            else
                echo "Unknown option: $1" >&2
                usage >&2
                exit 1
            fi
            ;;
    esac
done

# Validate command
if [[ -z "$COMMAND" ]]; then
    echo "Error: No command specified" >&2
    usage >&2
    exit 1
fi

# Validate key for input command
if [[ "$COMMAND" == "input" && -z "$KEY" ]]; then
    echo "Error: Key required for input command" >&2
    usage >&2
    exit 1
fi

# Base URL
BASE_URL="http://$HOST:$PORT"

# Helper function to make GET requests
make_get_request() {
    local endpoint="$1"
    local query_params="${2:-}"

    local url="$BASE_URL$endpoint"
    if [[ -n "$query_params" ]]; then
        url="$url?$query_params"
    fi

    if ! curl -s -f "$url"; then
        echo "Error: Failed to connect to debug server at $BASE_URL" >&2
        echo "Make sure the application is running with --debug-server flag" >&2
        exit 1
    fi
}

# Helper function to make POST requests
make_post_request() {
    local endpoint="$1"
    local data="$2"

    local url="$BASE_URL$endpoint"

    if ! curl -s -f -X POST -H "Content-Type: application/json" -d "$data" "$url"; then
        echo "Error: Failed to send POST request to debug server at $BASE_URL" >&2
        echo "Make sure the application is running with --debug-server flag" >&2
        exit 1
    fi
}

# Execute command
case "$COMMAND" in
    health)
        make_get_request "/health"
        ;;

    state)
        make_get_request "/state"
        ;;

    layout)
        make_get_request "/layout"
        ;;

    snapshot)
        if [[ "$COLOR" == true ]]; then
            make_get_request "/snapshot" "color=true"
        else
            make_get_request "/snapshot"
        fi
        ;;

    logs)
        make_get_request "/logs"
        ;;

    input)
        make_post_request "/input" "{\"key\":\"$KEY\"}"
        ;;

    reset)
        make_post_request "/reset" "{}"
        ;;

    launch-confirm-changes)
        # Create mock data with permissions that exist in test data
        mock_data='{
            "mock_changes": {
                "permission_moves": [
                    {
                        "name": "Read",
                        "from": "User",
                        "to": "Local"
                    },
                    {
                        "name": "Bash",
                        "from": "User",
                        "to": "Repo"
                    },
                    {
                        "name": "Glob",
                        "from": "Local",
                        "to": "User"
                    }
                ],
                "duplicate_resolutions": []
            }
        }'
        make_post_request "/launch-confirm-changes" "$mock_data"
        ;;

    load-settings)
        # Build JSON payload for load-settings
        json_parts=()

        if [[ -n "$USER_FILE" ]]; then
            json_parts+=("\"user_file\":\"$USER_FILE\"")
        fi

        if [[ -n "$REPO_FILE" ]]; then
            json_parts+=("\"repo_file\":\"$REPO_FILE\"")
        fi

        if [[ -n "$LOCAL_FILE" ]]; then
            json_parts+=("\"local_file\":\"$LOCAL_FILE\"")
        fi

        # Join parts with commas and build JSON
        if [[ ${#json_parts[@]} -gt 0 ]]; then
            json_content=$(printf "%s," "${json_parts[@]}")
            json_content="${json_content%,}"  # Remove trailing comma
            json_data="{$json_content}"
        else
            json_data="{}"
        fi

        make_post_request "/load-settings" "$json_data"
        ;;

    *)
        echo "Error: Unknown command: $COMMAND" >&2
        usage >&2
        exit 1
        ;;
esac
