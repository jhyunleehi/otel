TARGET = s3
ARCH = $(shell uname -m | sed 's/x86_64/x86/' | sed 's/aarch64/arm64/')

BPF_OBJ = ${TARGET:=.bpf.o}
USER_SKEL = ${TARGET:=.skel.h}

all: $(TARGET) 
.PHONY: all 

$(TARGET): $(USER_C) $(USER_SKEL) 
	echo  go build...
	go build main.go

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
	- rm main

find-map: find-map.c
	gcc -Wall -o find-map find-map.c -L../libbpf/src -l:libbpf.a -lelf -lz
