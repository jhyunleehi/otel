# Tracing and Visualizing
1. perf
2. uftrace
2. /proc 
2. chrom://tracing 
3. perfetto 
4. ftrace 
5. uftrace 

dstat 

## 1. perf : Performance analysis tools for Linux
perf는 리눅스 시스템에서의 성능 분석과 프로파일링을 위한 강력한 도구입니다. 이를 사용하면 CPU 사용량, 메모리 액세스, 함수 호출 및 여러 다른 이벤트를 추적하여 시스템의 동작을 분석할 수 있습니다.

1. perf 

```sh
$ sudo apt install linux-tools
$ sudo apt install linux-cloud-tools 
$ sudo perf list 
$ sudo perf list | grep  sys_enter_openat
  syscalls:sys_enter_openat                          [Tracepoint event]
  syscalls:sys_enter_openat2                         [Tracepoint event]
```

2. make hello trace
```sh
$ gcc -g -pg  -o hello hello.c
$ ldd hello
	linux-vdso.so.1 (0x00007ffc91b67000)
	libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x000072f894800000)
	/lib64/ld-linux-x86-64.so.2 (0x000072f894a5f000)

$ file  hello
hello: ELF 64-bit LSB pie executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, BuildID[sha1]=950f5919b87ec318f94ac423ef161d968cb7c6a9, for GNU/Linux 3.2.0, with debug_info, not stripped

$ readelf  -l hello

$ objdump -d hello

$ xxd ./hello | header 
```

* gdb를 통한 추적
```
$ gdb  ./hello
(gdb) b main
(gdb) info break 
(gdb) info file
(gdb) info proc
(gdb) info stack
(gdb) info frame
(gdb) info r

(gdb) p message 
 $4 = "hello, world!\n"

(gdb) x/20x  0x0000555555555120  <<-- text segment 
(gdb) disas main
(gdb) x/10x 0x00005555555552d4

```

* user process 추적
```sh
$ struss ./hello
```

이렇게 추적하는 것은 user space에서만 추적인 가능하다. 

### perf command 

#### 1. cpu event 분석
CPU의 instruction retired 이벤트를 추적하여 프로그램의 명령어 실행 수를 확인할 수 있습니다.

```sh
$ sudo perf stat -e instructions  ./hello

 Performance counter stats for './hello':

         1,438,114      instructions                                                          

       0.001684180 seconds time elapsed

       0.001680000 seconds user
       0.000000000 seconds sys

```


#### 2. Profiling 
프로그램 실행 시간을 추적하여 코드에서 시간이 가장 많이 소비되는 함수를 확인할 수 있습니다.
* -p : process 정보 지정
* -a : 모든 process 정보
* -g : stack 정보 저장 
* -e : 특정 event 저장 
* -F : 분석 frequency 지정 


* hello trace heat map 
```sh
$ sudo perf record  -g -- ./hello
$ perf report --stdio --sort comm,dso,symbol --tui
```

* event 지정 
```sh 
$ sudo perf record  -e cpu-clock -g -- ./hello
$ ll
-rw------- 1 jhyunlee jhyunlee 41138  4월 14 15:00 perf.data
$ sudo  perf report -f

e key
c key
```


#### 3. 히트맵 출력:
프로그램 실행 중에 perf를 사용하여 동적으로 히트맵을 출력합니다.

```sh
$ perf record -e cpu-clock  -ag -- sleep 10
$ perf report --stdio --sort comm,dso,symbol --tui
```
결론 ==> 이것은 분석하기 너무 힘들다.


#### 4. Flame Graph 
* 설치 방법 
```sh
$ git clone https://github.com/brendangregg/Flamegraph.git
$ sudo perf script | ../Flamegraph/stackcollapse-perf.pl | ../Flamegraph/flamegraph.pl > graph.svg
```

* 흥미로운 결과가 나온다. 
```sh
$ sudo perf record  -ag -- sleep 10
$ sudo perf script | ../Flamegraph/stackcollapse-perf.pl | ../Flamegraph/flamegraph.pl > graph.svg
```

