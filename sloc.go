package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime/pprof"
	"sort"
	"text/tabwriter"
)

const VERSION = `1.3`

var languages = []Language{
	Language{"Thrift", mExt(".thrift"), cComments},

	Language{"C", mExt(".c", ".h"), cComments},
	Language{"C++", mExt(".cc", ".cpp", ".cxx", ".hh", ".hpp", ".hxx"), cComments},
	Language{"Go", mExt(".go"), cComments},
	Language{"Rust", mExt(".rs", ".rc"), cComments},
	Language{"Scala", mExt(".scala"), cComments},
	Language{"Java", mExt(".java"), cComments},
	Language{"Objc", mExt(".m"), cComments},
	Language{"Swift", mExt(".swift"), cComments},

	Language{"YACC", mExt(".y"), cComments},
	Language{"Lex", mExt(".l"), cComments},

	Language{"Fortran", mExt(".f90"), fortranComments},

	Language{"Lua", mExt(".lua"), luaComments},

	Language{"SQL", mExt(".sql"), sqlComments},

	Language{"Haskell", mExt(".hs", ".lhs"), hsComments},
	Language{"ML", mExt(".ml", ".mli"), mlComments},

	Language{"Perl", mExt(".pl", ".pm"), perlComments},
	Language{"PHP", mExt(".php"), cComments},

	Language{"Shell", mExt(".sh"), shComments},
	Language{"Bash", mExt(".bash"), shComments},
	Language{"R", mExt(".r", ".R"), shComments},
	Language{"Tcl", mExt(".tcl"), shComments},

	Language{"MATLAB", mExt(".m"), matlabComments},

	Language{"Julia", mExt(".jl"), juliaComments},
	Language{"Ruby", mExt(".rb"), rubyComments},
	Language{"Python", mExt(".py"), pyComments},
	Language{"Assembly", mExt(".asm", ".s"), semiComments},
	Language{"Lisp", mExt(".lsp", ".lisp"), semiComments},
	Language{"Scheme", mExt(".scm", ".scheme"), semiComments},

	Language{"Make", mName("makefile", "Makefile", "MAKEFILE"), shComments},
	Language{"CMake", mName("CMakeLists.txt"), shComments},
	Language{"Jam", mName("Jamfile", "Jamrules"), shComments},

	Language{"Markdown", mExt(".md"), noComments},

	Language{"HAML", mExt(".haml"), noComments},
	Language{"Sass", mExt(".sass"), cssComments},
	Language{"SCSS", mExt(".scss"), cssComments},

	Language{"HTML", mExt(".htm", ".html", ".xhtml"), xmlComments},
	Language{"XML", mExt(".xml"), xmlComments},
	Language{"CSS", mExt(".css"), cssComments},
	Language{"JavaScript", mExt(".js"), cComments},
	Language{"TypeScript", mExt(".ts"), cComments},
	Language{"CoffeeScript", mExt(".coffee"), coffeeComments},
	Language{"JSON", mExt(".json"), noComments},

	Language{"Erlang", mExt(".erl"), erlangComments},
}

type Commenter struct {
	LineComment  string
	StartComment string
	EndComment   string
	Nesting      bool
}

var (
	noComments      = Commenter{"\000", "\000", "\000", false}
	xmlComments     = Commenter{"\000", `<!--`, `-->`, false}
	fortranComments = Commenter{"!", "\000", "\000", false}
	cComments       = Commenter{`//`, `/*`, `*/`, false}
	cssComments     = Commenter{"\000", `/*`, `*/`, false}
	shComments      = Commenter{`#`, "\000", "\000", false}
	semiComments    = Commenter{`;`, "\000", "\000", false}
	hsComments      = Commenter{`--`, `{-`, `-}`, true}
	mlComments      = Commenter{`\000`, `(*`, `*)`, false}
	sqlComments     = Commenter{`--`, `/*`, `*/`, false}
	luaComments     = Commenter{`--`, `--[[`, `]]`, false}
	pyComments      = Commenter{`#`, `"""`, `"""`, false}
	matlabComments  = Commenter{`%`, `%{`, `%}`, false}
	erlangComments  = Commenter{`%`, "\000", "\000", false}
	rubyComments    = Commenter{`#`, "=begin", "=end", false}
	coffeeComments  = Commenter{`#`, "###", "###", false}
	juliaComments   = Commenter{`#`, "#=", "=#", true}

	// TODO support POD and __END__
	perlComments = Commenter{`#`, "\000", "\000", false}
)

