package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime/pprof"
	"sort"
	"text/tabwriter"
)

const VERSION = `0.1`

var languages = []Language{
	Language{"C", mExt(".c", ".h"), cComments},
	Language{"C++", mExt(".cc", ".cpp", ".cxx", ".hh", ".hpp", ".hxx"), cComments},
	Language{"Go", mExt(".go"), cComments},
	Language{"Haskell", mExt(".hs", ".lhs"), noComments},
	Language{"Perl", mExt(".pl", ".pm"), shComments},

	Language{"Shell", mExt(".sh"), shComments},
	Language{"Bash", mExt(".bash"), shComments},

	Language{"Python", mExt(".py"), noComments},
	Language{"Lisp", mExt(".lsp", ".lisp"), semiComments},
	Language{"Make", mName("makefile", "Makefile", "MAKEFILE"), shComments},
	Language{"HTML", mExt(".htm", ".html", ".xhtml"), noComments},
}

type Commenter struct {
	LineComment  string
	StartComment string
	EndComment   string
	Nesting      bool
}

var (
	noComments = Commenter{"\000", "\000", "\000", false}
	cComments  = Commenter{`//`, `/*`, `*/`, false}
	shComments = Commenter{`#`, "\000", "\000", false}
	semiComments = Commenter{`;`, "\000", "\000", false}
)

type Language struct {
	Namer
	Matcher
	Commenter
}

func (l Language) Update(c []byte, s *Stats) {
	s.FileCount++

	inComment := 0 // this is an int for nesting
	inLComment := false
	blank := true
	lc := []byte(l.LineComment)
	sc := []byte(l.StartComment)
	ec := []byte(l.EndComment)
	lp, sp, ep := 0, 0, 0

	for _, b := range c {
		if b == lc[lp] && !(inComment > 0) {
			lp++
			if lp == len(lc) {
				inLComment = true
				lp = 0
			}
		} else { lp = 0 }
		if b == sc[sp] && !inLComment {
			sp++
			if sp == len(sc) {
				inComment++
				if inComment > 1 && !l.Nesting {
					inComment = 1
				}
				sp = 0
			}
		} else { sp = 0 }
		if b == ec[ep] && !inLComment && inComment > 0 {
			ep++
			if ep == len(ec) {
				if inComment > 0 { inComment-- }
				ep = 0
			}
		} else { ep = 0 }

		if b != byte(' ') && b != byte('\t') && b != byte('\n') {
			blank = false
		}

		// BUG(srl): lines with comment don't count towards code
		// Note that lines with both code and comment count towards
		// each, but are not counted twice in the total.
		if b == byte('\n') {
			s.TotalLines++
			if inComment > 0 || inLComment {
				inLComment = false
				s.CommentLines++
			} else if blank {
				s.BlankLines++
			} else { s.CodeLines++ }
			blank = true
			continue
		}
	}
}

type Namer string

func (l Namer) Name() string { return string(l) }

type Matcher func(string) bool

func (m Matcher) Match(fname string) bool { return m(fname) }

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

type Stats struct {
	FileCount    int
	TotalLines   int
	CodeLines    int
	BlankLines   int
	CommentLines int
}

var info = map[string]*Stats{}

func handleFile(fname string) {
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
	l.Update(c, i)
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
	if fi.Mode()&os.ModeType == 0 {
		files = append(files, n)
		return
	}

	println(fi.Mode())

invalid:
	fmt.Fprintf(os.Stderr, "  ! %s\n", n)
}

type LData []LResult

func (d LData) Len() int { return len(d) }

func (d LData) Less(i, j int) bool { return d[i].CodeLines > d[j].CodeLines }

func (d LData) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

type LResult struct {
	Name string
	FileCount int
	CodeLines int
	CommentLines int
	BlankLines int
	TotalLines int
}

func printInfo() {
	w := tabwriter.NewWriter(os.Stdout, 2, 8, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "Language\tFiles\tCode\tComment\tBlank\tTotal\t")
	d := LData([]LResult{})
	for n, i := range info {
		d = append(d, LResult{
			n,
			i.FileCount,
			i.CodeLines,
			i.CommentLines,
			i.BlankLines,
			i.TotalLines,
		})
	}
	sort.Sort(d)
	for _, i := range d {
		fmt.Fprintf(
			w,
			"%s\t%d\t%d\t%d\t%d\t%d\t\n",
			i.Name,
			i.FileCount,
			i.CodeLines,
			i.CommentLines,
			i.BlankLines,
			i.TotalLines)
	}
	w.Flush()
}

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	version = flag.Bool("V", false, "display version info and exit")
)

func main() {
	flag.Parse()
	if *version {
		fmt.Printf("sloc %s\n", VERSION)
		return
	}
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
			return
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	args := flag.Args()
	if len(args) == 0 {
		args = append(args, `.`)
	}

	for _, n := range args {
		add(n)
	}

	for _, f := range files {
		handleFile(f)
	}
	printInfo()
}
