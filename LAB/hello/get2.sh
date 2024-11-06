#!/bin/bash
echo 0 > /sys/kernel/debug/tracing/tracing_on
echo 0 > /sys/kernel/debug/tracing/options/stacktrace
echo 0 > /sys/kernel/debug/tracing/events/enable

echo   > /sys/kernel/debug/tracing/set_ftrace_filter

cp  /sys/kernel/debug/tracing/trace ftrace.log
