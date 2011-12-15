package main

import (
	"strings"
)

type LineLanguage struct {
	lName
	lMatch
}

func (l LineLanguage) Update(c string, i *Stats) {
	i.FileCount++

	lines := strings.Split(c, "\n")
	i.TotalLines += len(lines)
}