type Language struct {
	Namer
	Matcher
	Commenter
}

// TODO work properly with unicode
func (l Language) Update(c []byte, s *Stats) {
	s.FileCount++

	commentLen := 0
	codeLen := 0
	inComment := 0 // this is an int for nesting
	inLComment := false
	lc := []byte(l.LineComment)
	sc := []byte(l.StartComment)
	ec := []byte(l.EndComment)
	lp, sp, ep := 0, 0, 0

	for _, b := range c {
		if b != byte(' ') && b != byte('\t') && b != byte('\n') && b != byte('\r') {
			if !inLComment && inComment == 0 {
				codeLen++
			} else {
				commentLen++
			}
		}
		if inComment == 0 && b == lc[lp] {
			lp++
			if lp == len(lc) {
				if !inLComment {
					codeLen -= lp
				}
				inLComment = true
				lp = 0
			}
		} else {
			lp = 0
		}
		if !inLComment && b == sc[sp] {
			sp++
			if sp == len(sc) {
				if inComment == 0 {
					codeLen -= sp
				}
				inComment++
				if inComment > 1 && !l.Nesting {
					inComment = 1
				}
				sp = 0
			}
		} else {
			sp = 0
		}
		if !inLComment && inComment > 0 && b == ec[ep] {
			ep++
			if ep == len(ec) {
				if inComment > 0 {
					inComment--
				}
				if inComment == 0 {
					commentLen -= ep
				}
				ep = 0
			}
		} else {
			ep = 0
		}

		if b == byte('\n') {
			s.TotalLines++
			if commentLen > 0 {
				s.CommentLines++
			}
			if codeLen > 0 {
				s.CodeLines++
			}
			if commentLen == 0 && codeLen == 0 {
				s.BlankLines++
			}
			inLComment = false
			codeLen = 0
			commentLen = 0
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

func handleFileLang(fname string, l Language) {
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

func handleFile(fname string) {
	for _, lang := range languages {
		if lang.Match(fname) {
			handleFileLang(fname, lang)
			return
		}
	}
	// TODO No recognized extension - check for hashbang
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
			if f.Name() == ".nosloc" {
				return
			}
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

func (d LData) Less(i, j int) bool {
	if d[i].CodeLines == d[j].CodeLines {
		return d[i].Name > d[j].Name
	}
	return d[i].CodeLines > d[j].CodeLines
}

func (d LData) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

type LResult struct {
	Name         string
	FileCount    int
	CodeLines    int
	CommentLines int
	BlankLines   int
	TotalLines   int
}

func (r *LResult) Add(a LResult) {
	r.FileCount += a.FileCount
	r.CodeLines += a.CodeLines
	r.CommentLines += a.CommentLines
	r.BlankLines += a.BlankLines
	r.TotalLines += a.TotalLines
}

func printJSON() {
	bs, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bs))
}

func printInfo() {
	w := tabwriter.NewWriter(os.Stdout, 2, 8, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "Language\tFiles\tCode\tComment\tBlank\tTotal\t")
	d := LData([]LResult{})
	total := &LResult{}
	total.Name = "Total"
	for n, i := range info {
		r := LResult{n, i.FileCount, i.CodeLines, i.CommentLines, i.BlankLines, i.TotalLines}
		d = append(d, r)
		total.Add(r)
	}
	d = append(d, *total)
	sort.Sort(d)
	for _, i := range d {
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%d\t\n", i.Name, i.FileCount, i.CodeLines, i.CommentLines, i.BlankLines, i.TotalLines)
	}

	w.Flush()
}

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	useJson    = flag.Bool("json", false, "JSON-format output")
	version    = flag.Bool("V", false, "display version info and exit")
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

	if *useJson {
		printJSON()
	} else {
		printInfo()
	}
}
