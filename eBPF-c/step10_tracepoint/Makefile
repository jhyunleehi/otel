TARGET = tracepoint
ARCH = $(shell uname -m | sed 's/x86_64/x86/' | sed 's/aarch64/arm64/')

BPF_OBJ = ${TARGET:=.o}
USER_SKEL = ${TARGET:=.skel.h}

export GOPACKAGE := main

all: $(TARGET) 
.PHONY: all 

#$(TARGET): $(USER_C) $(USER_SKEL) 
$(TARGET): $(USER_SKEL) 
	echo  go build...	
	go get github.com/cilium/ebpf/cmd/bpf2go	
	go run github.com/cilium/ebpf/cmd/bpf2go  bpf ${TARGET}.c -- -I../headers
	# go run github.com/cilium/ebpf/cmd/bpf2go -target amd64  -type rtt_event bpf ${TARGET}.c -- -I../headers
	# go run github.com/cilium/ebpf/cmd/bpf2go -target amd64 bpf ${TARGET}.c -- -I../headers
    # go run github.com/cilium/ebpf/cmd/bpf2go  bpf tcprtt_sockops.c -- -I../headers
	go generate
	go build  

%.o: %.c vmlinux.h
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
	- rm *.o

find-map: find-map.c
	gcc -Wall -o find-map find-map.c -L../libbpf/src -l:libbpf.a -lelf -lz
