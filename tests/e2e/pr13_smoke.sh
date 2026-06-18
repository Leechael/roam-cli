#!/usr/bin/env bash
set -euo pipefail

# Manual-only smoke tests for PR #13 command layering changes.
# CI does not call this script. It writes scratch pages to the configured Roam graph.
# Set PR13_SMOKE_KEEP_PAGES=1 to keep scratch pages after failure for manual inspection.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TMP_DIR="$(mktemp -d)"
KEEP_PAGES="${PR13_SMOKE_KEEP_PAGES:-0}"
TIMEOUT_SECONDS="${PR13_SMOKE_TIMEOUT_SECONDS:-60}"
CLI=()
PAGES=()
NEW_PAGE=""
LAST_OUT=""

log() {
  printf '[%s] %s\n' "$(date '+%H:%M:%S')" "$*" >&2
}

section() {
  printf '\n[%s] %s\n' "$(date '+%H:%M:%S')" "$*" >&2
}

dump_output() {
  local label="$1"
  local content="$2"
  if [[ -z "$content" ]]; then
    log "$label: <empty>"
    return
  fi
  printf '[%s] %s:\n' "$(date '+%H:%M:%S')" "$label" >&2
  printf '%s\n' "$content" >&2
}

cleanup() {
  local status=$?
  set +e
  if [[ "$KEEP_PAGES" == "1" ]]; then
    log "PR13_SMOKE_KEEP_PAGES=1; leaving scratch pages for inspection: ${PAGES[*]:-<none>}"
  elif [[ "${#CLI[@]}" -gt 0 ]]; then
    log "cleaning scratch pages: ${PAGES[*]:-<none>}"
    for page in "${PAGES[@]}"; do
      log "+ roam-cli --timeout $TIMEOUT_SECONDS page clear $page"
      "${CLI[@]}" --timeout "$TIMEOUT_SECONDS" page clear "$page" >/dev/null 2>&1
      log "+ roam-cli --timeout $TIMEOUT_SECONDS page delete $page"
      "${CLI[@]}" --timeout "$TIMEOUT_SECONDS" page delete "$page" >/dev/null 2>&1
    done
  fi
  rm -rf "$TMP_DIR"
  exit "$status"
}
trap cleanup EXIT

require_env() {
  local missing=0
  for key in ROAM_API_TOKEN ROAM_API_GRAPH; do
    if [[ -z "${!key:-}" ]]; then
      printf 'MISSING: %s\n' "$key" >&2
      missing=1
    else
      printf 'OK: %s set\n' "$key" >&2
    fi
  done
  if [[ "$missing" -ne 0 ]]; then
    exit 2
  fi
}

build_cli() {
  if [[ -n "${ROAM_CLI:-}" ]]; then
    CLI=("$ROAM_CLI")
    log "using ROAM_CLI override: ${CLI[*]}"
    return
  fi
  local bin="$TMP_DIR/roam-cli"
  log "building fresh binary: $bin"
  (cd "$ROOT_DIR" && go build -o "$bin" ./cmd/roam-cli)
  CLI=("$bin")
}

new_page() {
  local suffix="$1"
  NEW_PAGE="roam-cli-pr13-smoke-$(date +%s)-$$-$suffix"
  PAGES+=("$NEW_PAGE")
  log "scratch page[$suffix]: $NEW_PAGE"
}

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  exit 1
}

pass() {
  printf 'PASS: %s\n' "$1"
}

capture_cli() {
  local out_var="$1"
  shift
  local cmd_out rc
  log "+ roam-cli --timeout $TIMEOUT_SECONDS $*"
  set +e
  cmd_out="$("${CLI[@]}" --timeout "$TIMEOUT_SECONDS" "$@" 2>&1)"
  rc=$?
  set -e
  log "exit=$rc"
  dump_output "output" "$cmd_out"
  LAST_OUT="$cmd_out"
  printf -v "$out_var" '%s' "$cmd_out"
  return "$rc"
}

