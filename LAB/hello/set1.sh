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
