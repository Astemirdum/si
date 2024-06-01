.PHONY: build-compiler
build-compiler:
	go build -o ./a.out cmd/main.go;

.PHONY: compile
compile: build-compiler
	@echo "compile f=$(f)"
	./a.out "${f}"


.PHONY: example
example:
	make compile f=example/fib.si


.PHONY: lint
lint:
	golangci-lint run --fix

.PHONY: test
test:
	go test -v  -timeout 90s -count=1 -shuffle=on  -coverprofile cover.out ./...
	@go tool cover -func cover.out | grep total | awk '{print $3}'
	go tool cover -html="cover.out" -o coverage.html

