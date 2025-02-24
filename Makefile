ifneq ($(wildcard .env),)
include .env
endif

SHELL := /usr/bin/env bash -o errexit -o pipefail -o nounset

DEBUG ?= 0
export DEBUG

GOTAGS ?=

PREFIX ?= blog-
SUFIX ?=

BINS = server cli
DIR ?= bin

GO ?= go

VERSION ?= release-$(shell git rev-parse HEAD | head -c8)

GOMODPATH := github.com/zanz1n/blog
LDFLAGS := -X $(GOMODPATH)/config.Version=$(VERSION)

ifeq ($(DEBUG), 1)
SUFIX += -debug
GOTAGS += debug
else
LDFLAGS += -s -w
endif

OS := $(if $(GOOS),$(GOOS),$(shell GOTOOLCHAIN=local $(GO) env GOOS))
ARCH := $(if $(GOARCH),$(GOARCH),$(shell GOTOOLCHAIN=local $(GO) env GOARCH))

ifeq ($(OS), windows)
SUFIX += .exe
endif

default: check build-server

all: $(addprefix build-, $(BINS))

run:
	cd web && bun run dev&
	air

build-server: generate

build-%: $(DIR)
ifneq ($(OUTPUT),) 
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -ldflags "$(LDFLAGS)" -tags "$(GOTAGS)" \
	-o $(OUTPUT) $(GOMODPATH)/cmd/$*
else
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -ldflags "$(LDFLAGS)" -tags "$(GOTAGS)"  \
	-o $(DIR)/$(PREFIX)$*-$(OS)-$(ARCH)$(SUFIX) $(GOMODPATH)/cmd/$*
endif
ifneq ($(POST_BUILD_CHMOD),)
	chmod $(POST_BUILD_CHMOD) $(DIR)/$(PREFIX)$*-$(OS)-$(ARCH)$(SUFIX)
endif

$(DIR):
	mkdir $(DIR)

TESTFLAGS := -v -race

ifeq ($(SHORTTESTS), 1)
TESTFLAGS += -short
endif

ifeq ($(NOTESTCACHE), 1)
TESTFLAGS += -count=1
endif

test:
ifneq ($(SKIPTESTS), 1)
	$(GO) test ./... $(TESTFLAGS)
else
    $(warning skipped tests)
endif

bench:
	$(GO) test -bench=. -count 3 -benchmem -run=^# ./...

.SILENT: gen-checksums
gen-checksums: $(DIR)
	DIR=$(DIR) ./scripts/gen-checksums.sh

check: deps generate test

update:
	$(GO) mod tidy
	$(GO) get -u ./...
	$(GO) mod tidy

deps:
	$(GO) install github.com/a-h/templ/cmd/templ@latest

generate:
	templ generate

fmt:
	$(GO) fmt ./...

debug:
	@echo DEBUG = $(DEBUG)
	@echo DIR = $(DIR)
	@echo BINNAME = $(PREFIX)%-$(OS)-$(ARCH)$(SUFIX)
	@echo GOMODPATH = $(GOMODPATH)
	@echo VERSION = $(VERSION)
	@echo BINS = $(BINS)
	@echo LDFLAGS = $(LDFLAGS)
