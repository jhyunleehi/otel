# bcc Python Developer Tutorial

This tutorial is about developing [bcc](https://github.com/iovisor/bcc) tools and programs using the Python interface. 
* There are two parts: observability then networking. Snippets are taken from various programs in bcc: see their files for licences.

Also see the bcc developer's [reference_guide.md](reference_guide.md), and a tutorial for end-users of tools: [tutorial.md](tutorial.md). There is also a lua interface for bcc.

* Observability

This observability tutorial contains 17 lessons, and 46 enumerated things to learn.

* vscode .launch.json
```json
{
            "name": "Python: sudo Run",
            "type": "debugpy",
            "request": "launch",
            "program": "${file}",
            "python": "python3",
            "sudo": true,
            "justMyCode": false,            
            "console": "integratedTerminal",            
            "args": [
                "-v",
                "-s",
                "--debuglevel==DEBUG"
            ]
        }
```        

## Lesson 1. Hello World

Start by running [examples/hello_world.py](../examples/hello_world.py), while running some commands (eg, "ls") in another session. It should print "Hello, World!" for new processes. If not, start by fixing bcc: see [INSTALL.md](../INSTALL.md).
* 가능하면 source 코드를 다운 받아서 실행한다. 

```
# ./examples/hello_world.py
            bash-13364 [002] d... 24573433.052937: : Hello, World!
            bash-13364 [003] d... 24573436.642808: : Hello, World!
[...]
```

Here's the code for hello_world.py:

```Python
from bcc import BPF
BPF(text='int kprobe__sys_clone(void *ctx) { bpf_trace_printk("Hello, World!\\n"); return 0; }').trace_print()
```
### 해설
There are six things to learn from this:
1. `text='...'`: This defines a BPF program inline. The program is written in C.

2. `kprobe__sys_clone()`: kprobe__ 접두사로 시작하면 kprobes를 사용한다는 의미이고, 그 위에 붙은 것은 커널 함수를 의미한다.  This is a short-cut for kernel dynamic tracing via kprobes. If the C function begins with `kprobe__`, the rest is treated as a kernel function name to instrument, in this case, `sys_clone()`. `kprobe__` 은 kprobe (dynamic tracing of a kernel function call)을 기능을 제공한다.  `BPF.attach_kprobe()` python 함수를 이용해서 kernel 함수에 붙이는 기능을 수행한다.  

3. `void *ctx`: ctx has arguments, but since we aren't using them here, we'll just cast it to `void *`.

4. `bpf_trace_()`:  ftrace를 이용하는 /sys/kernel/debug/tracing/tracing_on 해서 ftrace 하면  /sys/kernle/debug/traice/trace_pipe 한다. 커널에서 printf 사용 기능. (/sys/kernel/debug/tracing/trace_pipe).단순하게 뭔가 출력할때 사용하는 것은 ok 이지만 최대 arg는 3개만 사용할 수 있는 제약이 있다. trace_pipe는 global share 하기 때문에 출력이 충돌된다. 아무튼 BPF_PERF_OUTPUT을 사용해야 한다. This is ok for some quick examples, but has limitations: 3 args max, 1 %s only, and trace_pipe is globally shared, so concurrent programs will have clashing output. A better interface is via BPF_PERF_OUTPUT(), covered later.

5. `return 0;`: return에 대한 필요한 기본 요구사항, Necessary formality (if you want to know why, see [#139](https://github.com/iovisor/bcc/issues/139)).

6. `.trace_print()`: ptyhon 함수에서 BPF의 trace_print는  /sys/kernel/debug/tracing/trace_pipe에 있는 것을 출력 한다.  A bcc routine that reads trace_pipe and prints the output. 

```py
def trace_print(self, fmt=None):       
    while True:
        if fmt:
            fields = self.trace_fields(nonblocking=False)
            if not fields: continue
            line = fmt.format(*fields)
        else:
            line = self.trace_readline(nonblocking=False)
        print(line)
        sys.stdout.flush()

```


## Lesson 2. sys_sync()

Write a program that traces the sys_sync() kernel function. Print "sys_sync() called" when it runs. Test by running ```sync``` in another session while tracing. The hello_world.py program has everything you need for this.

Improve it by printing "Tracing sys_sync()... Ctrl-C to end." when the program first starts. Hint: it's just Python.

```py
#!/usr/bin/python
# Copyright (c) PLUMgrid, Inc.
# Licensed under the Apache License, Version 2.0 (the "License")

from bcc import BPF
from bcc.utils import printb

prog="""
int sync(void *cts){
    bpf_trace_printk("sync\\n");
    return 0;
}
"""
b=BPF(text=prog)
b.attach_kprobe(event=b.get_syscall_fnname("sync"), fn_name="sync")
print(f"Tracing sys_sync()... Ctrl-C to end.")

while True:
    try:
        task,pid,cpu,flags,ts,msg=b.trace_fields()
    except ValueError:
        continue
    except KeyboardInterrupt:
        break
    printb(b"%-18.9f %-16s %-6d %s" % (ts, task, pid, msg))
    printb(b"%18.9f %16s %6d %s" % (ts, task, pid, msg))
```
### 해설
* 이것을 간단한게 구현해 보면 다른 출력들과 섞어져서  나온다는 점이 좀 그렇다.
* /sys/kernel/debug/tracing/trace_pipe 내용으로 출력이 나온 것을 다시 읽어서 python 라이브러리로 나오게 한다
* kernel에서 나오는 데이터 들은 byte 데이터 들이라서 이것을 byte로 처리하는 작업들이 필요하다. 
* 문자열들을 다시 int 또는 float 변환하는 작업들...
* 여기서 printb는 bcc의 함수이다.  

## Lesson 3. hello_fields.py

This program is in [examples/tracing/hello_fields.py](../examples/tracing/hello_fields.py). Sample output (run commands in another session):

```
# ./examples/tracing/hello_fields.py
TIME(s)            COMM             PID    MESSAGE
24585001.174885999 sshd             1432   Hello, World!
24585001.195710000 sshd             15780  Hello, World!
24585001.991976000 systemd-udevd    484    Hello, World!
24585002.276147000 bash             15787  Hello, World!
```

Code:

```Python
from bcc import BPF
from bcc.utils import printb

# define BPF program
prog = """
int hello(void *ctx) {
    bpf_trace_printk("Hello, World!\\n");
    return 0;
}
"""

# load BPF program
b = BPF(text=prog)
b.attach_kprobe(event=b.get_syscall_fnname("clone"), fn_name="hello")

# header
print("%-18s %-16s %-6s %s" % ("TIME(s)", "COMM", "PID", "MESSAGE"))

# format output
while 1:
    try:
        (task, pid, cpu, flags, ts, msg) = b.trace_fields()
    except ValueError:
        continue
    except KeyboardInterrupt:
        exit()
    printb(b"%-18.9f %-16s %-6d %s" % (ts, task, pid, msg))
```
### 해설

This is similar to hello_world.py, and traces new processes via sys_clone() again, but has a few more things to learn:

1. `prog =`: This time we declare the C program as a variable, and later refer to it. CLI 기반에서 어떤 문자열을 대체하려고 할 때 좀 유용하다. This is useful if you want to add some string substitutions based on command line arguments.

2. `hello()`: lession1에서 kprobe__ 접두사로 system call을 추적하는 대신에 C 함수를 구현한 것이다. 이 함수는 kbrobe에서 사용할 것이기 때문에 pt_reg* ctx를 파라미터를 지정해야 한다. 만약에  probe하는 기능을 사용하지 않으려면  static inline 을 표시행서 컴파일러에서 inline 적용하도록 해야 한다.  Now we're just declaring a C function, instead of the ```kprobe__``` shortcut. We'll refer to this later. All C functions declared in the BPF program are expected to be executed on a probe, hence they all need to take a `pt_reg* ctx` as first argument. If you need to define some helper function that will not be executed on a probe, they need to be defined as `static inline` in order to be inlined by the compiler. Sometimes you would also need to add `_always_inline` function attribute to it.

3. `b.attach_kprobe(event=b.get_syscall_fnname("clone"), fn_name="hello")`: 커널 System Call중에서 new thread 생성할 때 clone이 사용되는데 이것이 호출되는 probe해서  hello 함수를 호출하도록 하는 기능을 구현한 것이다.  probe Creates a kprobe for the kernel clone system call function, which will execute our defined hello() function. You can call attach_kprobe() more than once, and attach your C function to multiple kernel functions.

4. `b.trace_fields()`: BPF.trace_fields()함수를 호출하면 /sys/kernel/debug/tracing/trace_pipe에서 고정된 필드의 집합을 돌려 준다? 아무튼 나중에서는 BPF_PERF_OUPUT을 사용한다는 것이다.   Returns a fixed set of fields from trace_pipe. Similar to trace_print(), this is handy for hacking, but for real tooling we should switch to BPF_PERF_OUTPUT().

### BPF.trace_fields
아래 코드를 보면 trace_readline을 nonblocking 모드로 읽어와서..

  ```py
      def trace_fields(self, nonblocking=False):
        """trace_fields(nonblocking=False)

        Read from the kernel debug trace pipe and return a tuple of the
        fields (task, pid, cpu, flags, timestamp, msg) or None if no
        line was read (nonblocking=True)
        """
        while True:
            line = self.trace_readline(nonblocking)디
            if not line and nonblocking: return (None,) * 6
            # don't print messages related to lost events
            if line.startswith(b"CPU:"): continue
            task = line[:16].lstrip()
            line = line[17:]
            ts_end = line.find(b":")
            try:
                pid, cpu, flags, ts = line[:ts_end].split()
            except Exception as e:
                continue
            cpu = cpu[1:-1]
            # line[ts_end:] will have ": [sym_or_addr]: msgs"
            # For trace_pipe debug output, the addr typically
            # is invalid (e.g., 0x1). For kernel 4.12 or earlier,
            # if address is not able to match a kernel symbol,
            # nothing will be printed out. For kernel 4.13 and later,
            # however, the illegal address will be printed out.
            # Hence, both cases are handled here.
            line = line[ts_end + 1:]
            sym_end = line.find(b":")
            msg = line[sym_end + 2:]
            try:
                return (task, int(pid), int(cpu), flags, float(ts), msg)
            except Exception as e:
                return ("Unknown", 0, 0, "Unknown", 0.0, "Unknown")
   ```             
* line = self.trace_readline(nonblocking)의 리턴 값은 아래와 같은데 이것을 잘 짤라 내서 파싱해서 리턴 값으로 돌려 준다. json이나 html 이라면 파싱이 쉬울 텐데.. 이런 문자열 파싱은 나중에 꼭 문제 되던데... ㅎㅎ
```   
b'           <...>-322350  [010] ...21 223559.266945: bpf_trace_printk: Hello, World!'   
b'322350  [010] ...21 223559.264003: bpf_trace_printk: Hello, World!'
```




## Lesson 4. sync_timing.py

Remember the days of sysadmins typing `sync` three times on a slow console before `reboot`, to give the first asynchronous sync time to complete? Then someone thought `sync;sync;sync` was clever, to run them all on one line, which became industry practice despite defeating the original purpose! And then sync became synchronous, so more reasons it was silly. Anyway.

The following example times how quickly the ```do_sync``` function is called, and prints output if it has been called more recently than one second ago. A ```sync;sync;sync``` will print output for the 2nd and 3rd sync's:

```
# ./examples/tracing/sync_timing.py
Tracing for quick sync's... Ctrl-C to end
At time 0.00 s: multiple syncs detected, last 95 ms ago
At time 0.10 s: multiple syncs detected, last 96 ms ago
```

This program is [examples/tracing/sync_timing.py](../examples/tracing/sync_timing.py):

```Python
from __future__ import print_function
from bcc import BPF
from bcc.utils import printb

# load BPF program
b = BPF(text="""
#include <uapi/linux/ptrace.h>

BPF_HASH(last);

int do_trace(struct pt_regs *ctx) {
    u64 ts, *tsp, delta, key = 0;

    // attempt to read stored timestamp
    tsp = last.lookup(&key);
    if (tsp != NULL) {
        delta = bpf_ktime_get_ns() - *tsp;
        if (delta < 1000000000) {
            // output if time is less than 1 second
            bpf_trace_printk("%d\\n", delta / 1000000);
        }
        last.delete(&key);
    }

    // update stored timestamp
    ts = bpf_ktime_get_ns();
    last.update(&key, &ts);
    return 0;
}
""")

b.attach_kprobe(event=b.get_syscall_fnname("sync"), fn_name="do_trace")
print("Tracing for quick sync's... Ctrl-C to end")

# format output
start = 0
while 1:
    try:
        (task, pid, cpu, flags, ts, ms) = b.trace_fields()
        if start == 0:
            start = ts
        ts = ts - start
        printb(b"At time %.2f s: multiple syncs detected, last %s ms ago" % (ts, ms))
    except KeyboardInterrupt:
        exit()
```

### 해설

1. `bpf_ktime_get_ns()`: Returns the time as nanoseconds. 시스템 부팅 이후로 부터 경과된 nano 시간을 제공한다.  bpf_helper 함수이다.  
```c
#include <bpf/bpf_helpers.h>
/*
 * bpf_ktime_get_ns
 * 	Return the time elapsed since system boot, in nanoseconds.
 * 	Does not include time the system was suspended.
 * 	See: **clock_gettime**\ (**CLOCK_MONOTONIC**)
 * Returns : 	Current *ktime*.
 */
static __u64 (*bpf_ktime_get_ns)(void) = (void *) 5;
```
2. `BPF_HASH(last)`: Creates a BPF map objec설t that is a hash (associative array), called "last". We didn't specify any further arguments, so it defaults to key and value types of u64. BPF 맵 객체를 만든다. map의 이름만 사용하고, key 값은 unsigined 64를 사용한다.  그런데 이 BPF_HASH는 macro 처럼 보이는데 어디에 이것을 정의하고 있는지 찾을 수 없다.  어딘가에는 있겠지..
3. `key = 0`: We'll only store one key/value pair in this hash, where the key is hardwired to zero.
4. `last.lookup(&key)`: Lookup the key in the hash, and return a pointer to its value if it exists, else NULL. We pass the key in as an address to a pointer. hash 맵에 대해서 key값으로 조회를 해서 valuse가 있으면 pointer가 리턴되고 없으면 NULL이 리턴된다. 
5. `if (tsp != NULL) {`: The verifier requires that pointer values derived from a map lookup must be checked for a null value before they can be dereferenced and used. 
6. `last.delete(&key)`: Delete the key from the hash. This is currently required because of [a kernel bug in `.update()`](https://git.kernel.org/cgit/linux/kernel/git/davem/net.git/commit/?id=a6ed3ea65d9868fdf9eff84e6fe4f666b8d14b02) (fixed in 4.8.10). 여기서 해당 key를 대상으로 삭제를  꼭해야 하는지? 어차피  last.update를 통해서 값을 갱신하기 때문에 필요 없는 코드 아니가?. 뭔가 ...교육적 차원에서 last.delete 하는 것은 아닌가?

7. `last.update(&key, &ts)`: Associate the value in the 2nd argument to the key, overwriting any previous value. This records the timestamp.

* 출력이 되는 것은 bpf_trace_printk에 써 놓은 데이터가 나오는 것..
* hash map에 써 놓은 데이터를 출력 받아야 할텐데..

#### BPF_HASH
#### hashmap.lookup(key)
#### hashmap.delete(kdy)
#### hashmap.update(key,value)


## Lesson 5. sync_count.py
위의 예제에서는 delta 값이  if (delta < 1,000,000,000) ns 이하 일때 만 즉 1초 이하 일때 만 bpf_trace_printk로 출력해서  /sys/kernel/debug/trace/trace_pipe에 출력하도록 하고 있다. 이것을  모두 출력하도록 수정하라는 이야기...
* count를 기존의 hash 맵에 새로운 key 인덱스 추가해서 BPF 프로그램에 기록할 수 있다.
* 기존 hash map에 key를 하나 더 추가해서 count 값을 저장하자.
Modify the sync_timing.py program (prior lesson) to store the count of all kernel sync system calls (both fast and slow), and print it with the output. This count can be recorded in the BPF program by adding a new key index to the existing hash.


### 문제풀이
```py
#!/usr/bin/python

from __future__ import print_function
from bcc import BPF
from bcc.utils import printb

# load BPF program
b = BPF(text="""
#include <uapi/linux/ptrace.h>

BPF_HASH(last);

int do_trace(struct pt_regs *ctx) {
    u64 ts, *tsp, delta,  *cnt, count;
    u64  key1 = 0, key2=1;

    // attempt to read stored timestamp
    tsp = last.lookup(&key1);
    if (tsp != NULL) {
        cnt = last.lookup(&key2);
        if (cnt == NULL){
            count=1;
            last.update(&key2, &count);    
        } else {
            count=*cnt+1;
            last.update(&key2, &count);    
        }
        delta = bpf_ktime_get_ns() - *tsp;             
        bpf_trace_printk("[%d][%d] \\n", delta / 1000000 , count);
        last.delete(&key1);
    }    
    // update stored timestamp
    ts = bpf_ktime_get_ns();
    last.update(&key1, &ts);
    return 0;
}
""")

b.attach_kprobe(event=b.get_syscall_fnname("sync"), fn_name="do_trace")
print("Tracing for quick sync's... Ctrl-C to end")

# format output
start = 0
while 1:
    try:
        (task, pid, cpu, flags, ts, ms) = b.trace_fields()
        if start == 0:
            start = ts
        ts = ts - start
        printb(b"At time %.2f s: multiple syncs detected, last %s ms ago" % (ts, ms))
    except KeyboardInterrupt:
        exit()
```
* BPF_HASH에 값을 넣을 때는 map.update(u64 *key,*valuse) 
* map.lookup(u64 *key) 그리고 이것에 return 에 대해서 NULL 체크를 해야한다. 

## Lesson 6. disksnoop.py
* hash map에 key type을 지정해서 생성하는 방법
* 여기서 부터는 syscall에 붙이는 것이 아니라 kprobe에 붙이려고 한다.  

Browse the [examples/tracing/disksnoop.py](../examples/tracing/disksnoop.py) program to see what is new. Here is some sample output:

```
# ./disksnoop.py
TIME(s)            T  BYTES    LAT(ms)
16458043.436012    W  4096        3.13
16458043.437326    W  4096        4.44
16458044.126545    R  4096       42.82
16458044.129872    R  4096        3.24
[...]
```

And a code snippet:

```Python
[...]
REQ_WRITE = 1		# from include/linux/blk_types.h

# load BPF program
b = BPF(text="""
#include <uapi/linux/ptrace.h>
#include <linux/blk-mq.h>

BPF_HASH(start, struct request *);

void trace_start(struct pt_regs *ctx, struct request *req) {
	// stash start timestamp by request ptr
	u64 ts = bpf_ktime_get_ns();

	start.update(&req, &ts);
}

void trace_completion(struct pt_regs *ctx, struct request *req) {
	u64 *tsp, delta;

	tsp = start.lookup(&req);
	if (tsp != 0) {
		delta = bpf_ktime_get_ns() - *tsp;
		bpf_trace_printk("%d %x %d\\n", req->__data_len, req->cmd_flags, delta / 1000);
		start.delete(&req);
	}
}
""")
if BPF.get_kprobe_functions(b"blk_start_request"):
        b.attach_kprobe(event="blk_start_request", fn_name="trace_start")
b.attach_kprobe(event="blk_mq_start_request", fn_name="trace_start")
if BPF.get_kprobe_functions(b"__blk_account_io_done"):
    b.attach_kprobe(event="__blk_account_io_done", fn_name="trace_completion")
else:
    b.attach_kprobe(event="blk_account_io_merge_bio", fn_name="trace_completion")
[...]
```

### 해설: 
* 중요한 점은 BPF_HASH를 사용하는 이유는 이것이 event handler 처럼 동작하기 때문에 전단계에서 발생한 값을 저장하는 방법이 마땅치 않다는 것이다. 그래서 전 단계에서 발생한 값을 hash에 넣어 놓고 그것을 다음 단계에서 사용할때 바로 이 hash map을 사용한다는 것이다.

1. `REQ_WRITE`: We're defining a kernel constant in the Python program because we'll use it there later. If we were using REQ_WRITE in the BPF program, it should just work (without needing to be defined) with the appropriate #includes.
2. `trace_start(struct pt_regs *ctx, struct request *req)`: kprobes에 붙일려고 만든 함수, This function will later be attached to kprobes. The arguments to kprobe functions are `struct pt_regs *ctx`, for registers and BPF context, and then the actual arguments to the function. We'll attach this to blk_start_request(), where the first argument is `struct request *`.
3. `start.update(&req, &ts)`:  hash map에 데이터를 저장할때 request struct pointer가 uniq하기 때문에 key로 사용하기 좋다. We're using the pointer to the request struct as a key in our hash. What? This is common place in tracing. Pointers to structs turn out to be great keys, as they are unique: two structs can't have the same pointer address. (Just be careful about when it gets free'd and reused.) So what we're really doing is tagging the request struct, which describes the disk I/O, with our own timestamp, so that we can time it. There's two common keys used for storing timestamps: pointers to structs, and, thread IDs (for timing function entry to return).
4. `req->__data_len`: We're dereferencing members of `struct request`. See its definition in the kernel source for what members are there. bcc actually rewrites these expressions to be a series of `bpf_probe_read_kernel()` calls. Sometimes bcc can't handle a complex dereference, and you need to call `bpf_probe_read_kernel()` directly.

This is a pretty interesting program, and if you can understand all the code, you'll understand many important basics. We're still using the bpf_trace_printk() hack, so let's fix that next.


#### Error 발생 
* 원인이 무엇인지는 모르겠지만 실행이 잘 안된다.
* trace_completion 프로그램을 blk_account_io_done에 붙일려고 하는데 할 수 없다는 내용 같은데 

```log
Exception has occurred: Exception       (note: full exception trace is shown but execution is paused at: _run_module_as_main)
Failed to attach BPF program b'trace_completion' to kprobe b'blk_account_io_done', it's not traceable (either non-existing, inlined, or marked as "notrace")
  File "/home/jhyunlee/.local/lib/python3.10/site-packages/bcc/__init__.py", line 855, in attach_kprobe
    raise Exception("Failed to attach BPF program %s to kprobe %s"
  File "/home/jhyunlee/code/eBPF/bcc/examples/tracing/disksnoop.py", line 52, in <module>
    b.attach_kprobe(event="blk_account_io_done", fn_name="trace_completion")
  File "/usr/lib/python3.10/runpy.py", line 86, in _run_code
    exec(code, run_globals)
  File "/usr/lib/python3.10/runpy.py", line 196, in _run_module_as_main (Current frame)
```
이 오류는 blk_account_io_done에 대한 kprobe를 생성하려고 할 때 trace_completion이라는 BPF 프로그램을 연결하지 못했다는 것을 나타냅니다. 이것은 추적할 수 없는 상태일 수 있습니다. 일반적으로 이러한 문제는 다음 중 하나로 인해 발생합니다.

1. 존재하지 않는 경우: blk_account_io_done 또는 trace_completion 중 하나가 실제로 존재하지 않는 경우.
2. 인라인 된 경우: blk_account_io_done 또는 trace_completion 중 하나가 인라인되어 추적할 수 없는 경우.
3. "notrace"로 표시된 경우: 해당 kprobe 또는 BPF 프로그램이 추적되지 않도록 "notrace"로 표시된 경우

그래서 /sys/kernel/debug/tracing/available_* 에서 찾아보면 ...

```log
root@Good:/sys/kernel/debug/tracing# grep  blk_account_io  avail*
available_filter_functions:blk_account_io_merge_bio
available_filter_functions:blk_account_io_completion.part.0
available_filter_functions_addrs:ffffffff87f77a90 blk_account_io_merge_bio
available_filter_functions_addrs:ffffffff87f7b5b0 blk_account_io_completion.part.0
```
그리고 kernel 심볼 /proc/kallsysms 에서 찾아보면 
```
$ cat /proc/kallsyms | grep blk_account_io_done
```
==> 없다. 그래서  blk_account_io_merge_bio를 대신해서 trace kprobe funtion으로 사용한다.  

그러면 잘 동작한다.  

#### system call과 kernel function에 대해서 event attach 방법
* 여기 예제를 보면 user space에서 kernel로 접근할때 사용하는 System Call 이름은 시스템 마다 달라서 이름을 찾아와서 붙이는 작업을 한다. 
* 하지만 kernel funtion 같은 경우는 그냥  /sys/kernel/debug/tracing/available_filter_function 에서 찾아서 있으면 그것을 사용하면 됩니다. 
```
b.attach_kprobe(event=b.get_syscall_fnname("sync"), fn_name="sync")
b.attach_kprobe(event=b.get_syscall_fnname("clone"), fn_name="hello")
b.attach_kprobe(event=b.get_syscall_fnname("sync"), fn_name="do_trace")
b.attach_kprobe(event="blk_start_request", fn_name="trace_start")
b.attach_kprobe(event="blk_mq_start_request", fn_name="trace_start")
b.attach_kprobe(event="blk_account_io_done", fn_name="trace_completion")
b.attach_kprobe(event="finish_task_switch", fn_name="count_sched")
```

## Lesson 7. hello_perf_output.py
이제 bpf_trace_printK 함수는 그만 사용하고 BPF_PERF_OUTPUT 인터페이스를 사용해서 처리하자.  당연히 trace_feild 함수를 이용해서 문자열을 파싱하는 것도 이제는 그만하자. ㅎㅎ 
* c program 영역에서 구조체를 만들어서 BPF_PERF_OUTPUT 스트림에 연결한다.
* python에서는 생성된 스트림에 call back 함수를 생성하여 등록한다. 

Let's finally stop using bpf_trace_printk() and use the proper BPF_PERF_OUTPUT() interface. This will also mean we stop getting the free trace_field() members like PID and timestamp, and will need to fetch them directly. Sample output while commands are run in another session:

```
# ./hello_perf_output.py
TIME(s)            COMM             PID    MESSAGE
0.000000000        bash             22986  Hello, perf_output!
0.021080275        systemd-udevd    484    Hello, perf_output!
0.021359520        systemd-udevd    484    Hello, perf_output!
0.021590610        systemd-udevd    484    Hello, perf_output!
[...]
```

Code is [examples/tracing/hello_perf_output.py](../examples/tracing/hello_perf_output.py):

```Python
from bcc import BPF

# define BPF program
prog = """
#include <linux/sched.h>

// define output data structure in C
struct data_t {
    u32 pid;
    u64 ts;
    char comm[TASK_COMM_LEN];
};
BPF_PERF_OUTPUT(events);

int hello(struct pt_regs *ctx) {
    struct data_t data = {};

    data.pid = bpf_get_current_pid_tgid();
    data.ts = bpf_ktime_get_ns();
    bpf_get_current_comm(&data.comm, sizeof(data.comm));

    events.perf_submit(ctx, &data, sizeof(data));

    return 0;
}
"""

# load BPF program
b = BPF(text=prog)
b.attach_kprobe(event=b.get_syscall_fnname("clone"), fn_name="hello")

# header
print("%-18s %-16s %-6s %s" % ("TIME(s)", "COMM", "PID", "MESSAGE"))

# process event
start = 0
def print_event(cpu, data, size):
    global start
    event = b["events"].event(data)
    if start == 0:
            start = event.ts
    time_s = (float(event.ts - start)) / 1000000000
    print("%-18.9f %-16s %-6d %s" % (time_s, event.comm, event.pid,
        "Hello, perf_output!"))

# loop with callback to print_event
b["events"].open_perf_buffer(print_event)
while 1:
    b.perf_buffer_poll()
```

### 해설

1. `struct data_t`:  커널영역에서 사용자 영역으로 데이터를 전달할때 사용하려는 데이터 구조체를 정의한다.  This defines the C struct we'll use to pass data from kernel to user space.
2. `BPF_PERF_OUTPUT(events)`: 출력 채널을 정의 한다. 이것은 macro로 정의된것 같은데 어디서 정의해 놨는지를 모르겠음.  This names our output channel "events".
3. `struct data_t data = {};`:  빈 객체 생성,  Create an empty data_t struct that we'll then populate.
4. `bpf_get_current_pid_tgid()`: bpf helper 함수이다.  * bpf_get_current_pid_tgid    
    * Get the current pid and tgid.        
    ```
    data.pid = bpf_get_current_pid_tgid() >> 32;
    data.uid = bpf_get_current_uid_gid() & 0xFFFFFFFF;
   ```
    
5. `bpf_get_current_comm()`: 현재 process의 이름을 이 함수의 첫번째 매개변수의 주소에 채워 넣는다.  Populates the first argument address with the current process name.
6. `events.perf_submit()`: userspace에서 읽을 수 있도록 링 버퍼를 이용하는 BPF_PERF_OUTPUT(events)에 채워 넣는다.  Submit the event for user space to read via a perf ring buffer.
7. `def print_event()`: kernel 공간에서 저장한 "event"장치에서 값을 처리할수 있는 handler를 생성한다. 결국 callback 함수를 생성해서 등록한다는 것... Define a Python function that will handle reading events from the `events` stream.
8. `b["events"].event(data)`: C 프로그램 부분에서 생성한 event data를 python 영역으로 가져와서 참조할 수 있도록 한다.   Now get the event as a Python object, auto-generated from the C declaration.
9. `b["events"].open_perf_buffer(print_event)`: Python print_event 함수를 events 스트림과 연결합니다. Associate the Python `print_event` function with the `events` stream.
10. `while 1: b.perf_buffer_poll()`: Block waiting for events.




## Lesson 8. sync_perf_output.py

Rewrite sync_timing.py, from a prior lesson, to use ```BPF_PERF_OUTPUT```.
#### 연습문제 풀이

```py
#!/usr/bin/python

from __future__ import print_function
from bcc import BPF
from bcc.utils import printb


prog="""
#include <uapi/linux/ptrace.h>
#include <linux/sched.h>

BPF_HASH(last);

struct data_t {
    int pid;
    int uid;
    u64 ts;
    u64 delta;
    char comm[TASK_COMM_LEN];
};

BPF_PERF_OUTPUT(events);

int do_trace(struct pt_regs *ctx) {
    u64 ts, *tsp, delta, key = 0;
    
    struct data_t data={};
    data.pid = bpf_get_current_pid_tgid() >> 32;
    data.uid = bpf_get_current_uid_gid() & 0xFFFFFFFF;
    data.ts=bpf_ktime_get_ns();
    bpf_get_current_comm(&data.comm, sizeof(data.comm));
    
    // attempt to read stored timestamp
    tsp = last.lookup(&key);
    if (tsp != NULL) {
        data.delta = bpf_ktime_get_ns() - *tsp;                    
        //bpf_trace_printk("%d\\n", delta / 1000000);
        events.perf_submit(ctx, &data, sizeof(data));         
        last.delete(&key);
    }

    // update stored timestamp
    ts = bpf_ktime_get_ns();
    last.update(&key, &ts);
    return 0;
}
"""
b = BPF(text=prog)
b.attach_kprobe(event=b.get_syscall_fnname("sync"), fn_name="do_trace")
print("Tracing for quick sync's... Ctrl-C to end")

# format output
start = 0
def print_event(cpu, data, size):
    global start
    event = b["events"].event(data)
    if start == 0:
        start = event.ts
    time_s = (float(event.ts - start)) / 1000000000
    printb(b"[%-18.9f] [%-6d] [%-6d] [%-6d] [%s]" % (time_s, event.pid, event.uid, event.delta, event.comm))
    #printb(b"%-18.9f %-16s %-6d %s" % (time_s, event.comm, event.pid, b"Hello, perf_output!"))

# loop with callback to print_event
b["events"].open_perf_buffer(print_event)
while 1:
    try:
        b.perf_buffer_poll()
    except KeyboardInterrupt:
        exit()
```        

## Lesson 9. bitehist.py

The following tool records a histogram of disk I/O sizes. Sample output:

```
# ./bitehist.py
Tracing... Hit Ctrl-C to end.
^C
     kbytes          : count     distribution
       0 -> 1        : 3        |                                      |
       2 -> 3        : 0        |                                      |
       4 -> 7        : 211      |**********                            |
       8 -> 15       : 0        |                                      |
      16 -> 31       : 0        |                                      |
      32 -> 63       : 0        |                                      |
      64 -> 127      : 1        |                                      |
     128 -> 255      : 800      |**************************************|
```

Code is [examples/tracing/bitehist.py](../examples/tracing/bitehist.py):

```Python
from __future__ import print_function
from bcc import BPF
from time import sleep

# load BPF program
b = BPF(text="""
#include <uapi/linux/ptrace.h>
#include <linux/blkdev.h>

BPF_HISTOGRAM(dist);

int kprobe__blk_account_io_done(struct pt_regs *ctx, struct request *req)
{
	dist.increment(bpf_log2l(req->__data_len / 1024));
	return 0;
}
""")

# header
print("Tracing... Hit Ctrl-C to end.")

# trace until Ctrl-C
try:
	sleep(99999999)
except KeyboardInterrupt:
	print()

# output
b["dist"].print_log2_hist("kbytes")
```

A recap from earlier lessons:

* `kprobe__`: 커널 함수를 사용한다는 접두사. This prefix means the rest will be treated as a kernel function name that will be instrumented using kprobe.
* `struct pt_regs *ctx, struct request *req`: Arguments to kprobe. The `ctx` is registers and BPF context, the `req` is the first argument to the instrumented function: `blk_account_io_done()`.
* 기본으로 제공하는 것은  ctx로 register 상태값을 확인할 수 있다.
* 실제 계측되어진 kernel 함수의 첫번째 첫번째 인수이다. 
* `req->__data_len`: Dereferencing that member. __data_len을 역참조 한 값.
* 여기서  `kprobe__blk_account_io_done` 이함수는 오류가 발생한다. 그래서 `blk_account_io_merge_bio` 이 함수로 수정하고 테스트 한다.  


### 해설

1. `BPF_HISTOGRAM(dist)`: 히스토그램 BPF 맵을 생성한다.  Defines a BPF map object that is a histogram, and names it "dist".
2. `dist.increment()`: 히스토그램 BPF 맵에서 해당 키 값으로 되었있는 value를 증가 시킨다.  Increments the histogram bucket index provided as first argument by one by default. Optionally, custom increments can be passed as the second argument.
3. `bpf_log2l()`: log2 값을 출력한다.  Returns the log-2 of the provided value. This becomes the index of our histogram, so that we're constructing a power-of-2 histogram.
4. `b["dist"].print_log2_hist("kbytes")`: Prints the "dist" histogram as power-of-2, with a column header of "kbytes". The only data transferred from kernel to user space is the bucket counts, making this efficient.

## Lesson 10. disklatency.py

디스크 I/O 시간을 측정하고 대기 시간에 대한 히스토그램을 인쇄하는 프로그램을 작성하세요. 디스크 I/O 계측 및 타이밍은 이전 강의의 disksnoop.py 프로그램에서 찾을 수 있으며, 히스토그램 코드는 이전 강의의 bithist.py에서 찾을 수 있습니다.

Write a program that times disk I/O, and prints a histogram of their latency. Disk I/O instrumentation and timing can be found in the disksnoop.py program from a prior lesson, and histogram code can be found in bitehist.py from a prior lesson.

### 문제풀이
```py
#!/usr/bin/python
#
# bitehist.py	Block I/O size histogram.
#		For Linux, uses BCC, eBPF. Embedded C.
#

from __future__ import print_function
from bcc import BPF
from time import sleep

# load BPF program
prog="""
#include <uapi/linux/ptrace.h>
#include <linux/blk-mq.h>

BPF_HISTOGRAM(dist);
BPF_HISTOGRAM(latency);
BPF_HASH(start, struct request *);

void trace_req_start(struct pt_regs *ctx, struct request *req) {
	// stash start timestamp by request ptr
	u64 ts = bpf_ktime_get_ns();
	start.update(&req, &ts);
}

void trace_req_done(struct pt_regs *ctx, struct request *req){
    u64 *tsp, delta;    
   	tsp = start.lookup(&req);
	if (tsp != 0) {
		delta = bpf_ktime_get_ns() - *tsp;		
        latency.increment(bpf_log2l(delta / 1000000));    
		start.delete(&req);
	} 
    dist.increment(bpf_log2l(req->__data_len / 1024));        
}
"""
b = BPF(text=prog)
b.attach_kprobe(event="blk_mq_start_request", fn_name="trace_req_start")
b.attach_kprobe(event="blk_account_io_merge_bio", fn_name="trace_req_done")

# header
print("Tracing... Hit Ctrl-C to end.")

# trace until Ctrl-C
try:
    sleep(99999999)
except KeyboardInterrupt:
    print()

# output
print("log2 histogram")
print("~~~~~~~~~~~~~~")
b["dist"].print_log2_hist("kbytes")
b["latency"].print_log2_hist("ms")
```
* BPF_HASH의 목적은 변수를 bcc에서 시작과 종료 사이에 계측이 필요한 값이 있다면 그것을 저장하기 위한 목적이다.
* 시작과 종료가 있다면 그것의  req* 값은 동일한 것을 사용하기 때문에 좋은 key 후보가 된다. 이렇게 하는 이유는 io 요청이 비동기 적으로 쏟아져 들어오고 그것의 종료 시점이 서로 다르기 때문에  두개의 event가 서로 동기 되지 않은 상태에서 동작하기 때문에 두개를 묶어 줄수 있는 방법이 필요하다. 
* 그리고 struct request 이것이 실제 어떻게 된 구조체 인지 알고 싶은데 잘 알기 어렵다.  


### request 객체 정의
```c
struct xsk_tx_metadata {
	__u64 flags;

	union {
		struct {
			/* XDP_TXMD_FLAGS_CHECKSUM */

			/* Offset from desc->addr where checksumming should start. */
			__u16 csum_start;
			/* Offset from csum_start where checksum should be stored. */
			__u16 csum_offset;
		} request;

		struct {
			/* XDP_TXMD_FLAGS_TIMESTAMP */
			__u64 tx_timestamp;
		} completion;
	};
};
```

## Lesson 11. vfsreadlat.py

This example is split into separate Python and C files. Example output:

```
# ./vfsreadlat.py 1
Tracing... Hit Ctrl-C to end.
     usecs               : count     distribution
         0 -> 1          : 0        |                                        |
         2 -> 3          : 2        |***********                             |
         4 -> 7          : 7        |****************************************|
         8 -> 15         : 4        |**********************                  |

     usecs               : count     distribution
         0 -> 1          : 29       |****************************************|
         2 -> 3          : 28       |**************************************  |
         4 -> 7          : 4        |*****                                   |
         8 -> 15         : 8        |***********                             |
        16 -> 31         : 0        |                                        |
        32 -> 63         : 0        |                                        |
        64 -> 127        : 0        |                                        |
       128 -> 255        : 0        |                                        |
       256 -> 511        : 2        |**                                      |
       512 -> 1023       : 0        |                                        |
      1024 -> 2047       : 0        |                                        |
      2048 -> 4095       : 0        |                                        |
      4096 -> 8191       : 4        |*****                                   |
      8192 -> 16383      : 6        |********                                |
     16384 -> 32767      : 9        |************                            |
     32768 -> 65535      : 6        |********                                |
     65536 -> 131071     : 2        |**                                      |

     usecs               : count     distribution
         0 -> 1          : 11       |****************************************|
         2 -> 3          : 2        |*******                                 |
         4 -> 7          : 10       |************************************    |
         8 -> 15         : 8        |*****************************           |
        16 -> 31         : 1        |***                                     |
        32 -> 63         : 2        |*******                                 |
[...]
```

#### c 프로그램과 Python 프로그램 분리 
* 이렇게 분리되어 있으니 코드 추적도 편하고 좋네..
* 여기서는 attach_kprobe, attach_kretprobe 이렇게 시작과 종료를 나눠서 추적하는 구나...
```py
b = BPF(src_file = "vfsreadlat.c")
b.attach_kprobe(event="vfs_read", fn_name="do_entry")
b.attach_kretprobe(event="vfs_read", fn_name="do_return")
```

Browse the code in [examples/tracing/vfsreadlat.py](../examples/tracing/vfsreadlat.py) and [examples/tracing/vfsreadlat.c](../examples/tracing/vfsreadlat.c). Things to learn:

1. `b = BPF(src_file = "vfsreadlat.c")`: Read the BPF C program from a separate source file.
2. `b.attach_kretprobe(event="vfs_read", fn_name="do_return")`: Attaches the BPF C function ```do_return()``` to the return of the kernel function ```vfs_read()```. This is a kretprobe: instrumenting the return from a function, rather than its entry.
3. `b["dist"].clear()`: Clears the histogram.

## Lesson 12. urandomread.py

Tracing while a ```dd if=/dev/urandom of=/dev/null bs=8k count=5``` is run:

```
# ./urandomread.py
TIME(s)            COMM             PID    GOTBITS
24652832.956994001 smtp             24690  384
24652837.726500999 dd               24692  65536
24652837.727111001 dd               24692  65536
24652837.727703001 dd               24692  65536
24652837.728294998 dd               24692  65536
24652837.728888001 dd               24692  65536
```

Hah! I caught smtp by accident. Code is [examples/tracing/urandomread.py](../examples/tracing/urandomread.py):

```Python
from __future__ import print_function
from bcc import BPF

# load BPF program
b = BPF(text="""
TRACEPOINT_PROBE(random, urandom_read) {
    // args is from /sys/kernel/debug/tracing/events/random/urandom_read/format
    bpf_trace_printk("%d\\n", args->got_bits);
    return 0;
}
""")

# header
print("%-18s %-16s %-6s %s" % ("TIME(s)", "COMM", "PID", "GOTBITS"))

# format output
while 1:
    try:
        (task, pid, cpu, flags, ts, msg) = b.trace_fields()
    except ValueError:
        continue
    print("%-18.9f %-16s %-6d %s" % (ts, task, pid, msg))
```
* `bpf_trace_printk` 함수는 /usr/include/bpf.h에 정의 되어 있음 

### TRACEPOINT_PROBE(category, event)
* catagory:event 방식으로 설정된 trace point를 정의한다.  
* This is a macro that instruments the tracepoint defined by category:event.
* The tracepoint name is <category>:<event>. The probe function name is tracepoint__<category>__<event>.
* Arguments are available in an args struct, which are the tracepoint arguments. 
* One way to list these is to cat the relevant format file under /sys/kernel/debug/tracing/events/category/event/format.
* pytho에서 사용하면 자동으로 이 함수가 attach 된다. 
The args struct can be used in place of ctx in each functions requiring a context as an argument. This includes notably perf_submit().

For example:
```
TRACEPOINT_PROBE(random, urandom_read) {
    // args is from /sys/kernel/debug/tracing/events/random/urandom_read/format
    bpf_trace_printk("%d\\n", args->got_bits);
    return 0;
}
```
This instruments the tracepoint random:urandom_read tracepoint, and prints the tracepoint argument got_bits. When using Python API, this probe is automatically attached to the right tracepoint target. For C++, this tracepoint probe can be attached by specifying the tracepoint target and function name explicitly: BPF::attach_tracepoint("random:urandom_read", "tracepoint__random__urandom_read") Note the name of the probe function defined above is tracepoint__random__urandom_read.

#### error 
* 일단 이 코드는 다음과 같은 에러가 발생한다.  
* 컴파일 에러 발생인데...
```
Exception has occurred: Exception       (note: full exception trace is shown but execution is paused at: _run_module_as_main)
Failed to compile BPF module <text>
  File "/home/jhyunlee/.local/lib/python3.10/site-packages/bcc/__init__.py", line 479, in __init__   raise Exception("Failed to compile BPF module %s" % (src_file or "<text>"))
  File "/home/jhyunlee/code/eBPF/bcc/docs/lesson12.urandomread.py", line 10, in <module>    b = BPF(text="""  File "/usr/lib/python3.10/runpy.py", line 86, in _run_code   exec(code, run_globals)
  File "/usr/lib/python3.10/runpy.py", line 196, in _run_module_as_main (Current frame)    return _run_code(code, main_globals, None,
Exception: Failed to compile BPF module <text>
```

#### random:urandom_read 심볼 찾기
* 없네...
```
root@Good:/sys/kernel/debug/tracing# sudo trace-cmd list | grep random
syscalls:sys_exit_getrandom
syscalls:sys_enter_getrandom

root@Good:/sys/kernel/debug/tracing# grep  random_read avail*
available_filter_functions:random_read_iter
available_filter_functions:urandom_read_iter
```
#### 코드 수정 
* random event에 대해서는 trace event가 없어서 sys_enter_clone으로 위치를 변경해서 테스트 해본다. 

```c
TRACEPOINT_PROBE(syscalls, sys_enter_clone) {
    // args is from /sys/kernel/debug/tracing/events/syscalls/sys_enter_clone/format
    bpf_trace_printk("%d\\n", args->parent_tidptr);
    return 0;
}
```

### 해설

1. `TRACEPOINT_PROBE(random, urandom_read)`: Instrument the kernel tracepoint ```random:urandom_read```. These have a stable API, and thus are recommend to use instead of kprobes, wherever possible. You can run ```perf list``` for a list of tracepoints. Linux >= 4.7 is required to attach BPF programs to tracepoints.
==> 여기서 random:urandom_read는 잘 되어서 다른 event 중에서 하나를 골라서 대체해서 테스트 한다.   `TRACEPOINT_PROBE(syscalls, sys_enter_clone)`
2. `args->parent_tidptr`: `args` 는 마크로 함수를 통해서 자동으로 생성되는 arg이므로 참조할 수 있는데 이것의 format은 `format`을 통해서 확인 할 수 있다.  is auto-populated to be a structure of the tracepoint arguments. The comment above says where you can see that structure. Eg:

```sh
# cat /sys/kernel/debug/tracing/events/syscalls/sys_enter_clone/format
name: sys_enter_clone
ID: 124
format:
	field:unsigned short common_type;	offset:0;	size:2;	signed:0;
	field:unsigned char common_flags;	offset:2;	size:1;	signed:0;
	field:unsigned char common_preempt_count;	offset:3;	size:1;	signed:0;
	field:int common_pid;	offset:4;	size:4;	signed:1;

	field:int __syscall_nr;	offset:8;	size:4;	signed:1;
	field:unsigned long clone_flags;	offset:16;	size:8;	signed:0;
	field:unsigned long newsp;	offset:24;	size:8;	signed:0;
	field:int * parent_tidptr;	offset:32;	size:8;	signed:0;
	field:int * child_tidptr;	offset:40;	size:8;	signed:0;
	field:unsigned long tls;	offset:48;	size:8;	signed:0;

print fmt: "clone_flags: 0x%08lx, newsp: 0x%08lx, parent_tidptr: 0x%08lx, child_tidptr: 0x%08lx, tls: 0x%08lx", ((unsigned long)(REC->clone_flags)), ((unsigned long)(REC->newsp)), ((unsigned long)(REC->parent_tidptr)), ((unsigned long)(REC->child_tidptr)), ((unsigned long)(REC->tls))
```

In this case, we were printing the `parent_tidptr` member.

## Lesson 13. disksnoop.py fixed

Convert disksnoop.py from a previous lesson to use the `block:block_rq_issue` and `block:block_rq_complete` tracepoints.

* 일단 block:block_rq_issue 가 있는지 확인 해본다.  : 

```sh
$ sudo  trace-cmd list | grep block_rq_issue
block:block_rq_issue

root@Good:/sys/kernel/debug/tracing# grep  block_rq_issue  avail*
available_events:block:block_rq_issue

root@Good:/sys/kernel/debug/tracing/events/block/block_rq_issue# ls
enable  filter  format  hist  id  inject  trigger

root@Good:/sys/kernel/debug/tracing# grep  block:block_rq_complete avail*
available_events:block:block_rq_complete

root@Good:/sys/kernel/debug/tracing/events/block/block_rq_complete# ls -l
합계 0
-rw-r----- 1 root root 0  2월 27 21:26 enable
-rw-r----- 1 root root 0  2월 27 21:26 filter
-r--r----- 1 root root 0  2월 27 21:26 format
-r--r----- 1 root root 0  2월 27 21:26 hist
-r--r----- 1 root root 0  2월 27 21:26 id
--w------- 1 root root 0  2월 27 21:26 inject
-rw-r----- 1 root root 0  2월 27 21:26 trigger
```

* block_rq_issue 
```
root@Good:/sys/kernel/debug/tracing/events/block/block_rq_issue# cat format 
name: block_rq_issue
ID: 1231
format:
	field:unsigned short common_type;	offset:0;	size:2;	signed:0;
	field:unsigned char common_flags;	offset:2;	size:1;	signed:0;
	field:unsigned char common_preempt_count;	offset:3;	size:1;	signed:0;
	field:int common_pid;	offset:4;	size:4;	signed:1;

	field:dev_t dev;	offset:8;	size:4;	signed:0;
	field:sector_t sector;	offset:16;	size:8;	signed:0;
	field:unsigned int nr_sector;	offset:24;	size:4;	signed:0;
	field:unsigned int bytes;	offset:28;	size:4;	signed:0;
	field:char rwbs[8];	offset:32;	size:8;	signed:0;
	field:char comm[16];	offset:40;	size:16;	signed:0;
	field:__data_loc char[] cmd;	offset:56;	size:4;	signed:0;

print fmt: "%d,%d %s %u (%s) %llu + %u [%s]", ((unsigned int) ((REC->dev) >> 20)), ((unsigned int) ((REC->dev) & ((1U << 20) - 1))), REC->rwbs, REC->bytes, __get_str(cmd), (unsigned long long)REC->sector, REC->nr_sector, REC->comm


root@Good:/sys/kernel/debug/tracing/events/block/block_rq_complete# cat format 
name: block_rq_complete
ID: 1234
format:
	field:unsigned short common_type;	offset:0;	size:2;	signed:0;
	field:unsigned char common_flags;	offset:2;	size:1;	signed:0;
	field:unsigned char common_preempt_count;	offset:3;	size:1;	signed:0;
	field:int common_pid;	offset:4;	size:4;	signed:1;

	field:dev_t dev;	offset:8;	size:4;	signed:0;
	field:sector_t sector;	offset:16;	size:8;	signed:0;
	field:unsigned int nr_sector;	offset:24;	size:4;	signed:0;
	field:int error;	offset:28;	size:4;	signed:1;
	field:char rwbs[8];	offset:32;	size:8;	signed:0;
	field:__data_loc char[] cmd;	offset:40;	size:4;	signed:0;

print fmt: "%d,%d %s (%s) %llu + %u [%d]", ((unsigned int) ((REC->dev) >> 20)), ((unsigned int) ((REC->dev) & ((1U << 20) - 1))), REC->rwbs, __get_str(cmd), (unsigned long long)REC->sector, REC->nr_sector, REC->error
```

* error 
```
/virtual/main.c:33:16: warning: incompatible pointer types passing 'struct tracepoint__bokck__block_rq_complete *' to parameter of type 'struct request **' [-Wincompatible-pointer-types]
                start.delete(args);
                             ^~~~
```
==> 원인은 이 macro TRACEPOINT_PROBE(block,block_rq_complete)의 리턴 값이  'struct tracepoint__bokck__block_rq_complete *'으로 정의 되어 있는데...  이것을  BPF_HASH를 선언할때 BPF_HASH(start,struct request *); 이렇게 해서 그렇다. 그래서 이것을 맞춰 주면 된다.  

### 문제풀이
* format 파일에서 정의한 모든 내용이 나오는 것은 아니다. 
```py
#!/usr/bin/python

from __future__ import print_function
from bcc import BPF
from bcc.utils import printb

REQ_WRITE = 1		# from include/linux/blk_types.h

# load BPF program
prog="""
#include <uapi/linux/ptrace.h>
#include <linux/blk-mq.h>

BPF_HASH(start, int);

TRACEPOINT_PROBE(block,block_rq_issue) {	
	u64 ts = bpf_ktime_get_ns();	
    u32 dev = args->dev;
    u64 sector = args->sector;
    u32 nr_sector = args->nr_sector;
    bpf_trace_printk("rq_issue [%u] [%llu] [%u]\\n",dev, sector, nr_sector);
	return 0;
}

TRACEPOINT_PROBE(block,block_rq_complete) {
	u64 *tsp, delta;	
    u32 dev = args->dev;
    u64 sector = args->sector;
    u32 nr_sector = args->nr_sector;
    bpf_trace_printk("rq_complete [%u] [%llu] [%u]\\n",dev, sector, nr_sector);    
	return 0;
}
"""
b=BPF(text=prog)
# header
print("%-18s %-2s %-7s %8s" % ("TIME(s)", "T", "BYTES", "LAT(ms)"))

b.trace_print()

# format output
while 1:
	try:
		(id, dev, sector, nr_sector) = b.trace_fields()		
	
		print(id, dev, sector, nr_sector)
	except KeyboardInterrupt:
		exit()
```


## Lesson 14. strlen_count.py
* user level의 strlen 함수 라이브러리를 얼마나 호출했는지를 확인하는 방법
* user level의 심볼을 참조해서 처리하려고하는 handler를 연결하는 방법으로 구현한다.  


This program instruments a user-level function, the ```strlen()``` library function, and frequency counts its string argument. Example output:

```
# ./strlen_count.py
Tracing strlen()... Hit Ctrl-C to end.
^C     COUNT STRING
         1 " "
         1 "/bin/ls"
         1 "."
         1 "cpudist.py.1"
         1 ".bashrc"
         1 "ls --color=auto"
         1 "key_t"
[...]
        10 "a7:~# "
        10 "/root"
        12 "LC_ALL"
        12 "en_US.UTF-8"
        13 "en_US.UTF-8"
        20 "~"
        70 "#%^,~:-=?+/}"
       340 "\x01\x1b]0;root@bgregg-test: ~\x07\x02root@bgregg-test:~# "
```

These are various strings that are being processed by this library function while tracing, along with their frequency counts. `strlen()` was called on "LC_ALL" 12 times, for example.

Code is [examples/tracing/strlen_count.py](../examples/tracing/strlen_count.py):

```Python
from __future__ import print_function
from bcc import BPF
from time import sleep

# load BPF program
b = BPF(text="""
#include <uapi/linux/ptrace.h>

struct key_t {
    char c[80];
};
BPF_HASH(counts, struct key_t);

int count(struct pt_regs *ctx) {
    if (!PT_REGS_PARM1(ctx))
        return 0;

    struct key_t key = {};
    u64 zero = 0, *val;

    bpf_probe_read_user(&key.c, sizeof(key.c), (void *)PT_REGS_PARM1(ctx));
    // could also use `counts.increment(key)`
    val = counts.lookup_or_try_init(&key, &zero);
    if (val) {
      (*val)++;
    }
    return 0;
};
""")
b.attach_uprobe(name="c", sym="strlen", fn_name="count")

# header
print("Tracing strlen()... Hit Ctrl-C to end.")

# sleep until Ctrl-C
try:
    sleep(99999999)
except KeyboardInterrupt:
    pass

# print output
print("%10s %s" % ("COUNT", "STRING"))
counts = b.get_table("counts")
for k, v in sorted(counts.items(), key=lambda counts: counts[1].value):
    print("%10d \"%s\"" % (v.value, k.c.encode('string-escape')))
```

### 해설

1. `PT_REGS_PARM1(ctx)`: This fetches the first argument to `strlen()`, which is the string.
2. `bpf_probe_read_user(&key.c, sizeof(key.c), (void *)PT_REGS_PARM1(ctx))`:  user 주소 공간에 있는 데이터를 BPF stack으로 NULL로 끝나는 문자열로 복사해 온다. 그래서 BPF에서는 user 공간의 데이터를 사용할 수 있게 된다. 만약 복사하려는 문자열이 size 보다 크면 len-1 크기로 복사하고 NULL로 끝나는 문자열을 복사한다.  문자열이 복사하려는 크기보다 작은 경우는 target은 NULL로 채워 지지 않는다. 이렇게 코딩되어 있어서 그런가 보다. 
```
void strcpy(char *s, char *t){
    while(*s++){
        *t=*s;
    }
}
```
3. `val = counts.lookup_or_try_init(&key, &zero);` 생성된 map인 count 객체에 &key 값이 있으면 리턴해주고, 없으면  두번째 변수로 초기화 한다.  
4. `b.attach_uprobe(name="c", sym="strlen", fn_name="count")`: Attach to library "c" (if this is the main program, use its pathname), instrument the user-level function `strlen()`, and on execution call our C function `count()`.


#### PT_REGS_PARM macro

attach_uprobe를 통해서 attach된 것은 매개변수로 struct pt_regs가 들어오고 이렇게 들어온 user probe 된 매개 변수를 체크하는 것은 PT_REGS_PARM macro를 통해서 가능하다. 

```
int count(struct pt_regs *ctx) {
    char buf[64];
    bpf_probe_read_user(&buf, sizeof(buf), (void *)PT_REGS_PARM1(ctx));
    bpf_trace_printk("%s %d", buf, PT_REGS_PARM2(ctx));
    return(0);
}
```




## Lesson 15. nodejs_http_server.py

This program instruments a user statically-defined tracing (USDT) probe, which is the user-level version of a kernel tracepoint. Sample output:

```
# ./nodejs_http_server.py 24728
TIME(s)            COMM             PID    ARGS
24653324.561322998 node             24728  path:/index.html
24653335.343401998 node             24728  path:/images/welcome.png
24653340.510164998 node             24728  path:/images/favicon.png
```

Relevant code from [examples/tracing/nodejs_http_server.py](../examples/tracing/nodejs_http_server.py):

```Python
from __future__ import print_function
from bcc import BPF, USDT
import sys

if len(sys.argv) < 2:
    print("USAGE: nodejs_http_server PID")
    exit()
pid = sys.argv[1]
debug = 0

# load BPF program
bpf_text = """
#include <uapi/linux/ptrace.h>
int do_trace(struct pt_regs *ctx) {
    uint64_t addr;
    char path[128]={0};
    bpf_usdt_readarg(6, ctx, &addr);
    bpf_probe_read_user(&path, sizeof(path), (void *)addr);
    bpf_trace_printk("path:%s\\n", path);
    return 0;
};
"""

# enable USDT probe from given PID
u = USDT(pid=int(pid))
u.enable_probe(probe="http__server__request", fn_name="do_trace")
if debug:
    print(u.get_text())
    print(bpf_text)

# initialize BPF
b = BPF(text=bpf_text, usdt_contexts=[u])
```

Things to learn:

1. `bpf_usdt_readarg(6, ctx, &addr)`: Read the address of argument 6 from the USDT probe into ```addr```.
2. `bpf_probe_read_user(&path, sizeof(path), (void *)addr)`: Now the string `addr` points to into our `path` variable.
3. `u = USDT(pid=int(pid))`: Initialize USDT tracing for the given PID.
4. `u.enable_probe(probe="http__server__request", fn_name="do_trace")`: Attach our `do_trace()` BPF C function to the Node.js `http__server__request` USDT probe.
5. `b = BPF(text=bpf_text, usdt_contexts=[u])`: Need to pass in our USDT object, `u`, to BPF object creation.


* 뭔가 리소스를 식별하고 event에 handler를 붙이는 작업을 하는 그런 공통 과정 

## Lesson 16. task_switch.c

This is an older tutorial included as a bonus lesson. Use this for recap and to reinforce what you've already learned.

This is a slightly more complex tracing example than Hello World. This program
will be invoked for every task change in the kernel, and record in a BPF map
the new and old pids.

The C program below introduces a new concept: the prev argument. This
argument is treated specially by the BCC frontend, such that accesses
to this variable are read from the saved context that is passed by the
kprobe infrastructure. The prototype of the args starting from
position 1 should match the prototype of the kernel function being
kprobed. If done so, the program will have seamless access to the
function parameters.

```c
#include <uapi/linux/ptrace.h>
#include <linux/sched.h>

struct key_t {
    u32 prev_pid;
    u32 curr_pid;
};

BPF_HASH(stats, struct key_t, u64, 1024);
int count_sched(struct pt_regs *ctx, struct task_struct *prev) {
    struct key_t key = {};
    u64 zero = 0, *val;

    key.curr_pid = bpf_get_current_pid_tgid();
    key.prev_pid = prev->pid;

    // could also use `stats.increment(key);`
    val = stats.lookup_or_try_init(&key, &zero);
    if (val) {
      (*val)++;
    }
    return 0;
}
```

The userspace component loads the file shown above, and attaches it to the
`finish_task_switch` kernel function.
The `[]` operator of the BPF object gives access to each BPF_HASH in the
program, allowing pass-through access to the values residing in the kernel. Use
the object as you would any other python dict object: read, update, and deletes
are all allowed.
```python
from bcc import BPF
from time import sleep

b = BPF(src_file="task_switch.c")
b.attach_kprobe(event="finish_task_switch", fn_name="count_sched")

# generate many schedule events
for i in range(0, 100): sleep(0.01)

for k, v in b["stats"].items():
    print("task_switch[%5d->%5d]=%u" % (k.prev_pid, k.curr_pid, v.value))
```

These programs can be found in the files [examples/tracing/task_switch.c](../examples/tracing/task_switch.c) and [examples/tracing/task_switch.py](../examples/tracing/task_switch.py) respectively.


#### event 확인 및 프로그램 수정 
* 일단 b.attach_kprobe(event="finish_task_switch.isra.0", fn_name="count_sched") 함수에 event로 적용하려면 정확한 kernel 함수를 찾아서 설정해 줘야 한다. 
* 그래서 get_syscall_fnname 처럼 명시적으로 kernel 함수 찾는 방법을 사용하여 envent 함수를 지정하였다.  b.attach_kprobe(event=b.get_syscall_fnname("sync"), fn_name="do_trace") 
* 다행히 `finish_task_switch.isra.0` 함수 이름으로 존재한다. 그래서 이것을 event 이름 지정해주면 된다. 

```
root@Good:/sys/kernel/debug/tracing# grep  inish_task_switch avail*
available_filter_functions:finish_task_switch.isra.0
available_filter_functions_addrs:ffffffff87944ce0 finish_task_switch.isra.0
```

#### TRACEPOINT_PROBE(category, event) 방식
* 사용할 수 있는 TRACEPOINT는  sched:sched_migrate_task를 적용해자 
```
root@Good:/sys/kernel/debug/tracing# grep  task  available_events
task:task_rename
task:task_newtask
irq:tasklet_exit
irq:tasklet_entry
sched:sched_wait_task
sched:sched_migrate_task
```

* 이 프로그램을 보다 간략하게 수정하려면  TRACEPOINT 방식으로 수정을 할수 있다. 
* 그렇게 하려면 event * catagory:event 방식으로 설정된 trace point를 정의한다. 
* This is a macro that instruments the tracepoint defined by category:event.
* The tracepoint name is <category>:<event>. The probe function name is tracepoint__<category>__<event>.
* 이것은 `trace-cmd list -l ` 방법으로 확인할 수 있다. 
* 그리고 이것이 전달하는 변수는 args로 넘어 오게 되는데 이것은  
* `cat /sys/kernel/debug/tracing/events/sched/sched_switch/format`에서 그 전달 argument를 구조와 형식을 확인할 수 있다.
* 커널 내부 함수를 지정해서 하는 것은 함수가 변경되거나 할 경우 trace가 실패될 가능성도 있기 때문에 category:event 방식으로 지정해서 처리하는 것이 효율적이다.  

```py
#!/usr/bin/python

from bcc import BPF
from time import sleep

prog="""
#include <uapi/linux/ptrace.h>
#include <linux/sched.h>

struct key_t {
    u32 prev_pid;
    u32 curr_pid;
};

BPF_HASH(stats, struct key_t, u64, 1024);

TRACEPOINT_PROBE(sched,sched_switch) {
    struct key_t key = {};
    u64 zero = 0, *val;

    key.curr_pid = bpf_get_current_pid_tgid();
    key.prev_pid = (u32)args->prev_pid;
    
    val = stats.lookup_or_try_init(&key, &zero);
    if (val) {
        (*val)++;
    }
    return 0;
}
"""
b = BPF(text=prog)
for i in range(0, 100): sleep(0.1)
for k, v in b["stats"].items():
    print("task_switch[%5d->%5d]=%u" % (k.prev_pid, k.curr_pid, v.value))
```


## Lesson 17. Further Study

For further study, see Sasha Goldshtein's [linux-tracing-workshop](https://github.com/goldshtn/linux-tracing-workshop), which contains additional labs. There are also many tools in bcc /tools to study.

Please read [CONTRIBUTING-SCRIPTS.md](../CONTRIBUTING-SCRIPTS.md) if you wish to contribute tools to bcc. At the bottom of the main [README.md](../README.md), you'll also find methods for contacting us. Good luck, and happy tracing!



## tool 
### kprobe event 어떻게 발견하는가?
* kprobe 
```sh
$ sudo apt install linux-tools-common
$ sudo apt install linux-tools-generic
$ sudo apt install linux-tools-6.5.0-18-generic
$ sudo apt install trace-cmd
$ sudo trace-cmd list -l | grep  exec
workqueue:workqueue_execute_end
workqueue:workqueue_execute_start
sched:sched_process_exec
sched:sched_kthread_work_execute_end
sched:sched_kthread_work_execute_start
syscalls:sys_exit_kexec_load
syscalls:sys_enter_kexec_load
syscalls:sys_exit_kexec_file_load
syscalls:sys_enter_kexec_file_load
syscalls:sys_exit_execveat
syscalls:sys_enter_execveat
syscalls:sys_exit_execve
syscalls:sys_enter_execve
writeback:writeback_exec
libata:ata_exec_command
```
### ftrace 가능한 커널 함수 목록
```sh
root@Good:/sys/kernel/debug/tracing# grep  blk_account_io  avail*
available_filter_functions:blk_account_io_merge_bio
available_filter_functions:blk_account_io_completion.part.0
available_filter_functions_addrs:ffffffff87f77a90 blk_account_io_merge_bio
available_filter_functions_addrs:ffffffff87f7b5b0 blk_account_io_completion.part.0
```

### trace argement format 
*  /sys/kernel/debug/tracing/events/random/urandom_read/format 디렉토리에서  format 파일을 통해 argument format를 찾을 수 있다.  

```c
// from /sys/kernel/debug/tracing/events/random/urandom_read/format
struct urandom_read_args {    
    u64 __unused__;
    u32 got_bits;
    u32 pool_left;
    u32 input_left;
};
```

### trace-cmd
```
$ sudo trace-cmd list | grep bio
block:block_bio_remap
block:block_bio_queue
block:block_bio_frontmerge
block:block_bio_backmerge
block:block_bio_bounce
block:block_bio_complete
```

### SYSCALL 목록 확인
* 현재 kernel의 system call 목록과 이름 확인 
```
$ grep   __SYSCALL /usr/include/asm-generic/unistd.h
$ grep   clone  /usr/include/asm-generic/unistd.h

#define __NR_clone 220
__SYSCALL(__NR_clone, sys_clone)
#define __NR_clone3 435
__SYSCALL(__NR_clone3, sys_clone3)

```
### 커널 심볼 
```
$ cat /proc/kallsyms | grep blk_account_io_done
```

### iotop 
```
source code
```

### BPF_HASH
* 제일 궁금한 것은 BPF_HASH(counter_table) 이것이 어디에 정의 되어 있는가?
* BPF_HASH() is a BCC macro that defines a hash table map. 라고 하는데 어디에 정의되어 있는지를 모르겠네..
* 무슨 소스 코드가 이렇게 되어 있냐?
  - You can navigate to the src/cc directory and find the bpf_helpers.h file where the BPF_HASH() macro is defined
  - The source code for the BPF_HASH() macro in BCC (BPF Compiler Collection) can be found in the BCC GitHub repository. 
  - BCC is an open-source project, and its source code is hosted on GitHub. 
  - You can find the definition of the BPF_HASH() macro in the bpf_helpers.h header file within the BCC repository.
  - 이것이 macro 인데 실제 파일에 가서 보면  
* bcc repository에서  소스 코드가 이렇게 되어 있는 것은 무엇을 의미하냐 ?  R"********(  
이런 사연이 있었구만 ...

소스 코드가 `R"********(`와 같은 형태로 시작되는 것은 C++11부터 도입된 Raw String Literal 문법을 나타냅니다. 이 문법을 사용하면 문자열을 이스케이프 문자 없이 그대로 표현할 수 있습니다. "********"는 임의의 종료 문자열로, 소스 코드 내에서 나오는 문자열이 이 문자열로 끝나는 것을 나타냅니다.

`bcc/src/export/helpers.h` 에 정의된 내용 을 보면 BPF_F_TABLE macro로 정의한 것을 사용한다. 

```c
R"********(

#define BPF_F_TABLE(_table_type, _key_type, _leaf_type, _name, _max_entries, _flags) \
struct _name##_table_t { \
  _key_type key; \
  _leaf_type leaf; \
  _leaf_type * (*lookup) (_key_type *); \
  _leaf_type * (*lookup_or_init) (_key_type *, _leaf_type *); \
  _leaf_type * (*lookup_or_try_init) (_key_type *, _leaf_type *); \
  int (*update) (_key_type *, _leaf_type *); \
  int (*insert) (_key_type *, _leaf_type *); \
  int (*delete) (_key_type *); \
  void (*call) (void *, int index); \
  void (*increment) (_key_type, ...); \
  void (*atomic_increment) (_key_type, ...); \
  int (*get_stackid) (void *, u64); \
  void * (*sk_storage_get) (void *, void *, int); \
  int (*sk_storage_delete) (void *); \
  void * (*inode_storage_get) (void *, void *, int); \
  int (*inode_storage_delete) (void *); \
  void * (*task_storage_get) (void *, void *, int); \
  int (*task_storage_delete) (void *); \
  u32 max_entries; \
  int flags; \
}; \
__attribute__((section("maps/" _table_type))) \
struct _name##_table_t _name = { .flags = (_flags), .max_entries = (_max_entries) }; \
BPF_ANNOTATE_KV_PAIR(_name, _key_type, _leaf_type)



#define BPF_TABLE(_table_type, _key_type, _leaf_type, _name, _max_entries) \
BPF_F_TABLE(_table_type, _key_type, _leaf_type, _name, _max_entries, 0)


#define BPF_HASH1(_name) \
  BPF_TABLE("hash", u64, u64, _name, 10240)
#define BPF_HASH2(_name, _key_type) \
  BPF_TABLE("hash", _key_type, u64, _name, 10240)
#define BPF_HASH3(_name, _key_type, _leaf_type) \
  BPF_TABLE("hash", _key_type, _leaf_type, _name, 10240)
#define BPF_HASH4(_name, _key_type, _leaf_type, _size) \
  BPF_TABLE("hash", _key_type, _leaf_type, _name, _size)

// helper for default-variable macro function
#define BPF_HASHX(_1, _2, _3, _4, NAME, ...) NAME


// Define a hash function, some arguments optional
// BPF_HASH(name, key_type=u64, leaf_type=u64, size=10240)
#define BPF_HASH(...) \
  BPF_HASHX(__VA_ARGS__, BPF_HASH4, BPF_HASH3, BPF_HASH2, BPF_HASH1)(__VA_ARGS__)

```



##  kprobe  and  ftrace (/sys/kernel/debug/tracing) 차이점 

When learning about the Linux kernel, understanding the difference between Kprobes and Ftrace can be crucial, as they are both tools used for kernel debugging and tracing, but they serve different purposes and operate at different levels of the kernel.

Kprobes: Kprobes is a dynamic kernel debugging mechanism that allows developers to insert breakpoints (probes) into running kernel code. These probes can be used to monitor the execution flow of the kernel, gather information about specific events, or debug kernel code without requiring recompilation or rebooting the system. Kprobes allows developers to attach "probe handlers" to specific locations in the kernel code, which are executed when the probe is hit. This mechanism is particularly useful for debugging complex kernel issues or analyzing kernel behavior in real-time.

Ftrace: Ftrace, on the other hand, is a kernel tracing framework that provides a set of tools for tracing various kernel events and functions. It allows developers to dynamically instrument the kernel to collect detailed information about its behavior, such as function call traces, context switches, interrupt activity, and more. Ftrace provides a powerful interface for analyzing kernel performance, identifying bottlenecks, and diagnosing issues. It consists of several components, including function tracer, function graph tracer, event tracer, and tracepoints. Ftrace is typically used for performance analysis, optimization, and understanding kernel internals.

Here's a summary of the key differences between Kprobes and Ftrace:


#### Purpose: 
* kprobe : 커널 코드에 대한 동적 디버깅하기 위해서 handler 커널에 코드를 집어 넣는 것
* ftrace : 커널 활동을 tracing 하고 성능을 분석하기위한 것
* Kprobes is primarily used for dynamic kernel debugging by inserting probes into running kernel code to monitor specific events or gather information
* Ftrace is used for tracing kernel activities and performance analysis.

#### Granularity: 
* kprobe : instruction level에서 커널의 동작을 확인하기 위해서 디거깅용 코그를 넣기 때문에 더 작은 단위
* ftrace : 좀더 high level에서  system  call, event 등을 tracing 하는 것 
* Kprobes operates at the instruction level, allowing developers to insert probes at specific locations within kernel code
* Ftrace operates at a higher level, tracing function calls, events, and system activities.

#### Flexibility: 
* kprobe : 커널 코드에 대한 세밀한 제어를 할 수 있다. 
* ftrace : 추적 기능과 분석 도구가 내장된 framework 
* Kprobes provides fine-grained control over the instrumentation of kernel code and allows developers to specify custom probe handlers
* Ftrace provides a more high-level tracing framework with built-in tracing capabilities and analysis tools.

#### Use Cases: 
* kprobe : 커널의 특정한 이슈에 대한 디버깅, 커널 개발자가 사용하는 도구
* ftrace :  성능 분석 및 최적화 커널의 전반적 동작을 이해하는데 사용 
* Kprobes is typically used for debugging specific kernel issues or analyzing kernel behavior in real-time
* Ftrace is used for performance analysis, optimization, and understanding the overall behavior of the kernel.

In summary, while both Kprobes and Ftrace are powerful tools for kernel debugging and tracing, they serve different purposes and offer different levels of granularity and flexibility. Developers may choose to use one or both of these tools depending on their specific debugging and tracing requirements.


## 결론 
* 커널 수준의 개발자가 이슈 디버깅을 위해서는 kprobe를 사용하는 것이 맞고 
    - BPF: kbpobes
    - BPF: kretprobe
* 성능 분석 및 모니터링 정도를 하려면 ftrace를 이용한 eBPF를 사용하는 것이 맞다.  
    - BPF: tracepoint 
```
TRACEPOINT_PROBE(random, urandom_read) {
    // args is from /sys/kernel/debug/tracing/events/random/urandom_read/format
    bpf_trace_printk("%d\\n", args->got_bits);
    return 0;
}
```

## ftrace에서 함수호출, event 추적

### 1.available_events 
* 커널 내에서 추적에 사용할 수 있는 이벤트를 나타냅니다. 
* 이러한 이벤트는 다양한 함수 호출, 스케줄러 이벤트, 인터럽트 또는 커널 내의 기타 추적 가능한 활동일 수 있습니다. 
* available_events이러한 이벤트는 일반적으로 추적 디렉터리( )의 파일 에 나열됩니다 /sys/kernel/debug/tracing/. 
* 런타임 중에 발생하는 이벤트에 대한 정보를 수집하기 위해 이러한 이벤트에 대한 추적을 활성화할 수 있습니다.
```
root@Good:/sys/kernel/debug/tracing# cat available_events  | grep openat
syscalls:sys_exit_openat2
syscalls:sys_enter_openat2
```

### 2. available_filter_functions
* 추적 데이터를 필터링하는 데 사용할 수 있는 커널 내의 함수입니다. 
* 이러한 기능은 추적 범위를 커널 내의 특정 관심 영역으로 좁히는 데 도움이 될 수 있습니다. 
* available_filter_functions추적 디렉터리의 파일 에 나열되는 경우가 많습니다 . 이러한 함수를 사용하여 함수 이름, 모듈 이름 또는 기타 속성과 같은 특정 기준에 따라 이벤트를 필터링할 수 있습니다.
```
root@Good:/sys/kernel/debug/tracing# cat available_filter_functions | grep openat2
__audit_openat2_how
do_sys_openat2
__x64_sys_openat2
__ia32_sys_openat2
io_openat2_prep
io_openat2
```

### 3. available_filter_functions_addrs : 
*available_filter_functions와 유사하지만 함수 이름을 나열하는 대신 커널 내 함수의 주소를 제공합니다. 이는 이름이 아닌 함수 주소를 기준으로 필터링해야 하는 경우 유용할 수 있습니다.
```
root@Good:/sys/kernel/debug/tracing# cat available_filter_functions_addrs | grep openat2
ffffffff87a46340 __audit_openat2_how
ffffffff87ca95c0 do_sys_openat2
ffffffff87ca97e0 __x64_sys_openat2
ffffffff87ca9820 __ia32_sys_openat2
ffffffff87fd3c10 io_openat2_prep
ffffffff87fd3cb0 io_openat2
```
### 4. available_tracers : 
* 파일 available_tracers에는 커널에서 사용할 수 있는 추적 프로그램이 나열되어 있습니다. 
* 추적 프로그램의 예로는 함수 추적 프로그램, 함수 그래프 추적 프로그램, 이벤트 추적 프로그램 등이 있습니다.
* 추적 프로그램은 ftrace커널 내의 특정 이벤트 또는 함수 호출에 대한 추적 데이터를 캡처할 수 있는 메커니즘입니다. 추적 프로그램마다 기능과 오버헤드가 다릅니다. 
```
root@Good:/sys/kernel/debug/tracing# cat available_tracers 
timerlat osnoise hwlat blk mmiotrace function_graph wakeup_dl wakeup_rt wakeup function no
```


## tracing with ftrace

### tracer 설정
ftrace는 nop, function, graph_function 트레이서를 제공합니다.
* nop: 기본 트레이서입니다. ftrace 이벤트만 출력합니다.**
* function: 함수 트레이서입니다. set_ftrace_filter로 지정한 함수를 누가 호출하는지 출력합니다.**
* graph_function: 함수 실행 시간과 세부 호출 정보를 그래프 포맷으로 출력합니다.**
```
root@raspberrypi:/sys/kernel/debug/tracing# cat current_tracer
nop
```
#### event trace.sh
```sh
#!/bin/bash
echo 0 > /sys/kernel/debug/tracing/tracing_on
echo 0 > /sys/kernel/debug/tracing/events/enable
echo function > /sys/kernel/debug/tracing/current_tracer
echo 1 > /sys/kernel/debug/tracing/events/sched/sched_wakeup/enable
echo 1 > /sys/kernel/debug/tracing/events/sched/sched_switch/enable
echo 1 > /sys/kernel/debug/tracing/events/irq/irq_handler_entry/enable
echo 1 > /sys/kernel/debug/tracing/events/irq/irq_handler_exit/enable
echo 1 > /sys/kernel/debug/tracing/events/raw_syscalls/enable
echo 1 > /sys/kernel/debug/tracing/options/func_stack_trace
echo 1 > /sys/kernel/debug/tracing/options/sym-offset
echo 1 > /sys/kernel/debug/tracing/tracing_on
```
###  set_ftrace_filter 설정
* set_ftrace_filter 파일에 트레이싱하고 싶은 함수를 지정하면 된다.
* 위의 tracer 설정의 function 혹은 function_graph으로 설정한 경우 작동하는 파일이다.
* 리눅스 커널에 존재하는 모든 함수를 필터로 지정할 수는 없다.
* /sys/kernel/debug/tracing/available_filter_functions 파일에 포함된 함수만 지정할 수 있다.
* 함수를 지정하지 않은 경우 모든 함수를 트레이싱하게 되어 락업이 상태에 빠지게 된다.
* available_filter_functions 파일에 없는 함수를 지정하려도 락업 상태가 될 수 있으니 주의하자.
* set_ftrace_filter에 아무것도 설정하지 않고 ftrace를 키면, ftrace는 모든 커널 함수에 대하여 트레이싱을 한다.
* 모든 커널 함수에 의해 트레이스가 발생되면, 그 오버헤드가 엄청나 시스템은 락업 상태에 빠진다.
* 그러므로 부팅 이후 절대 불리지 않을 함수secondary_start_kernel2를 트레이스 포인트로 찍어준다.

### kernel함수 trace (file open,read,write,close)
```sh
#!/bin/bash
echo 0 > /sys/kernel/debug/tracing/tracing_on
echo 0 > /sys/kernel/debug/tracing/events/enable
echo function > /sys/kernel/debug/tracing/current_tracer
echo do_sys_openat2  > /sys/kernel/debug/tracing/set_ftrace_filter
echo ksys_read   >> /sys/kernel/debug/tracing/set_ftrace_filter
echo ksys_write  >> /sys/kernel/debug/tracing/set_ftrace_filter
echo close_fd  >> /sys/kernel/debug/tracing/set_ftrace_filter
echo 1 > /sys/kernel/debug/tracing/options/func_stack_trace
echo 1 > /sys/kernel/debug/tracing/options/sym-offset
echo 1 > /sys/kernel/debug/tracing/tracing_on
```
```sh
#!/bin/bash
echo 0 > /sys/kernel/debug/tracing/tracing_on
echo 0 > /sys/kernel/debug/tracing/events/enable
echo 0 > /sys/kernel/debug/tracing/options/stacktrace
cp  /sys/kernel/debug/tracing/trace ftrace.log
```




#### trace-cmd
* interacts with ftrace linuc kernel internal tracer
* ftrace front utility

```c
# apt  install trace-cmd
# trace-cmd  record -p function ./hello
# trace-cmd  record -p function ./hello  
# trace-cmd  record -p function-graph ./hello  
# trace-cmd  record -p function ./hello  
# trace-cmd  repoort >t.log  
```

### perf stat
```log
root@gpu-1:~# perf stat ./hello
 Performance counter stats for './hello':
              0.49 msec task-clock                       #    0.541 CPUs utilized            
                 0      context-switches                 #    0.000 /sec                      
                 0      cpu-migrations                   #    0.000 /sec                      
                63      page-faults                      #  127.476 K/sec                    
         2,048,928      cycles                           #    4.146 GHz                      
         1,376,140      instructions                     #    0.67  insn per cycle            
           245,301      branches                         #  496.350 M/sec                    
             8,908      branch-misses                    #    3.63% of all branches          
       0.000914125 seconds time elapsed
       0.000985000 seconds user
       0.000000000 seconds sys
```

## tracing with uftrace
```sh
$ gdb  ./hello
(gdb) list
(gdb) break 5
(gdb) run
(gdb) info frame
(gdb) info files
(gdb) info local
(gdb) info proc
(gdb) info break
(gdb) print VAL
(gdb) display i
(gdb) disas main
$ stat ./hello
$ perf record -a -g  ./hello
$ perf report --header  -F overhead,comm,parent
$ perf stat ./hello
$ strace ./hello
$ stat  ./hello
$ sudo uftrace -K 5 ./hello
$ sudo uftrace record -K 5 ./hello
$ sudo uftrace tui
```