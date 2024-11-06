
### dependencies

* Linux kernel version 5.7 or later, for bpf_link support
* LLVM 11 or later 1 (clang and llvm-strip)
* libbpf headers 2
* Linux kernel headers 3
* Go compiler version supported by ebpf-go's Go module

### install package and heade file 

* clang --version 
* apt install libbpf-dev
* apt install linux-headers-amd64
* ln -sf /usr/include/asm-generic/ /usr/include/asm

### run vscode sudo 

```
root@Good:~# cat .profile 
export PATH=$PATH:/usr/local/go/bin

root@Good:~/go/src/ebpf# cat code.sh
sudo code --no-sandbox --user-data-dir=/root/.config/vscode_data
```

### vscode debug as root
https://github.com/golang/vscode-go/blob/master/docs/debugging.md#debugging-programs-and-tests-as-root

#### 1. Debug a program as root
* task.json
```json
{
    ...
    "tasks": [
        {
            "label": "go: build (debug)",
            "type": "shell",
            "command": "go",
            "args": [
                "build",
                "-gcflags=all=-N -l",
                "-o",
                "${fileDirname}/__debug_bin"
            ],
            "options": {
                "cwd": "${fileDirname}"
            },
            ...
        }
    ]
}
```
* launcher.json
```json    
    
        {
            "name": "Launch Package as root",
            "type": "go",
            "request": "launch",
            "mode": "exec",
            "asRoot": true,
            "console": "integratedTerminal",
            "program": "${fileDirname}/__debug_bin",
            "preLaunchTask": "go: build (debug)",
        }
    
```


#### 2. Debug a package test as root

* task.json
```json
    ...
    "tasks": [
        {
            "label": "go test (debug)",
            "type": "shell",
            "command": "go",
            "args": [
                "test",
                "-c",
                "-o",
                "${fileDirname}/__debug_bin"
            ],
            "options": {
                "cwd": "${fileDirname}",
            },
            ...
        }
    ]
```

* launch.json
```json
    ...
    "configurations": [        
        {
            "name": "Debug Package Test as root",
            "type": "go",
            "request": "launch",
            "mode": "exec",
            "asRoot": true,
            "program": "${fileDirname}/__debug_bin",
            "cwd": "${fileDirname}",
            "console": "integratedTerminal",
            "preLaunchTask": "go test (debug)"
        }
    ]
```





### cross compile 
bpf2go가 두 개의 파일 세트
*_bpfel.o*_bpfel.goamd64, arm64, riscv64 및 loong64와 같은 리틀 엔디안 아키텍처의 경우
*_bpfeb.o*_bpfeb.gos390(x), mips 및 sparc와 같은 빅엔디안 아키텍처의 경우


#### make error 
* libbpf libary compile and install 

```sh
root@Good:~/go/src/ebpf-go/step01# make
clang \
    -target bpf \
        -D __TARGET_ARCH_x86 \
    -Wall \
    -O2 -g -o counter.bpf.o -c counter.bpf.c
llvm-strip -g counter.bpf.o
bpftool gen skeleton counter.bpf.o > counter.skel.h
gcc -Wall -o counter counter.c -L../libbpf/src -l:libbpf.a -lelf -lz
In file included from counter.c:5:
counter.c:12:13: error: expected ‘=’, ‘,’, ‘;’, ‘asm’ or ‘__attribute__’ before ‘#pragma’
   12 | } pkt_count SEC(".maps");
      |             ^~~
counter.c:12:13: error: expected identifier or ‘(’ before ‘#pragma’
   12 | } pkt_count SEC(".maps");
      |             ^~~
counter.c:15:1: error: expected identifier or ‘(’ before ‘#pragma’
   15 | SEC("xdp")
      | ^~~
counter.c:26:18: error: expected ‘=’, ‘,’, ‘;’, ‘asm’ or ‘__attribute__’ before ‘#pragma’
   26 | char __license[] SEC("license") = "Dual MIT/GPL";
      |                  ^~~
counter.c:26:18: error: expected identifier or ‘(’ before ‘#pragma’
   26 | char __license[] SEC("license") = "Dual MIT/GPL";
      |                  ^~~
make: *** [Makefile:12: counter] 오류 1
```
==>

* reinstall libbpf
```
$ git clone --recurse-submodules https://github.com/lizrice/learning-ebpf
$ cd learning-ebpf/libbpf/src
$ sudo make install
```


