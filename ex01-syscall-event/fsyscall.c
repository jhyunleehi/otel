//go:build ignore
#include "vmlinux.h"
//#include <bpf/bpf.h>
#include "../headers/bpf_helpers.h"

char __license[] SEC("license") = "Dual MIT/GPL";


struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 1 << 24);
} events SEC(".maps");


struct event {
	u32 pid;
	u8 comm[80];
	unsigned long dfd;
	unsigned long filename; 
    unsigned long how;
    unsigned long usize;
};

//__attribute__((unused))는 C나 C++에서 사용되는 컴파일러 지시자입니다. 
//이 지시자는 컴파일러에게 해당 변수나 함수가 사용되지 않을 것임을 알려줍니다. 
//주로 개발자가 의도적으로 변수나 함수를 선언하지만 사용하지 않을 때 사용됩니다. 
//이렇게 하면 컴파일러는 해당 변수나 함수를 사용하지 않은 것으로 간주하고, 
//이에 따른 경고 메시지를 표시하지 않게 됩니다.
const struct event *unused __attribute__((unused));

struct alloc_info {	
	unsigned long dfd;
	unsigned long filename; 
    unsigned long how;
    unsigned long usize;
};

// This tracepoint is defined in mm/page_alloc.c:__alloc_pages_nodemask()
// Userspace pathname: /sys/kernel/tracing/events/kmem/mm_page_alloc
SEC("tracepoint/syscalls/sys_enter_openat")
int syscalls_sys_enter_openat(struct alloc_info *args) {
	u64 id   = bpf_get_current_pid_tgid();
	u32 tgid = id >> 32;
	struct event *task_info;

	task_info = bpf_ringbuf_reserve(&events, sizeof(struct event), 0);
	if (!task_info) {
		return 0;
	}

	task_info->pid = tgid;
	bpf_get_current_comm(&task_info->comm, 80);
	bpf_ringbuf_submit(task_info, 0);
	return 0;
}
