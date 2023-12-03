package embedded

import "embed"

//go:embed certs templates public build.json
var FS embed.FS
