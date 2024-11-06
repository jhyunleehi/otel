## test step

1. cat /sys/kernel/tracing/trace-pipe
2. echo 1 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_openat/enable
3. echo 1 > /sys/kernel/debug/tracing/tracing_on
4. ./hello
5. echo 0 > /sys/kernel/debug/tracing/tracing_on
6. echo 0 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_openat/enable