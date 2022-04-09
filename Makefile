.PHONY: clean build test test_new list all

.ONESHELL:
#SHELL =
#.SHELLFLAGS =

# Inject into binary via linker:
# ...in github actions comes from make -e version=git_ref
version=$(shell cat VERSION)
commit=$(shell git show --format=format:%H HEAD)
buildVersionVar=github.com/pja237/goslmailer/internal/version.buildVersion
buildCommitVar=github.com/pja237/goslmailer/internal/version.buildCommit

# various directories
bindirs=$(wildcard ./cmd/*)
installdir=build/goslmailer-$(version)

# list of files to include in build
bins=$(notdir $(bindirs))
readme=README.md
templates=templates/adaptive_card_template.json templates/telegramTemplate.html
config=cmd/goslmailer/goslmailer.conf.annotated_example cmd/gobler/gobler.conf

# can be replaced with go test ./... construct
testdirs=$(sort $(dir $(shell find ./ -name *_test.go)))

all: list test build install

list:
	@echo "================================================================================"
	@echo "bindirs  found: $(bindirs)"
	@echo "bins     found: $(bins)"
	@echo "testdirs found: $(testdirs)"
	@echo "================================================================================"

build:
	@echo "********************************************************************************"
	@echo Building $(bindirs)
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

clean:
	rm $(bins)
	rm -rf $(installdir)
