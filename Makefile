
.PHONY: default
default: build

.PHONY: build
build: test
	go build cmd/salus-packages-agent.go

.PHONY: clean
clean:
	rm -rf dist salus-packages-agent*

.PHONY: test
test:
	go test ./...

.PHONY: example
example:
	docker build -f Dockerfile.example .

.PHONY: release-snapshot
release-snapshot:
	goreleaser --snapshot --rm-dist

.PHONY: release-ci
release-ci:
	curl -sL https://git.io/goreleaser | bash