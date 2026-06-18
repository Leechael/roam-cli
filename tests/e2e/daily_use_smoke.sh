#!/usr/bin/env bash
set -euo pipefail

# Manual-only smoke tests for Daily Use flows.
# CI does not call this script. It writes scratch pages to the configured Roam graph.
# Set DAILY_USE_SMOKE_KEEP_PAGES=1 to keep scratch pages after failure for manual inspection.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TMP_DIR="$(mktemp -d)"
KEEP_PAGES="${DAILY_USE_SMOKE_KEEP_PAGES:-${PR13_SMOKE_KEEP_PAGES:-0}}"
TIMEOUT_SECONDS="${DAILY_USE_SMOKE_TIMEOUT_SECONDS:-${PR13_SMOKE_TIMEOUT_SECONDS:-60}}"
RUN_ID="$(date +%s)-$$"
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
    log "DAILY_USE_SMOKE_KEEP_PAGES=1; leaving scratch pages for inspection: ${PAGES[*]:-<none>}"
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
  NEW_PAGE="roam-cli-daily-use-smoke-$RUN_ID-$suffix"
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

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local message="$3"
  if ! grep -Fq "$needle" <<<"$haystack"; then
    fail "$message"
  fi
}

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local message="$3"
  if grep -Fq "$needle" <<<"$haystack"; then
    fail "$message"
  fi
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
log "run id: $RUN_ID"
log "using CLI: ${CLI[*]}"
log "request timeout: ${TIMEOUT_SECONDS}s"
must_cli status

# 1. page delete must remove a page, not just send a block delete action for a page UID.
section "TEST 1: page delete removes a scratch page"
new_page delete
delete_page="$NEW_PAGE"
must_pipe_cli ignored '- temporary content' save --title "$delete_page"
must_cli get "$delete_page"
assert_contains "$LAST_OUT" 'temporary content' "created page content was not readable before delete"
must_cli page delete "$delete_page"
assert_page_deleted "$delete_page"
pass "page delete removes a scratch page"

# 2. move --under must not create the destination section when the source block UID is invalid.
section "TEST 2: move --under does not create section when source UID is invalid"
new_page move-invalid
move_page="$NEW_PAGE"
missing_section="[[Daily Use Smoke Should Not Exist $RUN_ID]]"
bad_uid="bad-daily-use-$RUN_ID"
log "bad source uid: $bad_uid"
log "section that must not appear: $missing_section"
must_pipe_cli ignored '- existing content' save --title "$move_page"
if capture_cli move_out move --uid "$bad_uid" --title "$move_page" --under "$missing_section"; then
  fail "move --under unexpectedly succeeded with invalid source UID"
fi
must_cli get "$move_page"
assert_not_contains "$LAST_OUT" "$missing_section" "move --under created the section before validating the source UID"
pass "move --under does not pollute the page when source UID is invalid"

# 3. save --plain should return the new top-level content block UID for follow-up nesting.
section "TEST 3: save --plain returns a UID usable for nested follow-up content"
new_page plain
plain_page="$NEW_PAGE"
inbox="[[Daily Use Smoke Inbox $RUN_ID]]"
parent_text="Daily Use smoke parent $RUN_ID"
child_text="Daily Use smoke child $RUN_ID"
log "inbox section: $inbox"
log "parent text: $parent_text"
log "child text: $child_text"
must_pipe_cli parent_uid "- $parent_text" save --title "$plain_page" --under "$inbox" --plain
log "save --plain returned uid: $parent_uid"
must_pipe_cli ignored "- $child_text" save --parent "$parent_uid"
must_cli get "$plain_page"
if ! grep -qx "    $child_text" <<<"$LAST_OUT"; then
  fail "save --plain did not return the new parent block UID for nested follow-up content"
fi
must_cli get "$parent_uid"
assert_contains "$LAST_OUT" "$parent_text" "get block UID did not include saved parent text"
assert_contains "$LAST_OUT" "$child_text" "get block UID did not include nested child text"
pass "save --plain supports follow-up nesting under the saved block"

# 4. save --replace should clear existing page content before writing replacement content.
section "TEST 4: save --replace replaces named page content"
new_page replace
replace_page="$NEW_PAGE"
old_text="Daily Use old replace content $RUN_ID"
new_text="Daily Use new replace content $RUN_ID"
must_pipe_cli ignored "- $old_text" save --title "$replace_page"
must_pipe_cli ignored "- $new_text" save --title "$replace_page" --replace
must_cli get "$replace_page"
assert_contains "$LAST_OUT" "$new_text" "replace page is missing new content"
assert_not_contains "$LAST_OUT" "$old_text" "replace page still contains old content"
pass "save --replace replaces named page content"

# 5. search should find scratch content by page and by block.
section "TEST 5: search page/block modes find scratch content"
search_token="DailyUseSearchToken$RUN_ID"
new_page search
search_page="$NEW_PAGE"
must_pipe_cli ignored "- $search_token page result" save --title "$search_page"
must_cli search "$search_token" --type page --limit 1
assert_contains "$LAST_OUT" "$search_page" "search --type page did not include scratch page"
must_cli search "$search_token" --type block --page "$search_page" --limit 1
assert_contains "$LAST_OUT" "$search_token" "search --type block did not include scratch block text"
pass "search page/block modes find scratch content"

# 6. move should move an existing block to a named page section.
section "TEST 6: move sends an existing block to a named page section"
new_page move-source
move_source_page="$NEW_PAGE"
new_page move-target
move_target_page="$NEW_PAGE"
move_text="Daily Use moved block $RUN_ID"
move_section="[[Daily Use Move Target $RUN_ID]]"
must_pipe_cli move_uid "- $move_text" save --title "$move_source_page" --plain
log "move source uid: $move_uid"
must_pipe_cli ignored '- target seed' save --title "$move_target_page"
must_cli move --uid "$move_uid" --title "$move_target_page" --under "$move_section"
must_cli get "$move_target_page"
assert_contains "$LAST_OUT" "$move_section" "move target page is missing destination section"
assert_contains "$LAST_OUT" "$move_text" "move target page is missing moved block"
pass "move sends an existing block to a named page section"

# 7. page clear should remove page content while keeping the page readable.
section "TEST 7: page clear removes content and keeps page"
new_page clear
clear_page="$NEW_PAGE"
clear_text="Daily Use clear content $RUN_ID"
must_pipe_cli ignored "- $clear_text" save --title "$clear_page"
must_cli page clear "$clear_page"
must_cli get "$clear_page"
assert_not_contains "$LAST_OUT" "$clear_text" "page clear left old content behind"
pass "page clear removes content and keeps page"

printf 'DONE: Daily Use smoke tests passed.\n'
