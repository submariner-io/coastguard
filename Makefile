BASE_BRANCH ?= devel
export BASE_BRANCH

ifneq (,$(DAPPER_HOST_ARCH))

CONTROLLER_GEN := $(CURDIR)/bin/controller-gen

# Running in Dapper

GO ?= go

# Ensure we prefer binaries we build
export PATH := $(CURDIR)/bin:$(PATH)

# Targets to make

# Download controller-gen locally if not already downloaded.
$(CONTROLLER_GEN):
	mkdir -p $(@D)
	$(GO) build -o $@ sigs.k8s.io/controller-tools/cmd/controller-gen

controller-gen: $(CONTROLLER_GEN)

# Generate deep-copy code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="$(CURDIR)/hack/boilerplate.go.txt,year=$(shell date +"%Y")" paths="./..."

IMAGES ?= coastguard
images: build

UNIT_TEST_ARGS := test/e2e

build: bin/coastguard-controller
bin/coastguard-controller: $(shell find pkg)
	${SCRIPTS_DIR}/compile.sh $@ pkg/coastguard/main.go

include $(SHIPYARD_DIR)/Makefile.inc

else

# Not running in Dapper

Makefile.dapper:
	@echo Downloading $@
	@curl -sflO https://raw.githubusercontent.com/submariner-io/shipyard/$(BASE_BRANCH)/$@

include Makefile.dapper

endif

# Disable rebuilding Makefile
Makefile Makefile.inc: ;
