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
RAW=false

usage() {
    cat << EOF
Usage: $0 <command> [options]

Commands:
  health          - Check debug server health
  state           - Get application state (UI, data, files)
  layout          - Get layout diagnostics
  snapshot        - Capture screen content
  logs            - Get debug event logs
  input <key>     - Send key input to application
  reset           - Reset application state

Options:
  --port <port>   - Debug server port (default: $DEFAULT_PORT)
  --host <host>   - Debug server host (default: $DEFAULT_HOST)
  --raw           - For snapshot: strip ANSI codes

Key Input Examples:
  tab, enter, escape, up, down, left, right, space
  a, u, r, l, e, c, q, /

Examples:
  $0 health
  $0 state
  $0 layout
  $0 snapshot --raw
  $0 logs
  $0 input tab
  $0 input enter
  $0 reset
EOF
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        health|state|layout|snapshot|logs|input|reset)
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
        --raw)
            RAW=true
            shift
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
        if [[ "$RAW" == true ]]; then
            make_get_request "/snapshot" "raw=true"
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

    *)
        echo "Error: Unknown command: $COMMAND" >&2
        usage >&2
        exit 1
        ;;
esac