#### llvm-readelf
```
root@Good:~/go/src/ebpf-go/step03# readelf -SW s3.bpf.o
There are 11 section headers, starting at offset 0x3b0:

Section Headers:
  [Nr] Name              Type            Address          Off    Size   ES Flg Lk Inf Al
  [ 0]                   NULL            0000000000000000 000000 000000 00      0   0  0
  [ 1] .strtab           STRTAB          0000000000000000 000352 000057 00      0   0  1
  [ 2] .text             PROGBITS        0000000000000000 000040 000000 00  AX  0   0  4
  [ 3] socket            PROGBITS        0000000000000000 000040 000010 00  AX  0   0  8
  [ 4] .maps             PROGBITS        0000000000000000 000050 000020 00  WA  0   0  8
  [ 5] .BTF              PROGBITS        0000000000000000 000070 0001fb 00      0   0  4
  [ 6] .rel.BTF          REL             0000000000000000 000320 000010 10   I 10   5  8
  [ 7] .BTF.ext          PROGBITS        0000000000000000 00026c 000050 00      0   0  4
  [ 8] .rel.BTF.ext      REL             0000000000000000 000330 000020 10   I 10   7  8
  [ 9] .llvm_addrsig     LOOS+0xfff4c03  0000000000000000 000350 000002 00   E  0   0  1
  [10] .symtab           SYMTAB          0000000000000000 0002c0 000060 18      1   2  8
Key to Flags:
  W (write), A (alloc), X (execute), M (merge), S (strings), I (info),
  L (link order), O (extra OS processing required), G (group), T (TLS),
  C (compressed), x (unknown), o (OS specific), E (exclude),
  D (mbind), p (processor specific)  
```

```
# readelf --section-details --headers .output/opensnoop.bpf.o -W
```

#### llvm-objdump 
```
root@Good:~/go/src/ebpf-go/step03# llvm-objdump -SD  s3.bpf.o 

s3.bpf.o:       file format elf64-bpf

```

## bpftool 
* bpftool prog show id 540
* bpftool prog show name hello
* bpftool prog show tag d35b94b4c0c10efb
* bpftool prog show pinned /sys/fs/bpf/hello
* bpftool prog dump xlated name hello
* bpftool prog show id 487 --pretty
* bpftool prog list
* bpftool prog list name hello
* bpftool prog load hello.bpf.o  /sys/fs/bpf/hello
* bpftool prog load hello-func.bpf.o /sys/fs/bpf/hello
* bpftool prog load hello.bpf.o /sys/fs/bpf/hello
* bpftool prog show id 487 --pretty
* bpftool prog show name hello
* bpftool prog dump xlated name hello
* bpftool prog dump xlated name hello
* bpftool prog trace log
* bpftool prog trace log
* bpftool prog show name hello
* bpftool map list
* bpftool map show id $MAP_ID
* bpftool map dump id $MAP_ID
* bpftool map show id $MAP_ID 
* bpftool map lookup id $MAP_ID key 100 0 0 0 0 0 0 0
* bpftool map lookup id $MAP_ID key 105  0 0 0 0 0 0 0
* bpftool map lookup id $MAP_ID key 0x64 0 0 0 0 0 0 0
* bpftool map lookup id $MAP_ID key hex 64 0 0 0 0 0 0 0
* bpftool map update  id $MAP_ID key 255 0 0 0 0 0 0 0 value 255 0 0 0 0 0 0 0
* bpftool map dump name hello.rodata
* bpftool gen skeleton hello-buffer-config.bpf.o > hello-buffer-config.skel.h
* bpftool net list
* bpftool net detach xdp  dev lo
* bpftool net attach xdp name hello dev lo



## go mod env 

Go 환경 설정에서  

root 디렉토리에서  go.mod

```sh
module ebpf-go
go 1.22.0
```

각 개별 디렉토리에서 main 패키지

ebpf-go/fenty  디렉토리에서 main 패키지 설정하려면 

go.mod 파일 설정 필요 

```sh
module ebpf-go/fentry

go 1.22.0

require github.com/cilium/ebpf v0.13.2

require (
golang.org/x/exp v0.0.0-20230224173230-c95f2b4c22f2 // indirect
golang.org/x/sys v0.15.0 // indirect
)

```

#### fentry

eBPF (Extended Berkeley Packet Filter)는 리눅스 커널 내에서 안전하게 실행되는 프로그램을 삽입하여 네트워크 처리, 보안, 추적 및 다른 작업을 위한 기능을 확장하는 기술입니다. eBPF는 프로그램을 컴파일하여 커널에 주입하고, 이를 이용하여 커널 내의 다양한 동작을 확장할 수 있습니다.

