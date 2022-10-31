VERSION := $(shell git describe --tags --always --dirty="-dev")
LDFLAGS := -ldflags='-linkmode external -extldflags "-static" -X "main.version=$(VERSION)"'
PKG := ./cmd/worker
BIN := ./build/worker

# Formatter to use; select from `gofmt`, `goimports`, or `golines`
FORMATTER := golines

GOTESTFLAGS = -race
ifndef Q
GOTESTFLAGS += -v
endif

# Encourage use of ctlstore
export CGO_ENABLED=1
# Don't require dependency manager
export GO111MODULE=on
# Prevents having to provide auth inside Docker (and speeds things up)
export GOFLAGS=-mod=vendor

# To show commands and test output, run `make Q= <target>` (empty Q)

.PHONY: build
build:
	$Qgo build -o $(BIN) -a $(LDFLAGS) $(PKG)

.PHONY: install
install:
	$Qgo install -a $(LDFLAGS) $(PKG)

.PHONY: run
run:
	$Qdocker-compose up -d
	$Qgo run $(PKG)

.PHONY: vendor
vendor:
	$Qgo mod vendor

.PHONY: vet
vet:
	$Qgo vet ./...

.PHONY: generate
generate:
	$Qgo generate ./...

.PHONY: fmtchk
fmtchk: $(FORMATTER)
	$Qexit $(shell $(FORMATTER) -l . | grep -v '^vendor' | wc -l)

.PHONY: fmtfix
fmtfix: $(FORMATTER)
	$Q$(FORMATTER) -w $(shell find . -iname '*.go' | grep -v vendor)

.PHONY: goimports
goimports:
ifeq (, $(shell which goimports))
	$QGO111MODULE=off go get -u golang.org/x/tools/cmd/goimports
endif

.PHONY: golines
golines:
ifeq (, $(shell which golines))
	$QGO111MODULE=off go get -u github.com/segmentio/golines
endif

.PHONY: gofmt
gofmt:

