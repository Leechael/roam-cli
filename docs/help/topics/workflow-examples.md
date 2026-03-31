# Workflow Examples

Common daily workflows and multi-step patterns.

## Daily capture (most common operation)

    # Quick note to today's journal section
    printf '- Had a great idea about X' | roam-cli save --today --under '[[📽 Journaling]]'

    # Meeting notes under today's page
    cat meeting.md | roam-cli save --today --under '[[Meeting Notes]]'

## Morning review

    # Check today's daily page
    roam-cli get --today

    # Check yesterday's journal
    roam-cli journal --date yesterday

## Organize: move blocks to project pages

    # Move a block from daily page to a project
    roam-cli move --uid <block> --title "Project X" --under '[[Tasks]]'

    # Move to today's archive section
    roam-cli move --uid <block> --today --under '[[Archive]]'

## Save and follow up (composing commands)

    # Save content, get UID, add more under it
    UID=$(printf '- headline' | roam-cli save --today --under '[[📽 Journaling]]' --plain)
    printf '- detail 1\n- detail 2' | roam-cli save --parent "$UID"

## Build a tree in one call (not N calls)

    # WRONG: sequential block create calls
    #   uid1=$(roam-cli block create --parent $PAGE --text "Project A" --plain)
    #   roam-cli block create --parent $uid1 --text "Task 1"

    # RIGHT: single call with JSON tree
    echo '{"text":"Project A","children":[
      {"text":"Task 1"},
      {"text":"Task 2"}
    ]}' | roam-cli block create --parent "$PAGE"

## Batch: move + update multiple blocks

    echo '[
      {"action":"move-block","location":{"parent-uid":"target","order":"last"},"block":{"uid":"abc"}},
      {"action":"update-block","block":{"uid":"abc","string":"moved and updated"}}
    ]' | roam-cli batch run

## Prefer printf piping over --text flags

    # Shell-safe: printf preserves [[ ]] and emoji
    printf '- [[📽 Journaling]] entry with [[links]]' | roam-cli save --today