must_cli() {
  local out
  if ! capture_cli out "$@"; then
    fail "command failed: roam-cli $*"
  fi
}

pipe_cli() {
  local out_var="$1"
  local input="$2"
  shift 2
  local cmd_out rc
  log "+ printf input | roam-cli --timeout $TIMEOUT_SECONDS $*"
  dump_output "stdin" "$input"
  set +e
  cmd_out="$(printf '%s\n' "$input" | "${CLI[@]}" --timeout "$TIMEOUT_SECONDS" "$@" 2>&1)"
  rc=$?
  set -e
  log "exit=$rc"
  dump_output "output" "$cmd_out"
  LAST_OUT="$cmd_out"
  printf -v "$out_var" '%s' "$cmd_out"
  return "$rc"
}

must_pipe_cli() {
  local out_var="$1"
  local input="$2"
  shift 2
  if ! pipe_cli "$out_var" "$input" "$@"; then
    fail "command failed: printf input | roam-cli $*"
  fi
}

assert_page_deleted() {
  local page="$1"
  local out
  if capture_cli out get "$page"; then
    fail "page still exists after page delete: $page"
  fi
  if [[ "$out" != *"not found"* ]]; then
    fail "expected get to report not found after page delete: $page"
  fi
}

require_env
build_cli

log "root dir: $ROOT_DIR"
log "temp dir: $TMP_DIR"
log "using CLI: ${CLI[*]}"
log "request timeout: ${TIMEOUT_SECONDS}s"
must_cli status

# 1. page delete must remove a page, not just send a block delete action for a page UID.
section "TEST 1: page delete removes a scratch page"
new_page delete
delete_page="$NEW_PAGE"
must_pipe_cli ignored '- temporary content' save --title "$delete_page"
must_cli get "$delete_page"
must_cli page delete "$delete_page"
assert_page_deleted "$delete_page"
pass "page delete removes a scratch page"

# 2. move --under must not create the destination section when the source block UID is invalid.
section "TEST 2: move --under does not create section when source UID is invalid"
new_page move
move_page="$NEW_PAGE"
missing_section="[[PR13 Smoke Should Not Exist $$]]"
bad_uid="bad-pr13-$$"
log "bad source uid: $bad_uid"
log "section that must not appear: $missing_section"
must_pipe_cli ignored '- existing content' save --title "$move_page"
if capture_cli move_out move --uid "$bad_uid" --title "$move_page" --under "$missing_section"; then
  fail "move --under unexpectedly succeeded with invalid source UID"
fi
must_cli get "$move_page"
move_page_out="$LAST_OUT"
if grep -Fq "$missing_section" <<<"$move_page_out"; then
  fail "move --under created the section before validating the source UID"
fi
pass "move --under does not pollute the page when source UID is invalid"

# 3. save --plain should return the new top-level content block UID for follow-up nesting.
section "TEST 3: save --plain returns a UID usable for nested follow-up content"
new_page plain
plain_page="$NEW_PAGE"
inbox="[[PR13 Smoke Inbox $$]]"
parent_text="PR13 smoke parent $$"
child_text="PR13 smoke child $$"
log "inbox section: $inbox"
log "parent text: $parent_text"
log "child text: $child_text"
must_pipe_cli parent_uid "- $parent_text" save --title "$plain_page" --under "$inbox" --plain
log "save --plain returned uid: $parent_uid"
must_pipe_cli ignored "- $child_text" save --parent "$parent_uid"
must_cli get "$plain_page"
plain_page_out="$LAST_OUT"
if ! grep -qx "    $child_text" <<<"$plain_page_out"; then
  fail "save --plain did not return the new parent block UID for nested follow-up content"
fi
pass "save --plain supports follow-up nesting under the saved block"

printf 'DONE: PR #13 smoke tests passed.\n'
