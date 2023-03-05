lint:
  docker run --rm -v $(pwd):/app -v ~/.cache/golangci-lint/:/root/.cache -w /app golangci/golangci-lint golangci-lint run -v

build: lint
  go build -o waymond cmd/waymond/main.go