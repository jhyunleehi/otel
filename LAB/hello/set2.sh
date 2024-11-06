#!/bin/bash
echo 0 > /sys/kernel/debug/tracing/tracing_on
echo 0 > /sys/kernel/debug/tracing/events/enable
echo   > /sys/kernel/debug/tracing/trace 

echo function > /sys/kernel/debug/tracing/current_tracer

echo 1 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_openat2/enable

echo do_sys_openat2  > /sys/kernel/debug/tracing/set_ftrace_filter
#echo ksys_read   >> /sys/kernel/debug/tracing/set_ftrace_filter
#echo ksys_write  >> /sys/kernel/debug/tracing/set_ftrace_filter
#echo close_fd  >> /sys/kernel/debug/tracing/set_ftrace_filter

echo 1 > /sys/kernel/debug/tracing/options/func_stack_trace
echo 1 > /sys/kernel/debug/tracing/options/sym-offset
echo 1 > /sys/kernel/debug/tracing/tracing_on
