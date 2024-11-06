# perf,ftrace,kbrobe

## perf와 ftrace 차이점 비교
ChatGPT
perf와 ftrace는 모두 리눅스 커널에서 성능 분석 및 추적을 위한 도구입니다. 하지만 각각은 다음과 같은 차이점이 있습니다:

### 1. 목적:
* perf: perf는 퍼포먼스 분석 및 프로파일링 도구로 널리 사용됩니다. CPU 사용률, 메모리 사용량, 함수 호출 등과 같은 다양한 성능 지표를 수집하고 분석하는 데 사용됩니다.
* ftrace: ftrace는 리눅스 커널 내에서 발생하는 이벤트를 추적하기 위한 도구로, 커널 함수 호출, 인터럽트, 스케줄링 이벤트 등을 추적할 수 있습니다. 보통 커널의 동작을 디버깅하거나 커널 내부 동작을 이해하기 위해 사용됩니다.

### 2. 수집 대상:
* perf: 주로 유저 공간에서 발생하는 이벤트를 추적하며, CPU 성능 분석에 특히 유용합니다. 프로세스 간의 이벤트 추적, 시스템 전체의 프로파일링, 커널 함수 실행 시간 등을 측정할 수 있습니다.
* ftrace: 주로 커널 내에서 발생하는 이벤트를 추적하며, 커널 함수 호출, 인터럽트, 스케줄링 이벤트 등을 추적할 수 있습니다. 이는 커널 내부의 동작을 이해하거나 디버깅할 때 유용합니다.

### 3. 사용 방법:
* perf: perf 명령을 사용하여 성능 측정 및 프로파일링을 수행할 수 있습니다. 사용자 공간에서 실행되는 프로세스의 프로파일링, CPU 이벤트의 추적, 각종 리눅스 시스템 리소스 모니터링에 사용됩니다.
* ftrace: trace-cmd 명령을 사용하여 ftrace를 활성화하고 이벤트를 추적합니다. 커널 내부의 이벤트를 추적하기 때문에 일반적으로 시스템 관리자나 커널 개발자들이 사용합니다.

### 4. 정확성 및 오버헤드:
* perf: 성능 측정을 위한 높은 정확성을 제공하지만, 이를 위해 상당한 시스템 리소스를 소모합니다.
* ftrace: 커널 내부에서 동작하기 때문에 일반적으로 perf보다 더 낮은 오버헤드를 가집니다. 그러나 정확성 측면에서는 일부 제약이 있을 수 있습니다.

### 요약
* perf는 유저 공간에서의 성능 측정 및 프로파일링에 주로 사용되는 반면, 
* ftrace는 주로 커널 내부의 동작을 추적하고 이해하는 데 사용됩니다. 
* perf는 높은 정확성을 제공하지만 오버헤드가 크고, 
* ftrace는 상대적으로 낮은 오버헤드를 가지지만 정확성 측면에서는 일부 제약이 있을 수 있습니다.

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

## available_events  available_filter_functions

#### 1. available_events 
1. event enable은 perf 유틸리티를 사용하여 활성화되는 리눅스 커널 이벤트입니다.
2. perf를 사용하여 다양한 이벤트를 모니터링하고 분석할 수 있습니다.
3. 이벤트 활성화를 사용하여 프로파일링이나 성능 모니터링과 같은 작업을 수행할 수 있습니다.
4. 주로 성능 최적화, 시스템 모니터링 및 디버깅에 사용됩니다.

#### 2. set_ftrace_filter
1. /sys/kernel/debug/tracing 경로에 위치한 파일 중 하나로, ftrace의 필터를 설정하는 데 사용됩니다.
2. ftrace 필터는 커널의 특정 이벤트를 추적하고 관련된 정보만 보여주는 데 사용됩니다.
3. 일반적으로 BPF (Berkley Packet Filter)와 같은 도구와 함께 사용하여 리눅스 커널에서 실행 중인 프로세스의 동작을 추적하거나 디버깅하는 데 유용합니다.
4. 필터는 특정 이벤트, 프로세스, 함수 등을 추적하도록 설정할 수 있습니다.

요약하면, /sys/kernel/debug/tracing/set_ftrace_filter는 ftrace를 사용하여 특정 이벤트를 추적하는 데 사용되는 반면, 
event enable은 perf를 사용하여 리눅스 커널 이벤트를 활성화하여 다양한 목적으로 사용됩니다.



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
#### 1. event trace.sh
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


## kprobe event 어떻게 발견하는가?
* /sys/kernel/debug/tracing/available_events
* perf 
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
### ftrace 가능한 커널 함수 목록: 
* /sys/kernel/debug/tracing/available_filter
```sh
root@Good:/sys/kernel/debug/tracing# grep  blk_account_io  /sys/kernel/debug/tracing/available_filter
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
