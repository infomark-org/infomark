.PHONY: cover lint fmt build

test:
	go test ./... -covermode=atomic -coverprofile=coverage.txt

cover: test ## Run all the tests and opens the coverage report
	go tool cover -html=coverage.txt

fmt: ## Run goimports on all go files
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do goimports -w "$$file"; done

lint: ## Run all the linters
	golangci-lint run -D errcheck

build:
	go build infomark.go
