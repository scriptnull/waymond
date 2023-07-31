lint:
  docker run --rm -v $(pwd):/app -v ~/.cache/golangci-lint/:/root/.cache -w /app golangci/golangci-lint golangci-lint run -v

build: lint
  go build -o waymond cmd/waymond/main.go

site-install:
  pushd site && npm install && popd

deploy: site-install
  pushd site && USE_SSH=true npm run deploy && popd

run-site: site-install
  pushd site && npm run start && popd