.PHONY: clean build test test_new list all

.ONESHELL:
#SHELL =
#.SHELLFLAGS =

# endly version
# https://github.com/viant/endly
endly_version=0.52.0

# Inject into binary via linker:
# ...in github actions comes from make -e version=git_ref
version=$(shell cat VERSION)
commit=$(shell git show --no-patch --format=format:%H HEAD)
buildVersionVar=github.com/CLIP-HPC/goslmailer/internal/version.buildVersion
buildCommitVar=github.com/CLIP-HPC/goslmailer/internal/version.buildCommit

# various directories
bindirs=$(wildcard ./cmd/*)
installdir=build/goslmailer-$(version)

# list of files to include in build
bins=$(notdir $(bindirs))
readme=README.md
templates=templates/adaptive_card_template.json templates/telegramTemplate.html templates/matrix_template.md
config=cmd/goslmailer/goslmailer.conf.annotated_example cmd/gobler/gobler.conf

# can be replaced with go test ./... construct
testdirs=$(sort $(dir $(shell find ./ -name *_test.go)))

all: list test build get_endly test_endly install

list:
	@echo "================================================================================"
	@echo "bindirs  found: $(bindirs)"
	@echo "bins     found: $(bins)"
	@echo "testdirs found: $(testdirs)"
	@echo "================================================================================"

build:
	@echo "********************************************************************************"
	@echo Building $(bindirs)
	@echo Variables:
	@echo buildVersionVar: $(buildVersionVar)
	@echo version: $(version)
	@echo buildCommitVar: $(buildCommitVar)
	@echo commit: $(commit)
	@echo "********************************************************************************"
	for i in $(bindirs);
	do
		echo "................................................................................"
		echo "--> Now building: $$i"
		echo "................................................................................"
		go build -v -ldflags '-X $(buildVersionVar)=$(version) -X $(buildCommitVar)=$(commit)' $$i;
	done;

install:
	mkdir -p $(installdir)
	cp $(bins) $(readme) $(templates) $(config) $(installdir)

test_new:
	$(foreach dir, $(testdirs), go test -v -count=1 $(dir) || exit $$?;)

test:
	@echo "********************************************************************************"
	@echo Testing
	@echo "********************************************************************************"
	go test -v -count=1 ./...

endly_linux_$(endly_version).tar.gz:
	curl -L -O https://github.com/viant/endly/releases/download/v$(endly_version)/endly_linux_$(endly_version).tar.gz

test_e2e/endly:
	tar -C test_e2e/ -xzf endly_linux_$(endly_version).tar.gz

get_endly: endly_linux_$(endly_version).tar.gz test_e2e/endly

test_endly:
	cd test_e2e
	./endly

clean:
	rm $(bins)
	rm -rf $(installdir)
	rm endly_linux_$(endly_version).tar.gz
	rm test_e2e/endly
