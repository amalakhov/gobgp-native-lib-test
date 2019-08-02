# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
BINARY_NAME=bgpush
BINARY_UNIX=$(BINARY_NAME)_unix

run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)
