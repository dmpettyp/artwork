BUILDDIR := ./build/

run:
	go run ./cmd/artwork/

build: builddir
	go build -o $(BUILDDIR)/artwork cmd/artwork

builddir:
	mkdir -p $(BUILDDIR)
