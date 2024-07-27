build:
	go build -o build/meilibridge cmd/meilibridge/main.go

unit_test:
	go test ./...

docker:
	docker build -t ja7adr/meilibridge:latest .

fmt:
	gofumpt -l -w .

check:
	golangci-lint run --timeout=20m0s

devtools:
	@echo "Installing devtools"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59
	go install mvdan.cc/gofumpt@latest

.PHONY: build
.PHONY: unit_test fmt check
.PHONY: devtools