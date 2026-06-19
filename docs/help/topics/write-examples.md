# Write Examples

Examples for saving GFM markdown and creating content.

## Save GFM markdown to today's daily page (recommended)

    printf '%s\n' '- journal entry' | roam-cli save --today
    printf '%s\n' '- entry' | roam-cli save --today --under '[[📽 Journaling]]'
    cat highlights.md | roam-cli save --today --under '[[📖 Daily Reading]]'

## Create TODOs

    printf '%s\n' '- {{[[TODO]]}} Review PR' '- {{[[TODO]]}} Call dentist' \
      | roam-cli save --today --under '[[TODO]]'

## Save to a named page

    cat note.md | roam-cli save --title "New Page"
    cat note.md | roam-cli save --title "Project X" --replace
    roam-cli save --title "Project X" --under '[[Tasks]]' --file ./tasks.md

## Get UID back for follow-up

    UID=$(printf '%s\n' '- item' | roam-cli save --today --under '[[Inbox]]' --plain)
    printf '%s\n' '- detail' | roam-cli save --parent "$UID"

## Low-level: block create (JSON input, explicit UIDs)

    roam-cli block create --parent UID_HERE --text "hello"
    echo '{"text":"Root","children":[{"text":"Child"}]}' \
      | roam-cli block create --parent UID_HERE
    roam-cli block create --parent UID_HERE --attach-to "[[Section]]" --file tree.json

## Page operations

    roam-cli page clear "Project X"
    roam-cli page delete "Project X"
    roam-cli page clear --today

## Low-level: batch operations

    roam-cli batch run --file ./actions.json
    echo '[...]' | roam-cli batch run

## Other block operations

    roam-cli block update --uid UID_HERE --text "updated"
    roam-cli block delete --uid UID_HERE
