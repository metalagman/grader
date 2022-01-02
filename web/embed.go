package web

import "embed"

//go:embed template/app/*
var AppTemplates embed.FS
