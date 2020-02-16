PROGRAM := go-tilemap-xyz-grid-server
SRCS := $(shell find . -name '*.go')

.PHONY: build
build: $(PROGRAM)

$(PROGRAM): $(SRCS)
	go build -o $(PROGRAM) main.go
