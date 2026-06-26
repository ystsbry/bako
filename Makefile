.PHONY: build test run tidy fmt vet clean install uninstall \
	install-plugin uninstall-plugin install-codex-skills uninstall-codex-skills \
	install-codex-plugin uninstall-codex-plugin

BIN := bin/bako
PKG := ./cmd/bako

# bako は cgo を使わないので既定で無効化する。
export CGO_ENABLED ?= 0

# Override with `make install PREFIX=$HOME/.local` to avoid sudo.
PREFIX ?= /usr/local
INSTALL_DIR := $(DESTDIR)$(PREFIX)/bin

# Plugin (skills) install locations.
#   Claude Code: ~/.claude/skills/bako  (skills-dir plugin; auto-loads,
#                                        invoked as /bako:<skill>)
#   Codex CLI:   ~/.agents/skills/<skill>  (via scripts/install-codex.sh)
CLAUDE_SKILLS_DIR ?= $(HOME)/.claude/skills
PLUGIN_NAME := bako
PLUGIN_SRC := $(CURDIR)/plugin
PLUGIN_MARKETPLACE := bako

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

# Symlink the whole plugin/ directory into ~/.claude/skills/bako so Claude
# Code auto-loads it as the bako@skills-dir plugin. Skills are then invoked as
# /bako:<skill>. Restart Claude Code to pick it up; verify with
# `claude plugin list`.
install-plugin:
	@mkdir -p $(CLAUDE_SKILLS_DIR)
	@target=$(CLAUDE_SKILLS_DIR)/$(PLUGIN_NAME); \
	if [ -L $$target ]; then \
		rm -f $$target; \
	elif [ -e $$target ]; then \
		echo "skip: $$target already exists (not a symlink)"; \
		exit 0; \
	fi; \
	ln -s $(PLUGIN_SRC) $$target; \
	echo "Linked $$target -> $(PLUGIN_SRC)"; \
	echo "Restart Claude Code, then verify with: claude plugin list"

uninstall-plugin:
	@target=$(CLAUDE_SKILLS_DIR)/$(PLUGIN_NAME); \
	if [ -L $$target ]; then \
		rm -f $$target; \
		echo "Removed $$target"; \
	fi

# Codex install option (1): per-skill directory symlink into ~/.agents/skills,
# invoked as $register-repo. This is the default/fallback.
install-codex-skills:
	@$(CURDIR)/scripts/install-codex.sh

uninstall-codex-skills:
	@$(CURDIR)/scripts/install-codex.sh --uninstall

# Codex install option (2): as a Codex plugin via a local marketplace
# (.agents/plugins/marketplace.json points at ./plugin). Codex COPIES the
# plugin into its cache, so re-run after editing skills. Exact `codex plugin`
# subcommands depend on your codex version.
install-codex-plugin:
	codex plugin marketplace add $(CURDIR)
	@echo "Added marketplace '$(PLUGIN_MARKETPLACE)'. Verify/enable with 'codex plugin', then restart codex."

uninstall-codex-plugin:
	-codex plugin marketplace remove $(PLUGIN_MARKETPLACE)
