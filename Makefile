deps:
	glide -q install

test: deps
	go test `glide nv`

build: test
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags production -o bin/linux/amd64/coldbrew
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -tags production -o bin/linux/386/coldbrew
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -tags production -o bin/darwin/amd64/coldbrew
	GOOS=darwin GOARCH=386 CGO_ENABLED=0 go build -tags production -o bin/darwin/386/coldbrew
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags production -o bin/windows/amd64/coldbrew.exe
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -tags production -o bin/windows/386/coldbrew.exe

.PHONY: deps test build