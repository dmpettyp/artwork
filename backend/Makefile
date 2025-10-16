BUILDDIR := ./build/

run:
	LOG_LEVEL=debug go run ./cmd/artwork/

build: builddir
	go build -o $(BUILDDIR)/artwork cmd/artwork

builddir:
	mkdir -p $(BUILDDIR)
