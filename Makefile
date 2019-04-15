KUBICD_BIN := bin/kubicd
KUBICCTL_BIN := bin/kubicctl
API_GO := api/api.pb.go
PKG := "github.com/thkukuk/kubic-control"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

GO ?= go
GO_MD2MAN ?= go-md2man

VERSION	:= $(shell cat VERSION)
PORT	:= $(shell cat PORT)
LOCAL_LDFLAGS = -ldflags "-X=main.Version=$(VERSION) -X=main.port=$(PORT)\
	-X=github.com/thkukuk/kubic-control/pkg/kubicctl.Version=$(VERSION) \
	-X=github.com/thkukuk/kubic-control/pkg/kubicctl.port=$(PORT)"

.PHONY: all api build vendor
all: build

$(API_GO): api/api.proto
	@protoc -I api/ \
		--go_out=plugins=grpc:api \
		api/api.proto

api: $(API_GO) ## Auto-generate grpc go sources

dep: ## Get the dependencies
	@$(GO) get -v -d ./...

update: ## Get and update the dependencies
	@$(GO) get -v -d -u ./...

tidy: ## Clean up dependencies
	@$(GO) mod tidy

vendor: ## Create vendor directory
	@$(GO) mod vendor

build: api dep ## Build the binary files
	$(GO) build -i -v -o $(KUBICD_BIN) $(LOCAL_LDFLAGS) ./cmd/kubicd
	$(GO) build -i -v -o $(KUBICCTL_BIN) $(LOCAL_LDFLAGS) ./cmd/kubicctl

clean: ## Remove previous builds
	@rm -f $(KUBICD_BIN) $(KUBICCTL_BIN) $(API_GO)

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


.PHONY: release
release: ## create release package from git
	-git clone https://github.com/thkukuk/kubic-control
	-mv kubic-control kubic-control-$(VERSION)
	-make -C kubic-control-$(VERSION) vendor
	-tar cJf kubic-control-$(VERSION).tar.xz kubic-control-$(VERSION)
	-rm -rf kubic-control-$(VERSION)

#MANPAGES_MD := $(wildcard docs/man/*.md)
#MANPAGES    := $(MANPAGES_MD:%.md=%)

#docs/man/%.1: docs/man/%.1.md
#        $(GO_MD2MAN) -in $< -out $@

#.PHONY: docs
#docs: $(MANPAGES)

.PHONY: install
install:
	$(GO) install $(LOCAL_LDFLAGS) ./cmd/...
