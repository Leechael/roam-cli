# Read Examples

Examples for reading pages, blocks, daily notes, and searching.

## Read today's daily page

    roam-cli get --today
    roam-cli get --today --json
    roam-cli get --daily yesterday

## Read a page or block

    roam-cli get "Page Title"
    roam-cli get "((block-uid))"
    roam-cli get "Page Title" --json

## Journal by date

    roam-cli journal --date today
    roam-cli journal --date yesterday --topic "Work Log"

## Search pages (default)

    roam-cli search "meeting" "action item" --limit 10
    roam-cli search "TODO" --daily-topic "[[TODO]]" --json

## Search individual blocks

    roam-cli search keyword --type block --limit 20
    roam-cli search keyword --type block --page "Project" -i

## Datalog query

    roam-cli q '[:find ?title :where [?e :node/title ?title]]'

## Find block UID (low-level)

    roam-cli block find --text "[[📖 Daily Reading]]" --today
    roam-cli block find --text "Status" --page "Project Dashboard"
