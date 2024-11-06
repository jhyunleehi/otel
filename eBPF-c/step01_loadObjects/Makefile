TARGET = counter
ARCH = $(shell uname -m | sed 's/x86_64/x86/' | sed 's/aarch64/arm64/')

BPF_OBJ = ${TARGET:=.bpf.o}
USER_C = ${TARGET:=.c}
USER_SKEL = ${TARGET:=.skel.h}

export GOPACKAGE := main

all: $(TARGET) $(BPF_OBJ)
.PHONY: all 

$(TARGET): $(USER_C) $(USER_SKEL) 
	echo  go build...	
	go get github.com/cilium/ebpf/cmd/bpf2go
	# go run github.com/cilium/ebpf/cmd/bpf2go -type event ${TARGET} ${TARGET}.c -- -I../headers
	go generate
	go build  

%.bpf.o: %.bpf.c vmlinux.h
	clang \
	    -target bpf \
        -D __TARGET_ARCH_$(ARCH) \
	    -Wall \
	    -O2 -g -o $@ -c $<
	llvm-strip -g $@

$(USER_SKEL): $(BPF_OBJ)
	bpftool gen skeleton $< > $@

vmlinux.h:
	bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h

clean:
	- rm $(BPF_OBJ)
	- rm $(TARGET)	

