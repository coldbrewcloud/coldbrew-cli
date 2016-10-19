deps:
	glide -q install

test: deps
	go test `glide nv`

build: test
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags production -o bin/linux/amd64/coldbrew
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -tags production -o bin/darwin/amd64/coldbrew
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags production -o bin/windows/amd64/coldbrew.exe

.PHONY: deps test build