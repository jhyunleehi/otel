# Error list


### bpf2go compile error   
```
$(TARGET): $(USER_SKEL) 
	echo  go build...	
	go get github.com/cilium/ebpf/cmd/bpf2go
	go run github.com/cilium/ebpf/cmd/bpf2go  bpf ${TARGET}.c -- -I../headers
	go generate
	go build  
```

```sh
root@Good:~/go/src/ebpf-go/step09# make
clang \
    -target bpf \
        -D __TARGET_ARCH_x86 \
    -Wall \
    -O2 -g -o tcprtt.o -c tcprtt.c
llvm-strip -g tcprtt.o
bpftool gen skeleton tcprtt.o > tcprtt.skel.h
echo  go build...
go build...
go get github.com/cilium/ebpf/cmd/bpf2go
go run github.com/cilium/ebpf/cmd/bpf2go  bpf tcprtt.c -- -I../headers
/root/go/src/ebpf-go/step09/tcprtt.c:96:5: error: The eBPF is using target specific macros, please provide -target that is not bpf, bpfel or bpfeb
int BPF_KPROBE(vfs_open, struct path *path, struct file *file)
    ^
/root/go/src/ebpf-go/step09/bpf_tracing.h:461:20: note: expanded from macro 'BPF_KPROBE'
        return ____##name(___bpf_kprobe_args(args));                        \
                          ^
/root/go/src/ebpf-go/step09/bpf_tracing.h:441:2: note: expanded from macro '___bpf_kprobe_args'
        ___bpf_apply(___bpf_kprobe_args, ___bpf_narg(args))(args)
        ^
/root/go/src/ebpf-go/step09/bpf_helpers.h:157:29: note: expanded from macro '___bpf_apply'
#define ___bpf_apply(fn, n) ___bpf_concat(fn, n)
                            ^
note: (skipping 3 expansions in backtrace; use -fmacro-backtrace-limit=0 to see all)
/root/go/src/ebpf-go/step09/bpf_tracing.h:431:33: note: expanded from macro '___bpf_kprobe_args1'
        ___bpf_kprobe_args0(), (void *)PT_REGS_PARM1(ctx)
                                       ^
/root/go/src/ebpf-go/step09/bpf_tracing.h:341:29: note: expanded from macro 'PT_REGS_PARM1'
#define PT_REGS_PARM1(x) ({ _Pragma(__BPF_TARGET_MISSING); 0l; })
                            ^
<scratch space>:51:6: note: expanded from here
 GCC error "The eBPF is using target specific macros, please provide -target that is not bpf, bpfel or bpfeb"
     ^
/root/go/src/ebpf-go/step09/tcprtt.c:96:5: error: The eBPF is using target specific macros, please provide -target that is not bpf, bpfel or bpfeb
/root/go/src/ebpf-go/step09/bpf_tracing.h:461:20: note: expanded from macro 'BPF_KPROBE'
        return ____##name(___bpf_kprobe_args(args));                        \
                          ^
/root/go/src/ebpf-go/step09/bpf_tracing.h:441:2: note: expanded from macro '___bpf_kprobe_args'
        ___bpf_apply(___bpf_kprobe_args, ___bpf_narg(args))(args)
        ^
/root/go/src/ebpf-go/step09/bpf_helpers.h:157:29: note: expanded from macro '___bpf_apply'
#define ___bpf_apply(fn, n) ___bpf_concat(fn, n)
                            ^
note: (skipping 2 expansions in backtrace; use -fmacro-backtrace-limit=0 to see all)
/root/go/src/ebpf-go/step09/bpf_tracing.h:433:37: note: expanded from macro '___bpf_kprobe_args2'
        ___bpf_kprobe_args1(args), (void *)PT_REGS_PARM2(ctx)
                                           ^
/root/go/src/ebpf-go/step09/bpf_tracing.h:342:29: note: expanded from macro 'PT_REGS_PARM2'
#define PT_REGS_PARM2(x) ({ _Pragma(__BPF_TARGET_MISSING); 0l; })
                            ^
<scratch space>:53:6: note: expanded from here
 GCC error "The eBPF is using target specific macros, please provide -target that is not bpf, bpfel or bpfeb"
     ^
2 errors generated.
Error: can't execute clang: exit status 1
exit status 1
make: *** [Makefile:16: tcprtt] 오류 1
root@Good:~/go/src/ebpf-go/step09# 
```

