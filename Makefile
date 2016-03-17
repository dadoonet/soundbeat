BEATNAME=soundbeat
BEAT_DIR=github.com/dadoonet
SYSTEM_TESTS=false
TEST_ENVIRONMENT=false
ES_BEATS=./vendor/github.com/elastic/beats
GOPACKAGES=$(shell glide novendor)
PREFIX?=.

# Path to the libbeat Makefile
-include $(ES_BEATS)/libbeat/scripts/Makefile

.PHONY: init
init:
	glide update  --no-recursive
	make update
	git init
	git add .

.PHONY: update-deps
update-deps:
	glide update  --no-recursive

# This is called by the beats packer before building starts
.PHONY: before-build
before-build:
