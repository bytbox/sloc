TARG = sloc
GOFILES = sloc.go

${TARG}: ${GOFILES}
	go build -x

clean:
	rm -f ${TARG}

fmt:
	gofmt -w *.go

.PHONY: install clean fmt
