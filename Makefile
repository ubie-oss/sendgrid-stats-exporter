REPO      := sendgrid-stats-exporter
GIT_TAG   := $(shell git tag --points-at HEAD)
GIT_HASH  := $(shell git rev-parse HEAD)
IMAGE_URI := gcr.io/ubie-oss/$(REPO)
VERSION   := $(shell if [ -n "$(GIT_TAG)" ]; then echo "$(GIT_TAG)"; else echo "$(GIT_HASH)"; fi)
DIST_DIR  := $(shell if [ -n "$(GOOS)$(GOARCH)" ]; then echo "./dist/$(GOOS)-$(GOARCH)"; else echo "./dist"; fi)

default: build

.PHONY: build
build:
	@echo "version: $(VERSION) hash: $(GIT_HASH) tag: $(GIT_TAG)"
	go build -ldflags "-s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_HASH)" -o $(DIST_DIR)/exporter .

.PHONY: build-image
build-image:
	docker build -t "$(IMAGE_URI)" .
	docker tag "$(IMAGE_URI)":latest "$(IMAGE_URI)":"$(VERSION)"

.PHONY: push-image
push-image:
	docker push "$(IMAGE_URI)"

bin/golangci-lint:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s 'latest'

.PHONY: lint
lint: bin/golangci-lint
	./bin/golangci-lint run --tests

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test $(go list ./...)