#### fix it 
* only support -target amd64 
```Makefile 
$(TARGET): $(USER_SKEL) 
	echo  go build...	
	go get github.com/cilium/ebpf/cmd/bpf2go
    # go run github.com/cilium/ebpf/cmd/bpf2go -target bpfel  -type event bpf ${TARGET}.c -- -I../headers
    # go run github.com/cilium/ebpf/cmd/bpf2go -target bpfeb  -type event bpf ${TARGET}.c -- -I../headers
	go run github.com/cilium/ebpf/cmd/bpf2go -target amd64  -type event bpf ${TARGET}.c -- -I../headers
	go generate
	go build  
```

##  *btf.Pointer: not supported 
```sh
root@Good:~/go/src/ebpf-go/ex02-mount# go run github.com/cilium/ebpf/cmd/bpf2go -target amd64  -type event bpf mountsnoop.c -- -I../headers
Compiled /root/go/src/ebpf-go/ex02-mount/bpf_x86_bpfel.o
Stripped /root/go/src/ebpf-go/ex02-mount/bpf_x86_bpfel.o
Error: can't write /root/go/src/ebpf-go/ex02-mount/bpf_x86_bpfel.go: can't generate types: template: common:17:4: executing "common" at <$.TypeDeclaration>: error calling TypeDeclaration: Struct:"arg": field 2: type *btf.Pointer: not supported
exit status 1
root@Good:~/go/src/
```
==> remove 
* remove const char *  
```c
struct arg {
	__u64 ts;	
	__u64 flags;
	const char *src;  <<----- not support it 
	const char *dest;
	const char *fs;
	const char *data;
	enum op op;
};
```

```c
struct arg {
	__u64 ts;	
	__u64 flags;	
  __u64 src;    <<==== const char *  ---> const __u64
  __u64 dest;
  __u64 fs;
  __u64 data;
	enum op op;
};


static int probe_entry(const char *src, const char *dest, const char *fs,
                       __u64 flags, const char *data, enum op op) {
  __u64 pid_tgid = bpf_get_current_pid_tgid();
  __u32 pid = pid_tgid >> 32;
  __u32 tid = (__u32)pid_tgid;
  struct arg arg = {};

  if (target_pid && target_pid != pid)
    return 0;

  arg.ts = bpf_ktime_get_ns();
  arg.flags = flags;
  arg.src = (__u64)src;
  arg.dest = (__u64)dest;
  arg.fs = (__u64)fs;
  arg.data = (__u64)data;
  arg.op = op;
  bpf_map_update_elem(&args, &tid, &arg, BPF_ANY);
  return 0;
};
```





## collect C types: type name event: not found
```sh
root@Good:~/go/src/ebpf-go/ex02-mount# make
clang \
    -target bpf \
        -D __TARGET_ARCH_x86 \
    -Wall \
    -O2 -g -o mountsnoop.o -c mountsnoop.c
llvm-strip -g mountsnoop.o
bpftool gen skeleton mountsnoop.o > mountsnoop.skel.h
echo  go build...
go build...
go get github.com/cilium/ebpf/cmd/bpf2go
go run github.com/cilium/ebpf/cmd/bpf2go  -type event  bpf mountsnoop.c -- -I../headers
Compiled /root/go/src/ebpf-go/ex02-mount/bpf_bpfel.o
Stripped /root/go/src/ebpf-go/ex02-mount/bpf_bpfel.o
Error: collect C types: type name event: not found
exit status 1
make: *** [Makefile:16: mountsnoop] 오류 1
```
### fix 
* add it in pbf.c file
```c
const struct event *unused __attribute__((unused));
```




## C source files not allowed when not using cgo or SWIG: 
* check go file  