여기서 "fentry"는 eBPF의 한 유형의 프로그램이며, "fentry"는 "function entry"를 의미합니다. 즉, 이 프로그램은 함수가 호출될 때 발생하는 이벤트를 가로채고 처리하는 용도로 사용됩니다.

"fentry" 프로그램은 주로 함수 호출 시에 발생하는 이벤트를 모니터링하거나 트레이싱할 때 사용됩니다. 예를 들어, 네트워크 스택에서 특정 함수가 호출될 때마다 호출되는 시간을 측정하거나, 특정 시스템 호출이 얼마나 자주 발생하는지 확인할 때 사용될 수 있습니다.

eBPF의 "fentry" 프로그램은 주로 프로그램 실행 중 함수 호출의 흐름을 추적하고, 이를 통해 성능 프로파일링, 디버깅, 보안 검사 등 다양한 용도로 활용됩니다.


#### ebpf.o object file read 
```go
// Do not access this directly.
//
//go:embed fentry_bpfel.o
var _FentryBytes []byte
```

ChatGPT
The code snippet you've provided contains a directive comment //go:embed followed by a variable declaration. This syntax is used in Go 1.16 or later to embed files directly into the Go binary at compile time.

Here's what this snippet does:

//go:embed fentry_bpfel.o: This is a directive comment telling the Go compiler to embed the file named fentry_bpfel.o into the Go binary.
var _FentryBytes []byte: This declares a variable _FentryBytes of type []byte. The leading underscore _ is used to indicate that this variable is not meant to be directly accessed by the code.
So, essentially, this code embeds the contents of the file fentry_bpfel.o into the Go binary as a byte slice named _FentryBytes, making it accessible within the Go program. However, since the variable name starts with an underscore, it is not intended to be accessed directly by other parts of the code. Instead, it is typically used by other parts of the program that need access to the embedded file's contents.


## example  BPF_KPROBE
```c
/* SPDX-License-Identifier: (LGPL-2.1 OR BSD-2-Clause) */
/* Copyright (c) 2021 Hengqi Chen */
#include <vmlinux.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include "bits.bpf.h"
#include "fsdist.h"

#define MAX_ENTRIES	10240

const volatile pid_t target_pid = 0;
const volatile bool in_ms = false;

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, MAX_ENTRIES);
	__type(key, __u32);
	__type(value, __u64);
} starts SEC(".maps");

struct hist hists[F_MAX_OP] = {};

static int probe_entry()
{
	__u64 pid_tgid = bpf_get_current_pid_tgid();
	__u32 pid = pid_tgid >> 32;
	__u32 tid = (__u32)pid_tgid;
	__u64 ts;

	if (target_pid && target_pid != pid)
		return 0;

	ts = bpf_ktime_get_ns();
	bpf_map_update_elem(&starts, &tid, &ts, BPF_ANY);
	return 0;
}

static int probe_return(enum fs_file_op op)
{
	__u32 tid = (__u32)bpf_get_current_pid_tgid();
	__u64 ts = bpf_ktime_get_ns();
	__u64 *tsp, slot;
	__s64 delta;

	tsp = bpf_map_lookup_elem(&starts, &tid);
	if (!tsp)
		return 0;

	if (op >= F_MAX_OP)
		goto cleanup;

	delta = (__s64)(ts - *tsp);
	if (delta < 0)
		goto cleanup;

	if (in_ms)
		delta /= 1000000;
	else
		delta /= 1000;

	slot = log2l(delta);
	if (slot >= MAX_SLOTS)
		slot = MAX_SLOTS - 1;
	__sync_fetch_and_add(&hists[op].slots[slot], 1);

cleanup:
	bpf_map_delete_elem(&starts, &tid);
	return 0;
}

SEC("kprobe/dummy_file_read")
int BPF_KPROBE(file_read_entry){return probe_entry();}

SEC("kretprobe/dummy_file_read")
int BPF_KRETPROBE(file_read_exit){return probe_return(F_READ);}

SEC("kprobe/dummy_file_write")
int BPF_KPROBE(file_write_entry){return probe_entry();}

SEC("kretprobe/dummy_file_write")
int BPF_KRETPROBE(file_write_exit){return probe_return(F_WRITE);}

SEC("kprobe/dummy_file_open")
int BPF_KPROBE(file_open_entry){return probe_entry();}

SEC("kretprobe/dummy_file_open")
int BPF_KRETPROBE(file_open_exit){return probe_return(F_OPEN);}

SEC("kprobe/dummy_file_sync")
int BPF_KPROBE(file_sync_entry){return probe_entry();}

SEC("kretprobe/dummy_file_sync")
int BPF_KRETPROBE(file_sync_exit){return probe_return(F_FSYNC);}

SEC("kprobe/dummy_getattr")
int BPF_KPROBE(getattr_entry){return probe_entry();}

SEC("kretprobe/dummy_getattr")
int BPF_KRETPROBE(getattr_exit){return probe_return(F_GETATTR);}

SEC("fentry/dummy_file_read")
int BPF_PROG(file_read_fentry){return probe_entry();}

SEC("fexit/dummy_file_read")
int BPF_PROG(file_read_fexit){return probe_return(F_READ);}

SEC("fentry/dummy_file_write")
int BPF_PROG(file_write_fentry){return probe_entry();}

SEC("fexit/dummy_file_write")
int BPF_PROG(file_write_fexit){return probe_return(F_WRITE);}

SEC("fentry/dummy_file_open")
int BPF_PROG(file_open_fentry){return probe_entry();}

SEC("fexit/dummy_file_open")
int BPF_PROG(file_open_fexit){return probe_return(F_OPEN);}

SEC("fentry/dummy_file_sync")
int BPF_PROG(file_sync_fentry){return probe_entry();}

SEC("fexit/dummy_file_sync")
int BPF_PROG(file_sync_fexit){return probe_return(F_FSYNC);}

SEC("fentry/dummy_getattr")
int BPF_PROG(getattr_fentry){return probe_entry();}

SEC("fexit/dummy_getattr")
int BPF_PROG(getattr_fexit){return probe_return(F_GETATTR);}

char LICENSE[] SEC("license") = "Dual BSD/GPL";

```


