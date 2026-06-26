.PHONY: build test run tidy fmt vet clean install uninstall

BIN := bin/bako
PKG := ./cmd/bako

# bako は cgo を使わないので既定で無効化する。
export CGO_ENABLED ?= 0

# Override with `make install PREFIX=$HOME/.local` to avoid sudo.
PREFIX ?= /usr/local
INSTALL_DIR := $(DESTDIR)$(PREFIX)/bin

build:
	@mkdir -p bin
	go build -o $(BIN) $(PKG)

test:
	go test ./...

run: build
	@$(BIN) $(ARGS)

tidy:
	go mod tidy

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -rf bin dist

install: build
	@mkdir -p $(INSTALL_DIR)
	install -m 0755 $(BIN) $(INSTALL_DIR)/bako

uninstall:
	rm -f $(INSTALL_DIR)/bako