```sh
oot@Good:~/go/src/ebpf-go/ex02-mount# make
clang \
    -target bpf \
        -D __TARGET_ARCH_x86 \
    -Wall \
    -O2 -g -o mountsnoop.o -c mountsnoop.c
llvm-strip -g mountsnoop.o
bpftool gen skeleton mountsnoop.o > mountsnoop.skel.h
echo  go build...
go build...
go get github.com/cilium/ebpf/cmd/bpf2go
go run github.com/cilium/ebpf/cmd/bpf2go  -type event  bpf mountsnoop.c -- -I../headers
Compiled /root/go/src/ebpf-go/ex02-mount/bpf_bpfel.o
Stripped /root/go/src/ebpf-go/ex02-mount/bpf_bpfel.o
Wrote /root/go/src/ebpf-go/ex02-mount/bpf_bpfel.go
Compiled /root/go/src/ebpf-go/ex02-mount/bpf_bpfeb.o
Stripped /root/go/src/ebpf-go/ex02-mount/bpf_bpfeb.o
Wrote /root/go/src/ebpf-go/ex02-mount/bpf_bpfeb.go
# go run github.com/cilium/ebpf/cmd/bpf2go -target amd64  -type event bpf mountsnoop.c -- -I../headers
# go run github.com/cilium/ebpf/cmd/bpf2go -target amd64 bpf mountsnoop.c -- -I../headers
go generate
go build  
package ebpf-go/ex02-mount: C source files not allowed when not using cgo or SWIG: mountsnoop.c
make: *** [Makefile:20: mountsnoop] 오류 1
```

==>> c file don't need go build 
좀 허망하지만 go build할때 c 파일이 있으면 안된다는 것이다. 
그래서 해당 디렉토리에 c 파일을 제거하거나
주석으로 go build에서 제거한다고 표시를 해주면 된다. 
//go:build ignore

```c
//go:build ignore

/* SPDX-License-Identifier: GPL-2.0 */
/* Copyright (c) 2021 Hengqi Chen */
#include <vmlinux.h>
...
```




## bpf_trace_printk Error
* bpf_trace_printk 함수에서 에러 발생함 
```log
root@Good:~/go/src/ebpf-go/ex02-mount# sudo  ./ex02-mount 
2024/03/14 21:59:40 loading objects: field MountEntry: program mount_entry: load program: permission denied: 11: (85) call bpf_trace_printk#6: R2 type=map_value expected=scalar (17 line(s) omitted)
root@Good:~/go/src/ebpf-go/ex02-mount# sudo  ./ex02-mount 
2024/03/14 22:02:35 loading objects: field MountEntry: program mount_entry: load program: permission denied: 11: (85) call bpf_trace_printk#6: R2 type=map_value expected=scalar (17 line(s) omitted)
```

==> 원인은 함수 사용 가이드를 준수해야 한다. 
static long (*bpf_trace_printk)(const char *fmt, __u32 fmt_size, ...) = (void *) 6;
* 매개 변수는 적어도 3개가 되어야 한다. 

다음과 같이 함수를 사용해야 한다. 

```c
  int pid = bpf_get_current_pid_tgid() >> 32;
  const char fmt_str[] = "Hello, world, from BPF! My PID is [%d]";
  bpf_trace_printk(fmt_str, sizeof(fmt_str), pid);

  bpf_printk("===>> [%d]",pid);
```


==> 그리고 중요한것은 /sys/kernel/tracing/trace_pip를 통해서 로그를 실시간으로 받으려면
* trace_on 설정하고 나서 trace_pip를 모니터링 해야 한다.  
```
# echo 1 > /sys/kernel/debug/tracing/tracing_on
# cat      /sys/kernel/debug/tracing/trace_pip
```
===> https://nakryiko.com/posts/bpf-tips-printk/


## 

```
root@Good:~/go/src/ebpf-go/ex02-mount# sudo  ./ex02-mount 
2024/03/15 00:14:55 loading objects: field UmountExit: program umount_exit: load program: invalid argument: Unreleased reference id=5 alloc_insn=25 (162 line(s) omitted)
root@Good:~/go/src/ebpf-go/ex02-mount# 
```


## collect C types: type name event: not found