## BPF_MAP_TYPE

```c
enum bpf_map_type {
	BPF_MAP_TYPE_UNSPEC = 0,
	BPF_MAP_TYPE_HASH = 1,
	BPF_MAP_TYPE_ARRAY = 2,
	BPF_MAP_TYPE_PROG_ARRAY = 3,
	BPF_MAP_TYPE_PERF_EVENT_ARRAY = 4,
	BPF_MAP_TYPE_PERCPU_HASH = 5,
	BPF_MAP_TYPE_PERCPU_ARRAY = 6,
	BPF_MAP_TYPE_STACK_TRACE = 7,
	BPF_MAP_TYPE_CGROUP_ARRAY = 8,
	BPF_MAP_TYPE_LRU_HASH = 9,
	BPF_MAP_TYPE_LRU_PERCPU_HASH = 10,
	BPF_MAP_TYPE_LPM_TRIE = 11,
	BPF_MAP_TYPE_ARRAY_OF_MAPS = 12,
	BPF_MAP_TYPE_HASH_OF_MAPS = 13,
	BPF_MAP_TYPE_DEVMAP = 14,
	BPF_MAP_TYPE_SOCKMAP = 15,
	BPF_MAP_TYPE_CPUMAP = 16,
	BPF_MAP_TYPE_XSKMAP = 17,
	BPF_MAP_TYPE_SOCKHASH = 18,
	BPF_MAP_TYPE_CGROUP_STORAGE_DEPRECATED = 19,
	BPF_MAP_TYPE_CGROUP_STORAGE = 19,
	BPF_MAP_TYPE_REUSEPORT_SOCKARRAY = 20,
	BPF_MAP_TYPE_PERCPU_CGROUP_STORAGE = 21,
	BPF_MAP_TYPE_QUEUE = 22,
	BPF_MAP_TYPE_STACK = 23,
	BPF_MAP_TYPE_SK_STORAGE = 24,
	BPF_MAP_TYPE_DEVMAP_HASH = 25,
	BPF_MAP_TYPE_STRUCT_OPS = 26,
	BPF_MAP_TYPE_RINGBUF = 27,
	BPF_MAP_TYPE_INODE_STORAGE = 28,
	BPF_MAP_TYPE_TASK_STORAGE = 29,
	BPF_MAP_TYPE_BLOOM_FILTER = 30,
	BPF_MAP_TYPE_USER_RINGBUF = 31,
	BPF_MAP_TYPE_CGRP_STORAGE = 32,
};

enum bpf_prog_type {
	BPF_PROG_TYPE_UNSPEC = 0,
	BPF_PROG_TYPE_SOCKET_FILTER = 1,
	BPF_PROG_TYPE_KPROBE = 2,
	BPF_PROG_TYPE_SCHED_CLS = 3,
	BPF_PROG_TYPE_SCHED_ACT = 4,
	BPF_PROG_TYPE_TRACEPOINT = 5,
	BPF_PROG_TYPE_XDP = 6,
	BPF_PROG_TYPE_PERF_EVENT = 7,
	BPF_PROG_TYPE_CGROUP_SKB = 8,
	BPF_PROG_TYPE_CGROUP_SOCK = 9,
	BPF_PROG_TYPE_LWT_IN = 10,
	BPF_PROG_TYPE_LWT_OUT = 11,
	BPF_PROG_TYPE_LWT_XMIT = 12,
	BPF_PROG_TYPE_SOCK_OPS = 13,
	BPF_PROG_TYPE_SK_SKB = 14,
	BPF_PROG_TYPE_CGROUP_DEVICE = 15,
	BPF_PROG_TYPE_SK_MSG = 16,
	BPF_PROG_TYPE_RAW_TRACEPOINT = 17,
	BPF_PROG_TYPE_CGROUP_SOCK_ADDR = 18,
	BPF_PROG_TYPE_LWT_SEG6LOCAL = 19,
	BPF_PROG_TYPE_LIRC_MODE2 = 20,
	BPF_PROG_TYPE_SK_REUSEPORT = 21,
	BPF_PROG_TYPE_FLOW_DISSECTOR = 22,
	BPF_PROG_TYPE_CGROUP_SYSCTL = 23,
	BPF_PROG_TYPE_RAW_TRACEPOINT_WRITABLE = 24,
	BPF_PROG_TYPE_CGROUP_SOCKOPT = 25,
	BPF_PROG_TYPE_TRACING = 26,
	BPF_PROG_TYPE_STRUCT_OPS = 27,
	BPF_PROG_TYPE_EXT = 28,
	BPF_PROG_TYPE_LSM = 29,
	BPF_PROG_TYPE_SK_LOOKUP = 30,
	BPF_PROG_TYPE_SYSCALL = 31,
	BPF_PROG_TYPE_NETFILTER = 32,
};
```

