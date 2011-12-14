package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"text/tabwriter"
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		args = append(args, `.`)
	}

	for _, n := range args {
		add(n)
	}

	for _, f := range files {
		c, err := ioutil.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ! %s\n", f)
		}
		cs := string(c)
		handleFile(f, cs)
	}
	printInfo()
}

type Language interface{
	Name() string
	Update(string, *Stats)
}

type lName string

func (l lName) Name() string {
	return string(l)
}

type Stats struct{
	FileCount    int
	TotalLines   int
	CodeLines    int
	BlankLines   int
	CommentLines int
}

var info = map[string]*Stats{}

var languages = map[string]Language{
	".c": CLanguage{"C"},
	".go": CLanguage{"Go"},
}

func handleFile(fname, content string) {
	ext := path.Ext(fname)
	l, ok := languages[ext]
	if !ok {
		return // ignore this file
	}
	i, ok := info[l.Name()]
	if !ok {
		i = &Stats{}
		info[l.Name()] = i
	}
	c, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ! %s\n", fname)
		return
	}
	l.Update(string(c), i)
}

func printInfo() {
	w := tabwriter.NewWriter(os.Stdout, 2, 8, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "Language\tFiles\tLines\t")
	for n, i := range info {
		fmt.Fprintf(w, "%s\t%d\t%d\t\n", n, i.FileCount, i.TotalLines)
	}
	w.Flush()
}

var files []string

func add(n string) {
	fi, err := os.Stat(n)
	if err != nil {
		goto invalid
	}
	if fi.IsDir() {
		fs, err := ioutil.ReadDir(n)
		if err != nil {
			goto invalid
		}
		for _, f := range fs {
			if f.Name()[0] != '.' {
				add(path.Join(n, f.Name()))
			}
		}
		return
	}
	if fi.Mode() & os.ModeType == 0 {
		files = append(files, n)
		return
	}

	println(fi.Mode())

invalid:
	fmt.Fprintf(os.Stderr, "  ! %s\n", n)
}
