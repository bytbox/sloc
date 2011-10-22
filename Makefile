PARTS = main language
SOURCES = ${PARTS:=.cxx}
OBJECTS = ${PARTS:=.o}

all: sloc

depend: .depend
.depend: ${SOURCES}
	${CXX} ${CXXFLAGS} -MM ${SOURCES} > $@
-include .depend

sloc: ${OBJECTS}
	${CXX} -o $@ ${OBJECTS}

.cxx.o:
	${CXX} -c ${CXXFLAGS} -o $@ $<

clean:
	${RM} *.o sloc .depend

.SUFFIXES: .cxx .o
.PHONY: all clean

