package template

import "embed"

//go:embed app/*
var AppTemplates embed.FS
