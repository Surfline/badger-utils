install:
	@dep ensure

fmt:
	@go fmt ./...

test:
	@go test -v -cover ./...

.PHONY: install fmt test
