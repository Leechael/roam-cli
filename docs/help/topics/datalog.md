# Datalog Queries

Roam Research exposes a Datomic-flavored Datalog query API with Clojure built-in functions.

## Overview

The `roam-cli q` command sends a raw Datalog query to the Roam API. The query
dialect is Datomic Datalog, which means Clojure functions are available in
predicate clauses via `[(...)]` syntax.

Key attributes in the Roam schema:

| Attribute          | Type   | Description                          |
|--------------------|--------|--------------------------------------|
| `:node/title`      | string | Page title                           |
| `:block/string`    | string | Block text content                   |
| `:block/uid`       | string | Block unique identifier              |
| `:block/page`      | ref    | Page the block belongs to            |
| `:block/children`  | ref    | Direct child blocks                  |
| `:block/parents`   | ref    | Ancestor blocks (transitive)         |
| `:block/refs`      | ref    | Page/block references within text    |
| `:block/order`     | long   | Sort order among siblings            |
| `:create/time`     | long   | Creation timestamp (epoch ms)        |
| `:edit/time`       | long   | Last edit timestamp (epoch ms)       |

## Commands

- `roam-cli q '<query>'` -- run a Datalog query
- `roam-cli q '<query>' --json` -- output as JSON
- `roam-cli q '<query>' --json --jq '<expr>'` -- filter with jq

## Constraints

- The query string must be a valid EDN vector starting with `[:find ...]`.
- Only `:find` and `:where` clauses are supported. `:in` is not available;
  use string interpolation or `--args` instead.
- The Clojure standard library is partially available. Functions under
  `clojure.string/` and `clojure.core/` work. Not all Clojure functions
  are guaranteed.
- Regex must use `re-pattern` and `re-find` as separate predicate clauses
  (inline regex literals like `#"..."` do not work in the API).
- `pull` expressions support `*` (all attributes) and recursive wildcards
  like `{:block/children ...}`.
- Aggregates `(count ...)` and `(count-distinct ...)` are supported in
  `:find`.

## Clojure Functions Reference

### String operations

```clojure
;; Substring match (case-sensitive)
[(clojure.string/includes? ?s "term")]

;; Case-insensitive match
[(clojure.string/includes? (clojure.string/lower-case ?s) "term")]

;; Starts-with / ends-with
[(clojure.string/starts-with? ?s "prefix")]
[(clojure.string/ends-with? ?s "suffix")]
```

### Regex

Inline `#"..."` regex does not work in the API. Use `re-pattern` to compile
the pattern, then `re-find` to match:

```clojure
[(re-pattern "https?://x\\.com") ?pattern]
[?node :block/string ?text]
[(re-find ?pattern ?text)]
```

This is the Roam-specific idiom -- break `re-pattern` into its own clause.

### Arithmetic and comparison

```clojure
[(> ?order 2)]
[(< ?time 1700000000000)]
[(+ ?a ?b) ?sum]
```

## Examples

List all page titles:

```bash
roam-cli q '[:find ?title :where [?e :node/title ?title]]'
```

Find blocks containing a substring (case-insensitive):

```bash
roam-cli q '[:find ?uid ?s
 :where
   [?b :block/string ?s]
   [?b :block/uid ?uid]
   [(clojure.string/includes? (clojure.string/lower-case ?s) "meeting")]]' \
  --json
```

Find blocks matching a regex pattern:

```bash
roam-cli q '[:find (pull ?node [:block/uid :block/string])
 :where
   [(re-pattern "https?://x\\.com") ?pattern]
   [?node :block/string ?text]
   [(re-find ?pattern ?text)]]' \
  --json
```

Pull full page tree with children:

```bash
roam-cli q '[:find (pull ?e [* {:block/children ...}])
 :where [?e :node/title "My Page"]]' --json
```

Count blocks per page:

```bash
roam-cli q '[:find ?title (count ?b)
 :where
   [?b :block/page ?p]
   [?p :node/title ?title]]' --json
```

Get blocks created after a timestamp:

```bash
roam-cli q '[:find ?uid ?s
 :where
   [?b :block/uid ?uid]
   [?b :block/string ?s]
   [?b :create/time ?t]
   [(> ?t 1700000000000)]]' --json
```

## Related Topics

- `roam-cli help exit-codes`