* hello flame graph
```sh
$ sudo perf record  -g -- ./hello
$ sudo perf script | ../Flamegraph/stackcollapse-perf.pl | ../Flamegraph/flamegraph.pl > graph.svg
```


### Linux Perf profiler UI
[https://www.markhansen.co.nz/profiler-uis/](https://www.markhansen.co.nz/profiler-uis/)
#### 1. perf report 
```sh
$ perf report
$ perf report --stdio 
$ perf report --stdio --sort comm,dso,symbol --tui
```
#### 2. firefox profiler 
```sh
$ perf script -F +pid > perf_fox.perf
```

1. firefox open
2. open url https://profiler.firefox.com/
3. $ perf script -F +pid > /tmp/test.perf


#### 3. flame Graph 
* perf data가 있는 디렉토리에서 실행합니다. 
```sh
$ git clone https://github.com/brendangregg/Flamegraph.git
$ perf script | ./stackcollapse-perf.pl | ./flamegraph.pl > flame.html
```

#### 4. hotspot
https://github.com/KDAB/hotspot?tab=readme-ov-file#debian--ubuntu

```sh
$ sudo apt install hotspot
$ sudo hotspot
 ==> record 를 통해서 ./hello 프로그램을 trace 해본다.  
$ perf record -o /home/jhyunlee/go/src/ebpf-go/LAB/01.pef/perf.data --call-graph dwarf --aio --sample-cpu /home/jhyunlee/go/src/ebpf-go/LAB/01.pef/hell
```

#### 5. perfetto
* An amazing tracer
* I use it all the time for Android and Chrome tracing
* But doesn’t yet support reading perf.data files

### Quickstart: Record traces on Linux
https://perfetto.dev/docs/quickstart/linux-tracing?_gl=1*15ruz9r*_ga*MTQ3NjY1ODc1NS4xNzEzMDE5NzY0*_ga_BD89KT2P3C*MTcxMzA4OTIwMC4yLjEuMTcxMzA5MTE2NS4wLjAuMA..

Building from source

```sh
1. Check out the code:
$ git clone https://android.googlesource.com/platform/external/perfetto/ && cd perfetto

2. Download and extract build dependencies:
  If the script fails with SSL errors, try upgrading your openssl package.
$ tools/install-build-deps

3. Generate the build configuration
$ tools/gn gen --args='is_debug=false' out/linux
# Or use `tools/setup_all_configs.py` to generate more build configs.

4. Build the Linux tracing binaries (On Linux it uses a hermetic clang toolchain, downloaded as part of step 2):
$ tools/ninja -C out/linux tracebox traced traced_probes perfetto 
```


#### 7. pprof
* 이것은 좀더 TEST 해봐야
```sh
$ go install github.com/google/pprof@latest
```
* pprof can read perf.data files generated by the Linux perf tool by using the perf_to_profile program from the perf_data_converter package.

```sh
$ sudo perf record -p (pidof node_exporter) -g -F 4000
$ sudo chown mark perf.data
$ perf_to_profile -i perf.data -o node_exporter_pprof.profile
```



## 2. uftrace 

[https://github.com/namhyung/uftrace](https://github.com/namhyung/uftrace)

hese are the commands supported by uftrace:

* record : runs a program and saves the trace data
* replay : shows program execution in the trace data
* report : shows performance statistics in the trace data
* live : does record and replay in a row (default)
* info : shows system and program info in the trace data
* dump : shows low-level trace data
* recv : saves the trace data from network
* graph : shows function call graph in the trace data
* script : runs a script for recorded trace data
* tui : show text user interface for graph and report

### 1. make hello 

```sh 
$ gcc -g -pg  -o hello hello.c

$ ldd  ./hello
$ file ./hello
$ stat ./hello
$ readelf  -l hello
$ objdump -d hello
$ xxd ./hello | header 

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
```

### 2. trace kernel 
* kernel 함수 추적
```sh 
$ uftrace ./hello
$ sudo uftrace ./hello 

$ sudo uftrace -k 5 ./hello
```

* 기록
```sh
$ sudo uftrace record -K 5 ./hello
$ du -sk  uftrace.data/
18284	uftrace.data/

$ sudo uftrace report 
$ sudo uftrace replay
$ sudo uftrace tui
```

### 3. flame graph
```sh
$ uftrace dump --chrome > chrome.data
$ uftrace dump --flame-graph > flame.data
```
* ui.perfetto.dev 에서 chrome.data 읽기


* svg graph 생성 
```sh
 $ uftrace dump --flame-graph |../Flamegraph/flamegraph.pl > f.svg
 ```



## 3. ftrace 

###  기본 test step
```sh
1. # cat /sys/kernel/debug/tracing/trace_pipe
2. # echo 1 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_openat/enable
3. # echo 1 > /sys/kernel/debug/tracing/tracing_on
4. $ ./hello
5. # echo 0 > /sys/kernel/debug/tracing/tracing_on
6. # echo 0 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_openat/enable
```


#### ftrace 

ftrace는 리눅스 커널에서 발생하는 이벤트를 추적하고 분석하는 도구입니다. 
주로 커널 내부의 동작을 이해하거나 디버깅하기 위해 사용됩니다.

#### 특징

1. 이벤트 추적: ftrace를 사용하여 커널 내에서 발생하는 다양한 이벤트를 추적할 수 있습니다. 예를 들어, 함수 호출, 인터럽트, 스케줄링 이벤트 등을 추적할 수 있습니다.
2. 동적 추적: ftrace는 시스템을 실행 중에 동적으로 활성화하거나 비활성화할 수 있습니다. 이를 통해 필요한 이벤트만 추적하고 성능에 영향을 최소화할 수 있습니다.
3. 사용자 정의 훅: ftrace를 사용하여 사용자가 원하는 이벤트를 추적할 수 있는 사용자 정의 훅을 설정할 수 있습니다. 이를 통해 특정 조건에 따라 원하는 이벤트를 자동으로 캡처할 수 있습니다.
4. 많은 도구와 통합: ftrace는 여러 다른 도구와 통합되어 있어서, trace-cmd와 같은 도구를 통해 쉽게 사용할 수 있습니다.


#### tracer

ftrace는 nop, function, graph_function 트레이서를 제공합니다.
* nop: 기본 트레이서입니다. ftrace 이벤트만 출력합니다.**
* function: 함수 트레이서입니다. set_ftrace_filter로 지정한 함수를 누가 호출하는지 출력합니다.
* graph_function: 함수 실행 시간과 세부 호출 정보를 그래프 포맷으로 출력합니다.**

```sh
$ echo function_graph > /sys/kernel/debug/tracing/current_tracer
$ cat /sys/kernel/debug/tracing/trace
```

#####  available_tracers 
* 파일 available_tracers에는 커널에서 사용할 수 있는 추적 프로그램이 나열되어 있습니다. 
* 추적 프로그램의 예로는 함수 추적 프로그램, 함수 그래프 추적 프로그램, 이벤트 추적 프로그램 등이 있습니다.
* 추적 프로그램은 ftrace커널 내의 특정 이벤트 또는 함수 호출에 대한 추적 데이터를 캡처할 수 있는 메커니즘입니다. 추적 프로그램마다 기능과 오버헤드가 다릅니다. 

```sh
root@Good:/sys/kernel/debug/tracing# cat available_tracers 
no  <<---
function    <<----
function_graph  <<---
timerlat 
osnoise 
hwlat 
blk 
mmiotrace 
wakeup_dl 
wakeup_rt 
wakeup 
```


### 1. event trace 
#### available_events 
1. event enable은 perf 유틸리티를 사용하여 활성화되는 리눅스 커널 이벤트입니다.
2. perf를 사용하여 다양한 이벤트를 모니터링하고 분석할 수 있습니다.
3. 이벤트 활성화를 사용하여 프로파일링이나 성능 모니터링과 같은 작업을 수행할 수 있습니다.
4. 주로 성능 최적화, 시스템 모니터링 및 디버깅에 사용됩니다.

#### available_events 
* 커널 내에서 추적에 사용할 수 있는 이벤트를 나타냅니다. 
* 이러한 이벤트는 다양한 함수 호출, 스케줄러 이벤트, 인터럽트 또는 커널 내의 기타 추적 가능한 활동일 수 있습니다. 
* available_events이러한 이벤트는 일반적으로 추적 디렉터리( )의 파일 에 나열됩니다 /sys/kernel/debug/tracing/evnts 
* 런타임 중에 발생하는 이벤트에 대한 정보를 수집하기 위해 이러한 이벤트에 대한 추적을 활성화할 수 있습니다.
```
root@Good:/sys/kernel/debug/tracing# cat available_events  | grep openat
syscalls:sys_exit_openat2
syscalls:sys_enter_openat2
```

####  event trace.sh 예시
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


### 2. funtion trace 
#### set_ftrace_filter 설정
* set_ftrace_filter 파일에 트레이싱하고 싶은 함수를 지정하면 된다.
* 위의 tracer 설정의 function 혹은 function_graph으로 설정한 경우 작동하는 파일이다.
* 리눅스 커널에 존재하는 모든 함수를 필터로 지정할 수는 없다.
* /sys/kernel/debug/tracing/available_filter_functions 파일에 포함된 함수만 지정할 수 있다.
* 함수를 지정하지 않은 경우 모든 함수를 트레이싱하게 되어 락업이 상태에 빠지게 된다.
* available_filter_functions 파일에 없는 함수를 지정하려도 락업 상태가 될 수 있으니 주의하자.
* set_ftrace_filter에 아무것도 설정하지 않고 ftrace를 키면, ftrace는 모든 커널 함수에 대하여 트레이싱을 한다.
* 모든 커널 함수에 의해 트레이스가 발생되면, 그 오버헤드가 엄청나 시스템은 락업 상태에 빠진다.
* 그러므로 부팅 이후 절대 불리지 않을 함수secondary_start_kernel2를 트레이스 포인트로 찍어준다.

#### available_filter_functions
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

#### kernel함수 trace (file open,read,write,close) 예시
```sh
#!/bin/bash
echo 0 > /sys/kernel/debug/tracing/tracing_on
echo 0 > /sys/kernel/debug/tracing/events/enable
echo function > /sys/kernel/debug/tracing/current_tracer
echo do_sys_openat2  > /sys/kernel/debug/tracing/set_ftrace_filter
echo ksys_read   >> /sys/kernel/debug/tracing/set_ftrace_filter
echo ksys_write  >> /sys/kernel/debug/tracing/set_ftrace_filter
echo close_fd   >> /sys/kernel/debug/tracing/set_ftrace_filter
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


요약하면, 
1. set_ftrace_filter는 ftrace를 사용하여 특정 이벤트를 추적하는데 사용
2. event enable은 perf를 사용하여 리눅스 커널 이벤트를 활성화하여 다양한 목적으로 사용됩니다.



### trace_eventget.sh

* set.sh
```sh
#!/bin/bash
echo 0 > /sys/kernel/debug/tracing/tracing_on
echo 0 > /sys/kernel/debug/tracing/events/enable
echo   > /sys/kernel/debug/tracing/trace 

echo nop > /sys/kernel/debug/tracing/current_tracer

#echo 1 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_open/enable
echo 1 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_openat/enable
#echo 1 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_openat2/enable
#echo 1 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_open_tree/enable
#echo 1 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_open_by_handle_at/enable

# echo 1 > /sys/kernel/debug/tracing/events/sched/sched_switch/enable
# echo 1 > /sys/kernel/debug/tracing/events/irq/irq_handler_entry/enable
# echo 1 > /sys/kernel/debug/tracing/events/irq/irq_handler_exit/enable
# echo 1 > /sys/kernel/debug/tracing/events/raw_syscalls/enable

echo 1 > /sys/kernel/debug/tracing/options/func_stack_trace
echo 1 > /sys/kernel/debug/tracing/options/sym-offset
echo 1 > /sys/kernel/debug/tracing/tracing_on
```

* get.sh
```sh
#!/bin/bash
echo 0 > /sys/kernel/debug/tracing/tracing_on
echo 0 > /sys/kernel/debug/tracing/options/stacktrace
echo 0 > /sys/kernel/debug/tracing/events/enable

#echo 0 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_open/enable
echo 0 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_openat/enable
#echo 0 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_openat2/enable
#echo 0 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_open_tree/enable
#echo 0 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_open_by_handle_at/enable

#echo 0 > /sys/kernel/debug/tracing/events/sched/sched_switch/enable
#echo 0 > /sys/kernel/debug/tracing/events/irq/irq_handler_entry/enable
#echo 0 > /sys/kernel/debug/tracing/events/irq/irq_handler_exit/enable
#echo 0 > /sys/kernel/debug/tracing/events/raw_syscalls/enable

cp  /sys/kernel/debug/tracing/trace ftrace.log
```


### trace_funtion.sh
* set 

```sh
#!/bin/bash
echo 0 > /sys/kernel/debug/tracing/tracing_on
echo 0 > /sys/kernel/debug/tracing/events/enable
echo   > /sys/kernel/debug/tracing/trace 

echo function > /sys/kernel/debug/tracing/current_tracer

echo do_sys_openat2  > /sys/kernel/debug/tracing/set_ftrace_filter
#echo ksys_read   >> /sys/kernel/debug/tracing/set_ftrace_filter
#echo ksys_write  >> /sys/kernel/debug/tracing/set_ftrace_filter
#echo close_fd  >> /sys/kernel/debug/tracing/set_ftrace_filter

echo 1 > /sys/kernel/debug/tracing/options/func_stack_trace
echo 1 > /sys/kernel/debug/tracing/options/sym-offset
echo 1 > /sys/kernel/debug/tracing/tracing_on
```

* get.sh
```sh
#!/bin/bash
echo 0 > /sys/kernel/debug/tracing/tracing_on
echo 0 > /sys/kernel/debug/tracing/options/stacktrace
echo 0 > /sys/kernel/debug/tracing/events/enable
echo   > /sys/kernel/debug/tracing/set_ftrace_filter
cp  /sys/kernel/debug/tracing/trace ftrace.log
```




## 4. trace-cmd
* 맛보기
```sh
$ sudo trace-cmd record -p function ./hello
$ sudo trace-cmd record -p funtion  -P 1234
$ sudo trace-cmd report
```
==> 이것의 출력은 perf 출력이 아니다





### perf와 trace-cmd의 차이점 
### perf:
1. perf는 linux kernel의 일부분으로 성능분석 도구이다.   
2. cpu, memoery, io 등 전반적인 성능 분석을 위한 기능들을 제공한다.  
3. perf can be used for profiling CPU usage, monitoring hardware performance counters, tracing function calls, and more.
4. profiling 방식은 sampling-based profiling, tracing-based profiling 2가지를 지원한다.  
5. perf operates at the system level and can provide insights into both kernel-space and user-space activities.

### trace-cmd:
1. trace-cmd는 Ftrace 시스템과 연동하여 기능을 제공하는 coomand-line tool 이다.  
2. Ftrace는 a built-in Linux kernel tracing framework.
3. It provides a convenient way to control and manage kernel tracing sessions using Ftrace.
4. you can start and stop tracing sessions, configure trace buffers, view trace data, and more.
5. trace-cmd focuses specifically on kernel tracing and provides a higher-level interface compared to using Ftrace directly.



## 5. kprobe 

kprobe는 리눅스 커널 내에서 특정 함수나 코드 영역에 프로브(Probe)를 삽입하여 해당 위치에서 발생하는 이벤트를 추적하는 기술입니다. 
'kprobe'는 'kernel probe'의 줄임말이며, 커널 내부에서 동작하는 기능을 추적하고 분석하는 데 사용됩니다.

### 특징

1. 커널 함수 호출 추적: kprobe를 사용하여 커널 내의 특정 함수가 호출될 때 이벤트를 캡처할 수 있습니다. 이를 통해 커널 함수의 호출 빈도나 매개변수 값을 추적할 수 있습니다.
2. 디버깅 및 분석: kprobe를 사용하여 커널 내부의 동작을 디버깅하고 분석할 수 있습니다. 예를 들어, 커널 패닉이 발생할 때 특정 함수가 호출되었는지 추적하여 원인을 찾을 수 있습니다.
3. 성능 프로파일링: kprobe를 사용하여 커널 내에서 성능에 영향을 미치는 부분을 식별하고 프로파일링할 수 있습니다. 이를 통해 시스템의 병목 현상을 파악하고 최적화할 수 있습니다.
4. 사용자 정의 이벤트 추적: 사용자가 원하는 위치에 kprobe를 삽입하여 사용자 정의 이벤트를 추적할 수 있습니다. 이를 통해 특정 조건이나 상황에서 발생하는 이벤트를 캡처하여 분석할 수 있습니다.

kprobe는 perf나 ftrace와 함께 사용되어 시스템의 동작을 분석하고 디버깅하는 데 활용됩니다. 사용자가 원하는 위치에 프로브를 삽입하여 필요한 이벤트를 추적할 수 있어, 다양한 시나리오에 유용하게 사용됩니다.

###  kprobe  and  ftrace (/sys/kernel/debug/tracing) 차이점 

When learning about the Linux kernel, understanding the difference between Kprobes and Ftrace can be crucial, as they are both tools used for kernel debugging and tracing, but they serve different purposes and operate at different levels of the kernel.

1. Kprobes: Kprobes is a dynamic kernel debugging mechanism that allows developers to insert breakpoints (probes) into running kernel code. These probes can be used to monitor the execution flow of the kernel, gather information about specific events, or debug kernel code without requiring recompilation or rebooting the system. Kprobes allows developers to attach "probe handlers" to specific locations in the kernel code, which are executed when the probe is hit. This mechanism is particularly useful for debugging complex kernel issues or analyzing kernel behavior in real-time.

2. Ftrace: Ftrace, on the other hand, is a kernel tracing framework that provides a set of tools for tracing various kernel events and functions. It allows developers to dynamically instrument the kernel to collect detailed information about its behavior, such as function call traces, context switches, interrupt activity, and more. Ftrace provides a powerful interface for analyzing kernel performance, identifying bottlenecks, and diagnosing issues. It consists of several components, including function tracer, function graph tracer, event tracer, and tracepoints. Ftrace is typically used for performance analysis, optimization, and understanding kernel internals.

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


### 결론 
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


### kprobe 커널 함수 호출 추적:

#### 1. do_sys_open() 함수를 추적

```
$ grep do_sys_open /proc/kallsyms
```
### 2. kprobe를 설정하여 해당 함수 호출을 추적합니다.

```
$ echo 'p:function do_sys_open' > /sys/kernel/debug/tracing/kprobe_events
```

#### 3.  추적을 시작합니다.

```
$ echo 1 > /sys/kernel/debug/tracing/events/kprobes/enable
```
#### 4. 추적 결과
이제 특정 동작을 수행하면서 do_sys_open() 함수가 호출될 때마다 이를 추적합니다. 추적 결과는 다음과 같이 확인할 수 있습니다.

```
$ cat /sys/kernel/debug/tracing/trace
```

필요에 따라 trace-cmd 또는 perf와 같은 도구를 사용하여 추적 결과를 분석할 수 있습니다.





##  install bpftool 
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