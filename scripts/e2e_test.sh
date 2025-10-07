#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
EXAMPLE_DIR="$ROOT_DIR/pkg/objects/examples/e2e-openlane"
SETUP_BIN="$EXAMPLE_DIR/bin/setup"
MAIN_BIN="$EXAMPLE_DIR/bin/e2e-openlane"
BENCHMARK_TOKEN="$EXAMPLE_DIR/.benchmark-token"
BENCHMARK_ORG="$EXAMPLE_DIR/.benchmark-org-id"
TMP_DIR="$ROOT_DIR/tmp"
TMP_TOKEN="$TMP_DIR/e2e-openlane.token"
EVIDENCE_FILE="$ROOT_DIR/pkg/objects/examples/multi-tenant/testdata/sample-files/document.json"

ensure_dirs() {
  mkdir -p "$TMP_DIR" "$EXAMPLE_DIR/bin"
}

build_examples() {
  ensure_dirs
  echo "[e2e] building example binaries"
  (cd "$EXAMPLE_DIR" && go build -tags examples -o "$SETUP_BIN" ./setup)
  (cd "$EXAMPLE_DIR" && go build -tags examples -o "$MAIN_BIN" .)
}

ensure_binaries() {
  if [ ! -x "$SETUP_BIN" ] || [ ! -x "$MAIN_BIN" ]; then
    build_examples
  fi
}

run_setup() {
  ensure_binaries
  echo "[e2e] running setup to provision benchmark user and PAT"
  (
    cd "$EXAMPLE_DIR"
    ./bin/setup
  )
}

export_pat() {
  run_setup
  ensure_dirs
  local token
  token=$(tr -d '\n' < "$BENCHMARK_TOKEN")
  if [[ -z "$token" || $token == dummy-token || $token != tolp_* ]]; then
    echo "[e2e] setup did not produce a valid PAT" >&2
    exit 1
  fi
  printf '%s\n' "$token" > "$TMP_TOKEN"
  echo "[e2e] wrote token to $TMP_TOKEN"
}

run_example() {
  if [ ! -f "$BENCHMARK_TOKEN" ] || [ ! -f "$BENCHMARK_ORG" ]; then
    echo "[e2e] benchmark assets missing; run setup/export first" >&2
    exit 1
  fi
  if [ ! -f "$EVIDENCE_FILE" ]; then
    echo "[e2e] evidence file not found: $EVIDENCE_FILE" >&2
    exit 1
  fi
  local token
  token=$(tr -d '\n' < "$BENCHMARK_TOKEN")
  if [[ -z "$token" || $token == dummy-token || $token != tolp_* ]]; then
    echo "[e2e] invalid PAT detected; aborting" >&2
    exit 1
  fi
  echo "[e2e] running e2e example"
  "$MAIN_BIN" -token "$token" -file "$EVIDENCE_FILE"
}

cleanup() {
  echo "[e2e] cleaning artifacts"
  rm -f "$BENCHMARK_TOKEN" "$BENCHMARK_ORG" "$TMP_TOKEN"
}

usage() {
  cat <<USAGE
Usage: ${0##*/} [command]
  build   Build example binaries only
  setup   Build and run setup to create benchmark assets
  export  Ensure setup ran and export PAT to tmp token file
  run     Build, setup, export, and execute the example
  clean   Remove generated artifacts
  help    Show this message
USAGE
}

cmd=${1:-run}
case "$cmd" in
  build)
    build_examples
    ;;
  setup)
    ensure_binaries
    run_setup
    ;;
  export)
    ensure_binaries
    export_pat
    ;;
  run)
    ensure_binaries
    export_pat
    run_example
    ;;
  clean)
    cleanup
    ;;
  help|-h|--help)
    usage
    ;;
  *)
    echo "Unknown command: $cmd" >&2
    usage >&2
    exit 1
    ;;
esac
