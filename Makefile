APPNAME := netstatus
GO := go
# LDFLAGS := -s -w
LDGLAFS :=
SUBMAKE := make --no-print-directory

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: | build-linux build-windows build-darwin ## Build binaries

init:
	go get

_build:
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -ldflags "$(LDFLAGS)" -o build/$(OS)/$(ARCH)/$(BINNAME)

build-linux: init ## Build Linux binary
	@$(SUBMAKE) _build OS=linux ARCH=amd64 BINNAME=$(APPNAME)

build-windows: | init ## Build Windows binaries
	@$(SUBMAKE) _build OS=windows ARCH=386 BINNAME=$(APPNAME).exe
	@$(SUBMAKE) _build OS=windows ARCH=amd64 BINNAME=$(APPNAME).exe

build-darwin: | init ## Build Mac OS X binaries
	@$(SUBMAKE) _build OS=darwin ARCH=amd64 BINNAME=$(APPNAME)
