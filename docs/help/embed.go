package helpdocs

import "embed"

//go:embed topics/*.md
var Topics embed.FS
