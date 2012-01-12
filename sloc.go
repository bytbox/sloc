package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
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

type Language struct{
	Namer
	Matcher
}

var (
	blankR = regexp.MustCompile(`^[ \t]*$`)
)

func (l Language) Update(c string, s *Stats) {
	s.FileCount++
	b := bytes.Buffer{}
	for _, r := range c {
		if r == '\n' {
			s.TotalLines++
			if blankR.Match(b.Bytes()) { s.BlankLines++ }
			b.Reset()
			continue
		}
		b.WriteRune(r)
	}
}

type Namer string

func (l Namer) Name() string {
	return string(l)
}

type Matcher func(string) bool

func (m Matcher) Match(fname string) bool {
	return m(fname)
}

func mExt(exts ...string) Matcher {
	return func(fname string) bool {
		for _, ext := range exts {
			if ext == path.Ext(fname) {
				return true
			}
		}
		return false
	}
}

func mName(names ...string) Matcher {
	return func(fname string) bool {
		for _, name := range names {
			if name == path.Base(fname) {
				return true
			}
		}
		return false
	}
}

type Stats struct{
	FileCount    int
	TotalLines   int
	CodeLines    int
	BlankLines   int
	CommentLines int
}

var info = map[string]*Stats{}

var languages = []Language{
	Language{"C", mExt(".c", ".h")},
	Language{"C++", mExt(".cc", ".cpp", ".cxx", ".hh", ".hpp", ".hxx")},
	Language{"Go", mExt(".go")},
	/*Language{"Haskell", mExt(".hs", ".lhs")},
	Language{"Python", mExt(".py")},
	Language{"Lisp", mExt(".lsp")},
	Language{"Make", mName("makefile", "Makefile", "MAKEFILE")},
	Language{"HTML", mExt(".htm", ".html", ".xhtml")},*/
}

func handleFile(fname, content string) {
	var l Language
	ok := false
	for _, lang := range languages {
		if lang.Match(fname) {
			ok = true
			l = lang
			break
		}
	}
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
	fmt.Fprintln(w, "Language\tFiles\tCode\tComment\tBlank\tTotal\t")
	for n, i := range info {
		fmt.Fprintf(
			w,
			"%s\t%d\t%d\t%d\t%d\t%d\t\n",
			n,
			i.FileCount,
			i.CodeLines,
			i.CommentLines,
			i.BlankLines,
			i.TotalLines)
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