```sh
root@Good:~/go/src/ebpf-go/ex05-opensnoop# make
bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h
clang \
    -target bpf \
        -D __TARGET_ARCH_x86 \
    -Wall \
    -O2 -g -o opensnoop.o -c opensnoop.c
llvm-strip -g opensnoop.o
bpftool gen skeleton opensnoop.o > opensnoop.skel.h
echo  go build...
go build...
go get github.com/cilium/ebpf/cmd/bpf2go
# go run github.com/cilium/ebpf/cmd/bpf2go  -target amd64  bpf opensnoop.c -- -I../headers
go run github.com/cilium/ebpf/cmd/bpf2go -type event bpf opensnoop.c -- -I../headers
Compiled /root/go/src/ebpf-go/ex05-opensnoop/bpf_bpfel.o
Stripped /root/go/src/ebpf-go/ex05-opensnoop/bpf_bpfel.o
Error: collect C types: type name event: not found
exit status 1
make: *** [Makefile:17: opensnoop] 오류 1
root@Good:~/go/src/ebpf-go/ex05-opensnoop# 
```

### fix 
* add it in pbf.c file
```c
const struct event *unused __attribute__((unused));
```




## type *btf.Pointer: not supported

```sh
root@Good:~/go/src/ebpf-go/ex05-opensnoop# make
clang \
    -target bpf \
        -D __TARGET_ARCH_x86 \
    -Wall \
    -O2 -g -o opensnoop.o -c opensnoop.c
llvm-strip -g opensnoop.o
bpftool gen skeleton opensnoop.o > opensnoop.skel.h
echo  go build...
go build...
go get github.com/cilium/ebpf/cmd/bpf2go
# go run github.com/cilium/ebpf/cmd/bpf2go  -target amd64  bpf opensnoop.c -- -I../headers
go run github.com/cilium/ebpf/cmd/bpf2go -type event bpf opensnoop.c -- -I../headers
Compiled /root/go/src/ebpf-go/ex05-opensnoop/bpf_bpfeb.o
Stripped /root/go/src/ebpf-go/ex05-opensnoop/bpf_bpfeb.o
Error: can't write /root/go/src/ebpf-go/ex05-opensnoop/bpf_bpfeb.go: can't generate types: template: common:17:4: executing "common" at <$.TypeDeclaration>: error calling TypeDeclaration: Struct:"args_t": field 0: type *btf.Pointer: not supported
exit status 1
make: *** [Makefile:17: opensnoop] 오류 1
```

### fix 
* not support  *btf.Pointeer 
```c
struct args_t {
	const char *fname;
	int flags;
};
```

## failed to create kprobe 'nfs_file_read+0x0' 
* libbpf: prog 'file_read_entry': failed to create kprobe 'nfs_file_read+0x0' 
* perf event: No such file or directory
* Error in bpf_object__probe_loading():Operation not permitted(1). 
* Couldn't load trivial BPF program. 
* Make sure your kernel supports BPF (CONFIG_BPF_SYSCALL=y) and/or that RLIMIT_MEMLOCK is set to big enough value.
```sh
$ ./nfsslower 
libbpf: Failed to bump RLIMIT_MEMLOCK (err = -1), you might need to do it explicitly!
libbpf: Error in bpf_object__probe_loading():Operation not permitted(1). Couldn't load trivial BPF program. Make sure your kernel supports BPF (CONFIG_BPF_SYSCALL=y) and/or that RLIMIT_MEMLOCK is set to big enough value.
libbpf: failed to load object 'fsslower_bpf'
libbpf: failed to load BPF skeleton 'fsslower_bpf': -1
failed to load BPF object: -1


j$ sudo  ./nfsslower 
[sudo] jhyunlee 암호: 
libbpf: prog 'file_read_entry': failed to create kprobe 'nfs_file_read+0x0' perf event: No such file or directory
failed to attach kprobe: -2
failed to attach BPF programs: -2
```

==> nfs kernel 모듈이 load 되었는지 부터 확인하라.


https://github.com/iovisor/bcc/issues/3438

