BUILDDIR := ./build/

run:
	LOG_LEVEL=debug go run cmd/artwork/main.go

build: builddir
	go build -o $(BUILDDIR)/artwork cmd/artwork/main.go

builddir:
	mkdir -p $(BUILDDIR)
