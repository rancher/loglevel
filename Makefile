# Modified by Acorn Labs
.PHONY: build
build:
	mkdir -p bin
	PLATFORM=$(shell uname)
    ifeq ($(PLATFORM),"Darwin")
      CGO_ENABLED=0 go build -ldflags "-extldflags -static -s" -o bin/loglevel
    else
      CGO_ENABLED=0 go build -o bin/loglevel
    endif