root@Good:~/go/src/ebpf-go/step06-filelife# go run github.com/cilium/ebpf/cmd/bpf2go -h
Usage: bpf2go [options] <ident> <source file> [-- <C flags>]

ident is used as the stem of all generated Go types and functions, and
must be a valid Go identifier.

source is a single C file that is compiled using the specified compiler
(usually some version of clang).

You can pass options to the compiler by appending them after a '--' argument
or by supplying -cflags. Flags passed as arguments take precedence
over flags passed via -cflags. Additionally, the program expands quotation
marks in -cflags. This means that -cflags 'foo "bar baz"' is passed to the
compiler as two arguments "foo" and "bar baz".

The program expects GOPACKAGE to be set in the environment, and should be invoked
via go generate. The generated files are written to the current directory.

Some options take defaults from the environment. Variable name is mentioned
next to the respective option.

Options:

  -cc binary
        binary used to compile C to BPF ($BPF2GO_CC) (default "clang")
  -cflags string
        flags passed to the compiler, may contain quoted arguments ($BPF2GO_CFLAGS)
  -go-package string
        package for output go file (default as ENV GOPACKAGE)
  -makebase directory
        write make compatible depinfo files relative to directory ($BPF2GO_MAKEBASE)
  -no-global-types
        Skip generating types for map keys and values, etc.
  -no-strip
        disable stripping of DWARF
  -output-dir string
        target directory of generated files (defaults to current directory)
  -output-stem string
        alternative stem for names of generated files (defaults to ident)
  -strip binary
        binary used to strip DWARF from compiled BPF ($BPF2GO_STRIP)
  -tags value
        Comma-separated list of Go build tags to include in generated files
  -target string
        clang target(s) to compile for (comma separated) (default "bpfel,bpfeb")
  -type Name
        Name of a type to generate a Go declaration for, may be repeated

Supported targets:
        bpf
        bpfel
        bpfeb
        386
        amd64
        arm
        arm64
        loong64
        mips
        ppc64
        ppc64le
        riscv64
        s390x
root@Good:~/go/src/ebpf-go/step06-filelife# 




https://docs.kernel.org/bpf/libbpf/index.html


https://docs.kernel.org/bpf/libbpf/libbpf_overview.html

