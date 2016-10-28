VERSION := $(shell cat VERSION)

deps:
	glide -q install

test: deps
	go test `glide nv`

build: test
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags production -ldflags "-X main.appVersion=$(VERSION)" -o bin/linux/amd64/v$(VERSION)/coldbrew
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -tags production -ldflags "-X main.appVersion=$(VERSION)" -o bin/linux/386/v$(VERSION)/coldbrew
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -tags production -ldflags "-X main.appVersion=$(VERSION)" -o bin/darwin/amd64/v$(VERSION)/coldbrew
	GOOS=darwin GOARCH=386 CGO_ENABLED=0 go build -tags production -ldflags "-X main.appVersion=$(VERSION)" -o bin/darwin/386/v$(VERSION)/coldbrew
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags production -ldflags "-X main.appVersion=$(VERSION)" -o bin/windows/amd64/v$(VERSION)/coldbrew.exe
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -tags production -ldflags "-X main.appVersion=$(VERSION)" -o bin/windows/386/v$(VERSION)/coldbrew.exe
	mkdir -p bin/linux/amd64/latest/; cp bin/linux/amd64/v$(VERSION)/coldbrew bin/linux/amd64/latest/coldbrew
	mkdir -p bin/linux/386/latest/; cp bin/linux/386/v$(VERSION)/coldbrew bin/linux/386/latest/coldbrew
	mkdir -p bin/darwin/amd64/latest/; cp bin/darwin/amd64/v$(VERSION)/coldbrew bin/darwin/amd64/latest/coldbrew
	mkdir -p bin/darwin/386/latest/; cp bin/darwin/386/v$(VERSION)/coldbrew bin/darwin/386/latest/coldbrew
	mkdir -p bin/windows/amd64/latest/; cp bin/windows/amd64/v$(VERSION)/coldbrew.exe bin/windows/amd64/latest/coldbrew.exe
	mkdir -p bin/windows/386/latest/; cp bin/windows/386/v$(VERSION)/coldbrew.exe bin/windows/386/latest/coldbrew.exe


.PHONY: deps test build