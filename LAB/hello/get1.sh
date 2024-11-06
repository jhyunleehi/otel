#!/bin/bash
echo 0 > /sys/kernel/debug/tracing/tracing_on
echo 0 > /sys/kernel/debug/tracing/options/stacktrace
echo 0 > /sys/kernel/debug/tracing/events/enable

echo 0 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_open/enable
echo 0 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_openat/enable
echo 0 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_openat2/enable
echo 0 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_open_tree/enable
echo 0 > /sys/kernel/debug/tracing/events/syscalls/sys_enter_open_by_handle_at/enable
#echo 0 > /sys/kernel/debug/tracing/events/sched/sched_switch/enable
#echo 0 > /sys/kernel/debug/tracing/events/irq/irq_handler_entry/enable
#echo 0 > /sys/kernel/debug/tracing/events/irq/irq_handler_exit/enable
#echo 0 > /sys/kernel/debug/tracing/events/raw_syscalls/enable

cp  /sys/kernel/debug/tracing/trace ftrace.log
