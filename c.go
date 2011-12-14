package main

import (
	"strings"
)

type CLanguage struct {
	lName
}

func (l CLanguage) Update(c string, i *Stats) {
	i.FileCount++

	lines := strings.Split(c, "\n")
	i.TotalLines += len(lines)
}
