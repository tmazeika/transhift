ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

build:
	cd lib/hash_2_cga && cargo build --release
	cp lib/hash_2_cga/target/release/libhash_2_cga.so lib/
	go build -ldflags="-r $(ROOT_DIR)lib" cga.go

run: build
	./cga
