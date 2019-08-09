TARGETS := $(shell ls scripts | grep -v e2e)

.dapper:
	@echo Downloading dapper
	@curl -sL https://releases.rancher.com/dapper/latest/dapper-`uname -s`-`uname -m` > .dapper.tmp
	@@chmod +x .dapper.tmp
	@./.dapper.tmp -v
	@mv .dapper.tmp .dapper

$(TARGETS): .dapper
	./.dapper -m bind $@

e2e: .dapper ./hacking/e2e_subm.sh
ifneq ($(status),clean)
	./hacking/e2e_subm.sh
	./.dapper -m bind e2e $(status)
endif

e2e-coastguard: .dapper
	./.dapper -m bind e2e keep

ifneq ($(status),keep)
	./hacking/e2e_subm.sh clean
endif

.DEFAULT_GOAL := ci

.PHONY: $(TARGETS)