```sh
root@Good:/lib/modules/6.5.0-28-generic/kernel/fs# sudo modprobe nfs

root@Good:/lib/modules/6.5.0-28-generic/kernel/fs# lsmod | grep nfs
nfs                   581632  0
lockd                 143360  1 nfs
fscache               389120  1 nfs
netfs                  61440  2 fscache,nfs
sunrpc                811008  3 lockd,nfs

root@Good:/lib/modules/6.5.0-28-generic/kernel/fs# ls
9p      befs            ceph    erofs     fscache  hpfs   minix       nilfs2  omfs       qnx6      sysv    xfs
adfs    bfs             coda    exfat     fuse     isofs  netfs       nls     orangefs   quota     ubifs   zonefs
affs    binfmt_misc.ko  cramfs  f2fs      gfs2     jffs2  nfs         ntfs    overlayfs  reiserfs  udf
afs     btrfs           dlm     fat       hfs      jfs    nfs_common  ntfs3   pstore     romfs     ufs
autofs  cachefiles      efs     freevxfs  hfsplus  lockd  nfsd        ocfs2   qnx4       smb       vboxsf



root@Good:/lib/modules/6.5.0-28-generic/kernel/fs# grep   nfs_file_read  /proc/kallsyms 
ffffffffc26e2dc0 t nfs_file_read	[nfs]

root@Good:/sys/kernel/debug/tracing# grep nfs_file_read avail*
available_filter_functions:kernfs_file_read_iter
available_filter_functions:nfs_file_read [nfs]
available_filter_functions_addrs:ffffffff99383ea0 kernfs_file_read_iter
available_filter_functions_addrs:ffffffffc26e2dc0 nfs_file_read [nfs]

```

==>

```sh
jhyunlee@Good:~/go/src/eBPF/bcc/libbpf-tools$ sudo ./fsslower -t xfs
libbpf: prog 'file_read_entry': failed to create kprobe 'xfs_file_read_iter+0x0' perf event: No such file or directory
failed to attach kprobe: -2
failed to attach BPF programs: -2
jhyunlee@Good:~/go/src/eBPF/bcc/libbpf-tools$ sudo ./fsslower -t nfs
Tracing nfs operations slower than 10 ms... Hit Ctrl-C to end.
TIME     COMM             PID     T BYTES   OFF_KB   LAT(ms) FILENAME

```


### bpftool not found 

```sh
jh@client:~/code/ebpf-go/ex05-opensnoop$ make
bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h
/bin/sh: 1: bpftool: not found
make: *** [Makefile:36: vmlinux.h] Error 127
```
==> install bpftool 
* ~/code/eBPF/bcc/libbpf-tools/bpftool

```sh
$ sudo apt-get update -y
$ sudo  apt-get install -y build-essential curl libbpf-dev clang libelf-dev linux-tools-$(uname -r) llvm
$ git clone --recurse-submodules https://github.com/libbpf/bpftool.git
$ git submodule update --init
$ cd  libbpf/src
$ make 
$ sudo  make install

$ sudo apt install llvm
$ cd  bpftool/src
$ make
$ export  LANG=C
$ make
...                        libbfd: [ OFF ]
...               clang-bpf-co-re: [ on  ]
...                          llvm: [ OFF ]
...                        libcap: [ OFF ]
  GEN      profiler.skel.h
  CC       prog.o
  CC       struct_ops.o
  CC       tracelog.o
  CC       xlated_dumper.o
  CC       disasm.o
  LINK     bpftool

$ sudo make install
[sudo] password for jhyunlee: 
...                        libbfd: [ OFF ]
...               clang-bpf-co-re: [ on  ]
...                          llvm: [ OFF ]
...                        libcap: [ OFF ]
  INSTALL  bpftool
$ bpftool

Usage: bpftool [OPTIONS] OBJECT { COMMAND | help }
       bpftool batch file FILE
       bpftool version

       OBJECT := { prog | map | link | cgroup | perf | net | feature | btf | gen | struct_ops | iter }
       OPTIONS := { {-j|--json} [{-p|--pretty}] | {-d|--debug} |
                    {-V|--version} }

$ which bpftool
/usr/local/sbin/bpftool
```

==> Install and compile BCC
```sh
$ git clone https://github.com/iovisor/bcc.git
$ mkdir bcc/build; cd bcc/build
$ cmake ..
$ make
$ sudo make install
$ cmake -DPYTHON_CMD=python3 .. # build python3 binding
$ pushd src/python/
$ make
$ sudo make install
$ popd
```





