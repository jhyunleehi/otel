# Tracing and Visualizing File System Internals with eBPF
## hello world
```c
#include <stdio.h>
char filename[] = "output.txt";
char message[] = "hello, world!\n";
FILE * file;
void rcall(int n){
    if (n<0) return;
    fputs( message, file);    
    fprintf(file,"[%d]:%s", n, message);
    rcall(n-1);
}
int main()
{
    file = fopen(filename, "w");
    if (file == NULL) return -1;    
    rcall(10);
    fclose(file);
    return 0;
}
```
```sh
$ gcc -g -pg -o hello hello.c
```
## Filesystem Internal
* sys_open
* vfs_open
* ext_file_open
* block devices
* HDD driver
* disk
[linux Kernel Map](https://makelinux.github.io/kernel/map)
[Storage function](https://en.wikibooks.org/wiki/The_Linux_Kernel/Storage)

## tracing with ftrace
ftrace 란
 
ftrace 리눅스 커널에서 제공하는 가장 강력한 트레이서입니다.
1. 인터럽트, 스케줄링, 커널 타이머 커널 동작을 상세히 추적해줍니다.
2. 함수 필터를 지정하면 자신을 호출한 함수와 전체 콜스택까지 출력합니다. 물론 코드를 수정할 필요가 습니다.
3. 함수를 어느 프로세스가 실행하는지 알 수 있습니다.
4. 함수 실행 시각을 알 수 있습니다.
5. ftrace 로그를 키면 시스템 동작에 부하를 주지 않습니다.
### 이벤트 (event)
* available_event
* `ls /sys/kernel/debug/tracing/available_events`
```
pi@raspberrypi:~$ sudo su -
root@raspberrypi:~# cd /sys/kernel/debug/tracing
root@raspberrypi:/sys/kernel/debug/tracing# ls -l available_*
-r--r--r-- 1 root root 0 Jan  1  1970 available_events
-r--r--r-- 1 root root 0 Jan  1  1970 available_filter_functions
-r--r--r-- 1 root root 0 Jan  1  1970 available_tracer
```
#### setf.sh
```
root@gpu-1:/sys/kernel/debug/tracing
# chmod 755 ~/setf.sh
root@gpu-1:/sys/kernel/debug/tracing# vi ~/setf.sh
```
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
### tracer 설정
ftrace는 nop, function, graph_function 트레이서를 제공합니다.
* nop: 기본 트레이서입니다. ftrace 이벤트만 출력합니다.**
* function: 함수 트레이서입니다. set_ftrace_filter로 지정한 함수를 누가 호출하는지 출력합니다.**
* graph_function: 함수 실행 시간과 세부 호출 정보를 그래프 포맷으로 출력합니다.**
```
root@raspberrypi:/sys/kernel/debug/tracing# cat current_tracer
nop
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
##  uftrace
```log
root@gpu-1:~# uftrace repla
uftrace: ./cmds/record.c:1542:find_in_path
 ERROR: Cannot trace 'repla': No such executable file.
root@gpu-1:~# uftrace record  -a  hello
root@gpu-1:~# uftrace replay
# DURATION     TID     FUNCTION
   0.552 us [171032] | __monstartup();
   0.121 us [171032] | __cxa_atexit();
            [171032] | main() {
 248.661 us [171032] |   fopen("output.txt", "w") = 0x5624e930c960;
            [171032] |   rcall(10) {
   3.810 us [171032] |     fputs("hello, world!\n", 0x5624e930c960) = 1;
   0.724 us [171032] |     fprintf(0x5624e930c960, "[%d]:%s") = 19;
            [171032] |     rcall(9) {
   0.164 us [171032] |       fputs("hello, world!\n", 0x5624e930c960) = 1;
   0.203 us [171032] |       fprintf(0x5624e930c960, "[%d]:%s") = 18;
            [171032] |       rcall(8) {
   1.403 us [171032] |         fputs("hello, world!\n", 0x5624e930c960) = 1;
   0.152 us [171032] |         fprintf(0x5624e930c960, "[%d]:%s") = 18;
            [171032] |         rcall(7) {
   0.095 us [171032] |           fputs("hello, world!\n", 0x5624e930c960) = 1;
   0.148 us [171032] |           fprintf(0x5624e930c960, "[%d]:%s") = 18;
            [171032] |           rcall(6) {
   0.094 us [171032] |             fputs("hello, world!\n", 0x5624e930c960) = 1;
   0.142 us [171032] |             fprintf(0x5624e930c960, "[%d]:%s") = 18;
            [171032] |             rcall(5) {
   0.092 us [171032] |               fputs("hello, world!\n", 0x5624e930c960) = 1;
   0.149 us [171032] |               fprintf(0x5624e930c960, "[%d]:%s") = 18;
            [171032] |               rcall(4) {
   1.297 us [171032] |                 fputs("hello, world!\n", 0x5624e930c960) = 1;
   0.151 us [171032] |                 fprintf(0x5624e930c960, "[%d]:%s") = 18;
            [171032] |                 rcall(3) {
   0.092 us [171032] |                   fputs("hello, world!\n", 0x5624e930c960) = 1;
   0.141 us [171032] |                   fprintf(0x5624e930c960, "[%d]:%s") = 18;
            [171032] |                   rcall(2) {
   0.089 us [171032] |                     fputs("hello, world!\n", 0x5624e930c960) = 1;
   0.146 us [171032] |                     fprintf(0x5624e930c960, "[%d]:%s") = 18;
            [171032] |                     rcall(1) {
   0.089 us [171032] |                       fputs("hello, world!\n", 0x5624e930c960) = 1;
   0.146 us [171032] |                       fprintf(0x5624e930c960, "[%d]:%s") = 18;
            [171032] |                       rcall(0) {
   1.294 us [171032] |                         fputs("hello, world!\n", 0x5624e930c960) = 1;
   0.206 us [171032] |                         fprintf(0x5624e930c960, "[%d]:%s") = 18;
   0.073 us [171032] |                         rcall(-1);
   2.031 us [171032] |                       } /* rcall */
   2.643 us [171032] |                     } /* rcall */
   3.217 us [171032] |                   } /* rcall */
   3.791 us [171032] |                 } /* rcall */
   5.591 us [171032] |               } /* rcall */
   6.173 us [171032] |             } /* rcall */
   6.735 us [171032] |           } /* rcall */
   7.308 us [171032] |         } /* rcall */
   9.224 us [171032] |       } /* rcall */
   9.979 us [171032] |     } /* rcall */
  15.392 us [171032] |   } /* rcall */
  37.456 us [171032] |   fclose(0x5624e930c960) = 0;
 302.814 us [171032] | } = 0; /* main */
```
### uftrace -K 5  -a  ./hello
#### file open, read, write, close
```log
# DURATION     TID     FUNCTION
   0.635 us [171545] | __monstartup();
   0.114 us [171545] | __cxa_atexit();
            [171545] | main() {
            [171545] |   fopen("output.txt", "w") {
# sys_open                
            [171545] |     __x64_sys_openat() {
            [171545] |       do_sys_openat2() {
            [171545] |         getname() {
            [171545] |           getname_flags.part.0() {
   0.502 us [171545] |             kmem_cache_alloc();
   0.691 us [171545] |             __check_object_size();
   1.755 us [171545] |           } /* getname_flags.part.0 */
   2.094 us [171545] |         } /* getname */
            [171545] |         get_unused_fd_flags() {
            [171545] |           alloc_fd() {
   0.149 us [171545] |             _raw_spin_lock();
   0.172 us [171545] |             expand_files();
   0.147 us [171545] |             _raw_spin_unlock();
   1.202 us [171545] |           } /* alloc_fd */
   1.533 us [171545] |         } /* get_unused_fd_flags */
            [171545] |         do_filp_open() {
            [171545] |           path_openat() {
   2.522 us [171545] |             alloc_empty_file();
   0.454 us [171545] |             path_init();
   6.524 us [171545] |             link_path_walk.part.0.constprop.0();
   1.415 us [171545] |             open_last_lookups();
   9.924 us [171545] |             do_open();
   0.910 us [171545] |             terminate_walk();
  23.088 us [171545] |           } /* path_openat */
  23.451 us [171545] |         } /* do_filp_open */
   0.154 us [171545] |         fd_install();
            [171545] |         putname() {
   0.186 us [171545] |           kmem_cache_free();
   0.516 us [171545] |         } /* putname */
  28.879 us [171545] |       } /* do_sys_openat2 */
  29.232 us [171545] |     } /* __x64_sys_openat */
...
# sys_read
            [171545] |     __x64_sys_read() {
            [171545] |       ksys_read() {
            [171545] |         __fdget_pos() {
   0.155 us [171545] |           __fget_light();
   0.503 us [171545] |         } /* __fdget_pos */
            [171545] |         vfs_read() {
            [171545] |           rw_verify_area() {
   0.565 us [171545] |             security_file_permission();
   0.900 us [171545] |           } /* rw_verify_area */
            [171545] |           seq_read() {
   0.158 us [171545] |             __get_task_ioprio();
  51.269 us [171545] |             seq_read_iter();
  51.974 us [171545] |           } /* seq_read */
  53.483 us [171545] |         } /* vfs_read */
  54.521 us [171545] |       } /* ksys_read */
  54.935 us [171545] |     } /* __x64_sys_read */
...
# sys_write
            [171545] |     __x64_sys_write() {
            [171545] |       ksys_write() {
            [171545] |         __fdget_pos() {
   0.164 us [171545] |           __fget_light();
   0.506 us [171545] |         } /* __fdget_pos */
            [171545] |         vfs_write() {
            [171545] |           rw_verify_area() {
   0.564 us [171545] |             security_file_permission();
   0.904 us [171545] |           } /* rw_verify_area */
   0.147 us [171545] |           __cond_resched();
   0.151 us [171545] |           __get_task_ioprio();
            [171545] |           ext4_file_write_iter() {
  25.975 us [171545] |             ext4_buffered_write_iter();
  26.336 us [171545] |           } /* ext4_file_write_iter */
   0.254 us [171545] |           __fsnotify_parent();
   0.153 us [171545] |           irq_enter_rcu();
            [171545] |           __sysvec_irq_work() {
   0.506 us [171545] |             __wake_up();
   0.500 us [171545] |             __wake_up();
   1.596 us [171545] |           } /* __sysvec_irq_work */
            [171545] |           irq_exit_rcu() {
   0.192 us [171545] |             idle_cpu();
   0.540 us [171545] |           } /* irq_exit_rcu */
  32.355 us [171545] |         } /* vfs_write */
  33.392 us [171545] |       } /* ksys_write */
  33.743 us [171545] |     } /* __x64_sys_write */
...
# sys_close
            [171545] |     __x64_sys_close() {
            [171545] |       close_fd() {
   0.154 us [171545] |         _raw_spin_lock();
   0.155 us [171545] |         pick_file();
   0.151 us [171545] |         _raw_spin_unlock();
            [171545] |         filp_close() {
   0.161 us [171545] |           dnotify_flush();
   0.162 us [171545] |           locks_remove_posix();
            [171545] |           fput() {
   0.260 us [171545] |             task_work_add();
   0.617 us [171545] |           } /* fput */
   1.703 us [171545] |         } /* filp_close */
   3.108 us [171545] |       } /* close_fd */
   3.458 us [171545] |     } /* __x64_sys_close */
```



### availabel filter_function
 * find supported kernel funtion  
```
root@gpu-2:/sys/kernel/debug/tracing# egrep  "ksys_read|ksys_write|close_fd|do_sys_openat2" avail*
available_filter_functions:do_sys_openat2
available_filter_functions:ksys_read
available_filter_functions:ksys_write
available_filter_functions:close_fd
available_filter_functions:ksys_readahead
```


### event trace 
```sh
root@Good:/sys/kernel/debug/tracing# grep sys_enter_open available_events 
syscalls:sys_enter_openat2
syscalls:sys_enter_openat
syscalls:sys_enter_open


root@Good:/sys/kernel/tracing/events/syscalls/sys_enter_openat2# cat format 
name: sys_enter_openat2
ID: 688
format:
	field:unsigned short common_type;	offset:0;	size:2;	signed:0;
	field:unsigned char common_flags;	offset:2;	size:1;	signed:0;
	field:unsigned char common_preempt_count;	offset:3;	size:1;	signed:0;
	field:int common_pid;	offset:4;	size:4;	signed:1;

	field:int __syscall_nr;	offset:8;	size:4;	signed:1;
	field:int dfd;	offset:16;	size:8;	signed:0;
	field:const char * filename;	offset:24;	size:8;	signed:0;
	field:struct open_how * how;	offset:32;	size:8;	signed:0;
	field:size_t usize;	offset:40;	size:8;	signed:0;

print fmt: "dfd: 0x%08lx, filename: 0x%08lx, how: 0x%08lx, usize: 0x%08lx", 
((unsigned long)(REC->dfd)), 
((unsigned long)(REC->filename)), 
((unsigned long)(REC->how)), 
((unsigned long)(REC->usize))
```

#### hello ftrace

```t
           hello-411245  [010] ..... 289429.761830: sys_openat(dfd: ffffff9c, filename: 7fb27f04c21b, flags: 80000, mode: 0)
           hello-411245  [010] ..... 289429.761857: sys_openat(dfd: ffffff9c, filename: 7fb27f01c140, flags: 80000, mode: 0)
           hello-411245  [010] ..... 289429.762192: sys_openat(dfd: ffffff9c, filename: 55ee1a8c7010, flags: 241, mode: 1b6)
           hello-411245  [010] ..... 289429.762354: sys_openat(dfd: ffffff9c, filename: 7fb27eddb628, flags: 20241, mode: 1b6)
```           