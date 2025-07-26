export GOTOOLCHAIN=local

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