https://libbpf.readthedocs.io/en/latest/api.html

https://docs.kernel.org/bpf/libbpf/program_types.html


### bpf_map_get_next_key

```c
#include <stdint.h>
#include <stddef.h>
#include <linux/bpf.h>
#include <linux/in.h>
#include <bpf/bpf_helpers.h>

// Define the eBPF map
struct bpf_map_def SEC("maps") my_map = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(uint32_t),  // Assuming IPv4 addresses
    .value_size = sizeof(uint64_t), // Assuming some value associated with each IP
    .max_entries = 1024,
};

SEC("next_key_example")
int next_key_example(struct __sk_buff *skb)
{
    uint32_t key;
    uint64_t *value;

    // Initialize iteration with an empty key
    key = 0;

    // Iterate through all keys in the map
    while (bpf_map_get_next_key(&my_map, &key, &key) == 0) {
        // Lookup the value associated with the current key
        value = bpf_map_lookup_elem(&my_map, &key);
        if (value) {
            // Do something with the key and value
            // For example, print them
            bpf_printk("Key: %u, Value: %lu\n", key, *value);
        }
    }

    return 0;
}

char _license[] SEC("license") = "GPL";
```


### map []char ???

```c
struct counter {
	__u64 last_sector;
	__u64 bytes;
	__u32 sequential;
	__u32 random;
};

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 64);
	__type(key, u32);
	__type(value, struct counter);
} counters SEC(".maps");
```


```c
struct piddata {
	char comm[TASK_COMM_LEN];
	u32 pid;
};

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, MAX_ENTRIES);
	__type(key, struct request *);
	__type(value, struct piddata);
} infobyreq SEC(".maps");
```

```c
struct event {
	pid_t pid;
	pid_t ppid;
	uid_t uid;
	int retval;
	int args_count;
	unsigned int args_size;
	char comm[TASK_COMM_LEN];
	char args[FULL_MAX_ARGS_ARR];
};

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 10240);
	__type(key, pid_t);
	__type(value, struct event);
} execs SEC(".maps");
```


```c
struct file_stat {
	__u64 reads;
	__u64 read_bytes;
	__u64 writes;
	__u64 write_bytes;
	__u32 pid;
	__u32 tid;
	char filename[PATH_MAX];
	char comm[TASK_COMM_LEN];
	char type;
};


struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, MAX_ENTRIES);
	__type(key, struct file_id);
	__type(value, struct file_stat);
} entries SEC(".maps");

```




## tracepoint/syscall event 
* tracepoint paramenter 
```c
SEC("tracepoint/syscalls/sys_enter_mount")
int mount_entry(struct trace_event_raw_sys_enter *ctx) {
	...
}
```	

* /sys/kernel/debug/tracing/events/syscalls/sys_enter_mount/format 
```sh
root@Good:/sys/kernel/debug/tracing/events/syscalls/sys_enter_mount# cat format 
name: sys_enter_mount
ID: 834
format:
	field:unsigned short common_type;	offset:0;	size:2;	signed:0;
	field:unsigned char common_flags;	offset:2;	size:1;	signed:0;
	field:unsigned char common_preempt_count;	offset:3;	size:1;	signed:0;
	field:int common_pid;	offset:4;	size:4;	signed:1;

	field:int __syscall_nr;	offset:8;	size:4;	signed:1;
	field:char * dev_name;	offset:16;	size:8;	signed:0;
	field:char * dir_name;	offset:24;	size:8;	signed:0;
	field:char * type;	offset:32;	size:8;	signed:0;
	field:unsigned long flags;	offset:40;	size:8;	signed:0;
	field:void * data;	offset:48;	size:8;	signed:0;

print fmt: "dev_name: 0x%08lx, dir_name: 0x%08lx, type: 0x%08lx, flags: 0x%08lx, data: 0x%08lx",
 ((unsigned long)(REC->dev_name)), 
 ((unsigned long)(REC->dir_name)), 
 ((unsigned long)(REC->type)), 
 ((unsigned long)(REC->flags)), 
 ((unsigned long)(REC->data))
```
* data type in vmlinux.h
```sh
root@Good:~/go/src/ebpf-go/ex03-file# grep  trace_event_raw_sys_enter  vmlinux.h 
struct trace_event_raw_sys_enter {
root@Good:~/go/src/ebpf-go/ex03-file

struct trace_event_raw_sys_enter {
	struct trace_entry ent;
	long int id;
	long unsigned int args[6];
	char __data[0];
};

```