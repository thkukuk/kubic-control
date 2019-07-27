KUBICD_BIN := bin/kubicd
KUBICCTL_BIN := bin/kubicctl

GO ?= go
GO_MD2MAN ?= go-md2man

VERSION	:= $(shell cat VERSION)
PORT	:= $(shell cat PORT)
USE_VENDOR =
LOCAL_LDFLAGS = -buildmode=pie -ldflags "-X=main.Version=$(VERSION) \
	-X=main.port=$(PORT)\
	-X=github.com/thkukuk/kubic-control/pkg/kubicctl.Version=$(VERSION) \
	-X=github.com/thkukuk/kubic-control/pkg/kubicctl.port=$(PORT)"

.PHONY: all api build vendor
all: api dep build

api: ## Auto-generate grpc go sources
	@protoc -I api/ \
		--go_out=plugins=grpc:api \
		api/api.proto

dep: ## Get the dependencies
	@$(GO) get -v -d ./...

update: ## Get and update the dependencies
	@$(GO) get -v -d -u ./...

tidy: ## Clean up dependencies
	@$(GO) mod tidy

vendor: dep ## Create vendor directory
	@$(GO) mod vendor

build: ## Build the binary files
	$(GO) build -v -o $(KUBICD_BIN) $(USE_VENDOR) $(LOCAL_LDFLAGS) ./cmd/kubicd
	$(GO) build -v -o $(KUBICCTL_BIN) $(USE_VENDOR) $(LOCAL_LDFLAGS) ./cmd/kubicctl

clean: ## Remove previous builds
	@rm -f $(KUBICD_BIN) $(KUBICCTL_BIN) $(API_GO)

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


.PHONY: release
release: ## create release package from git
	git clone https://github.com/thkukuk/kubic-control
	mv kubic-control kubic-control-$(VERSION)
	sed -i -e 's|USE_VENDOR =|USE_VENDOR = -mod vendor|g' kubic-control-$(VERSION)/Makefile
	make -C kubic-control-$(VERSION) api
	make -C kubic-control-$(VERSION) vendor
	tar cJf kubic-control-$(VERSION).tar.xz kubic-control-$(VERSION)
	rm -rf kubic-control-$(VERSION)

#MANPAGES_MD := $(wildcard docs/man/*.md)
#MANPAGES    := $(MANPAGES_MD:%.md=%)

#docs/man/%.1: docs/man/%.1.md
#        $(GO_MD2MAN) -in $< -out $@

#.PHONY: docs
#docs: $(MANPAGES)

#.PHONY: install
#install:
#	$(GO) install $(LOCAL_LDFLAGS) ./cmd/...
