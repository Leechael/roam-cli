# Exit Codes

Stable exit codes for scripting and automation.

## Overview

roam-cli uses a small set of exit codes to distinguish broad error categories.
Scripts can check the exit code to decide whether to retry, re-authenticate,
or report a missing resource.

## Codes

| Code | Name     | Meaning                          |
|------|----------|----------------------------------|
| 0    | OK       | Success                          |
| 1    | Error    | General error                    |
| 2    | Auth     | Authentication failure (401/403) |
| 3    | NotFound | Resource not found (404)         |

## Commands

All commands return these exit codes. The mapping is determined by the
HTTP status code returned by the Roam API.

## Constraints

- Exit codes 2 and 3 are only produced when the Roam API returns the
  corresponding HTTP status. Network errors and timeouts produce exit code 1.
- There is no exit code for rate limiting (429). The client retries
  automatically with exponential backoff; if all retries are exhausted,
  exit code 1 is returned.

## Examples

```bash
roam-cli status
echo $?    # 0 if OK, 2 if auth failed, 1 otherwise

roam-cli get "NonExistent Page"
echo $?    # 3 if page not found
```
