install:
	go get ./...

test: install
	go test -v

fmt:
	gofmt -w *.go

tags:
	find ./ -name '*.go' -print0 | xargs -0 gotags > TAGS
