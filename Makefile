BASE_BRANCH ?= devel
export BASE_BRANCH

ifneq (,$(DAPPER_HOST_ARCH))

# Running in Dapper

IMAGES ?= coastguard
images: build

UNIT_TEST_ARGS := test/e2e

build: bin/coastguard-controller
bin/coastguard-controller: vendor/modules.txt $(shell find pkg)
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
