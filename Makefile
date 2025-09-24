BUILDDIR := ./build/

build: builddir
	go build -o $(BUILDDIR)/artwork cmd/artwork/main.go

builddir:
	mkdir -p $(BUILDDIR)
