//go:build ignore
#include "vmlinux.h"
#include "bpf_helpers.h"

#include "bpf_endian.h"
#include "bpf_tracing.h"

#include <bpf/bpf_core_read.h>

#define FMODE_CREATED	0x100000

#define AF_INET 2

const volatile pid_t targ_tgid = 0;

char __license[] SEC("license") = "Dual MIT/GPL";


struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 8192);
	__type(key, struct dentry *);
	__type(value, u64);
} start SEC(".maps");


struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 1 << 24);
} events SEC(".maps");

/**
 * The sample submitted to userspace over a ring buffer.
 * Emit struct event's type info into the ELF's BTF so bpf2go
 * can generate a Go type from it.
 */
struct event {
	u16 sport;
	u16 dport;
	u32 saddr;
	u32 daddr;
	u32 srtt;
};
struct event *unused_event __attribute__((unused));


SEC("fentry/tcp_close")
int BPF_PROG(tcp_close, struct sock *sk) {
	if (sk->__sk_common.skc_family != AF_INET) {
		return 0;
	}

	// The input struct sock is actually a tcp_sock, so we can type-cast
	struct tcp_sock *ts = bpf_skc_to_tcp_sock(sk);
	if (!ts) {
		return 0;
	}

	struct event *tcp_info;
	tcp_info = bpf_ringbuf_reserve(&events, sizeof(struct event), 0);
	if (!tcp_info) {
		return 0;
	}

	tcp_info->saddr = sk->__sk_common.skc_rcv_saddr;
	tcp_info->daddr = sk->__sk_common.skc_daddr;
	tcp_info->dport = bpf_ntohs(sk->__sk_common.skc_dport);
	tcp_info->sport = sk->__sk_common.skc_num;

	tcp_info->srtt = ts->srtt_us >> 3;
	tcp_info->srtt /= 1000;

	bpf_ringbuf_submit(tcp_info, 0);

	return 0;
}

static __always_inline int
probe_create(struct dentry *dentry)
{
	u64 id = bpf_get_current_pid_tgid();
	u32 tgid = id >> 32;
	u64 ts;

	if (targ_tgid && targ_tgid != tgid)
		return 0;

	ts = bpf_ktime_get_ns();
	bpf_map_update_elem(&start, &dentry, &ts, 0);
	return 0;
}



SEC("kprobe/vfs_open")
int BPF_KPROBE(vfs_open, struct path *path, struct file *file)
{
	struct dentry *dentry = BPF_CORE_READ(path, dentry);
	int fmode = BPF_CORE_READ(file, f_mode);

	if (!(fmode & FMODE_CREATED))
		return 0;

	return probe_create(dentry);
}