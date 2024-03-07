package main

import (
	"html/template"
	"time"
)

func unescapeHTML(s string) any {
	return template.HTML(s)
}

func copyright() string {
	return time.Now().Format("2006")
}
