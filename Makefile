.PHONY: clean build test test_new list all

.ONESHELL:
#SHELL = 
#.SHELLFLAGS =

bindirs=$(wildcard ./cmd/*)
bins=$(notdir $(bindirs))
# can be replaced with go test ./... construct
#testdirs=$(sort $(dir $(shell find ./ -name *_test.go)))

all: list test build

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
		go build -v $$i;
	done;

test_new:
	$(foreach dir, $(testdirs), go test -v -count=1 $(dir) || exit $$?;)

test:
	@echo "********************************************************************************"
	@echo Testing
	@echo "********************************************************************************"
	go test -v -count=1 ./...


clean:
	rm ./goslmailer
