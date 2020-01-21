
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

.PHONY: integration-test
integration-test:
	docker-compose up

.PHONY: example
example:
	docker build -t salus-packages-agent .
	docker run --rm salus-packages-agent
	docker run --rm salus-packages-agent --line-protocol-to-console

.PHONY: release-snapshot
release-snapshot:
	goreleaser --snapshot --rm-dist

.PHONY: release-ci
release-ci:
	curl -sL https://git.io/goreleaser | bash