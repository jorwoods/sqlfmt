export GOTOOLCHAIN=local

OS   ?= linux
ARCH ?= amd64
BIN  := sqlfmt$(if $(filter windows,$(OS)),.exe,)

build:
	GOOS=$(OS) GOARCH=$(ARCH) go build -o dist/$(OS)_$(ARCH)/$(BIN) .

generate:
	go generate ./parser

test:
	go test ./...

clean:
	for file in parser/*.go; do \
		if [ "$$file" != "parser/generate.go" ]; then \
			rm -f "$$file"; \
		fi; \
	done
	rm -f parser/*.interp parser/*.tokens
