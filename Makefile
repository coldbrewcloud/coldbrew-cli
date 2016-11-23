VERSION := $(shell cat VERSION)
OSS := linux darwin windows

deps:
	glide -q install

test: deps
	go test `glide nv`

build: test
	@for OS in $(OSS); do \
		echo "Building $$OS..."; \
		export OUTFILE=coldbrew; if [ ! "$$OS" != "windows" ]; then export OUTFILE=coldbrew.exe; fi; \
		GOOS=$$OS GOARCH=amd64 CGO_ENABLED=0 go build -tags production -ldflags "-X main.appVersion=$(VERSION)" -o bin/$$OS/amd64/v$(VERSION)/$$OUTFILE; \
		(cd bin/$$OS/amd64/v$(VERSION); tar -cvzf coldbrew.tar.gz $$OUTFILE; rm $$OUTFILE); \
		GOOS=$$OS GOARCH=386 CGO_ENABLED=0 go build -tags production -ldflags "-X main.appVersion=$(VERSION)" -o bin/$$OS/386/v$(VERSION)/$$OUTFILE; \
		(cd bin/$$OS/386/v$(VERSION); tar -cvzf coldbrew.tar.gz $$OUTFILE; rm $$OUTFILE); \
		mkdir -p bin/$$OS/amd64/latest/; cp bin/$$OS/amd64/v$(VERSION)/* bin/$$OS/amd64/latest/; \
		mkdir -p bin/$$OS/386/latest/; cp bin/$$OS/386/v$(VERSION)/* bin/$$OS/386/latest/; \
	done

.PHONY: deps test build