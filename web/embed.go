package web

import "embed"

//go:embed template/app/*.html
var AppTemplates embed.FS
