include ${GOROOT}/src/Make.inc

TARG = sloc
GOFILES = sloc.go

include ${GOROOT}/src/Make.cmd

fmt:
	gofmt -w *.go

