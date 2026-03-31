# Write Examples

Examples for saving GFM markdown and creating content.

## Save GFM markdown to today's daily page (recommended)

    printf '- journal entry' | roam-cli save --today
    printf '- entry' | roam-cli save --today --under '[[📽 Journaling]]'
    cat highlights.md | roam-cli save --today --under '[[📖 Daily Reading]]'

## Create TODOs

    printf '- {{[[TODO]]}} Review PR\n- {{[[TODO]]}} Call dentist' \
      | roam-cli save --today --under '[[TODO]]'

## Save to a named page

    cat note.md | roam-cli save --title "New Page"
    roam-cli save --title "Project X" --under '[[Tasks]]' --file ./tasks.md

## Get UID back for follow-up

    UID=$(printf '- item' | roam-cli save --today --under '[[Inbox]]' --plain)
    printf '- detail' | roam-cli save --parent "$UID"

## Low-level: block create (JSON input, explicit UIDs)

    roam-cli block create --parent <uid> --text "hello"
    echo '{"text":"Root","children":[{"text":"Child"}]}' \
      | roam-cli block create --parent <uid>
    roam-cli block create --parent <uid> --attach-to "[[Section]]" --file tree.json

## Low-level: batch operations

    roam-cli batch run --file ./actions.json
    echo '[...]' | roam-cli batch run

## Other block operations

    roam-cli block update --uid <uid> --text "updated"
    roam-cli block delete --uid <uid>
