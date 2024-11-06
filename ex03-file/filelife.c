//go:build ignore

// SPDX-License-Identifier: GPL-2.0
// Copyright (c) 2020 Wenbo Zhang
#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_tracing.h>

/* linux: include/linux/fs.h */
#define FMODE_CREATED	0x100000

#define DNAME_INLINE_LEN	32
#define TASK_COMM_LEN		16

enum op {
  CREATE,
  OPEN,
  INODE,
  UNLINK,
};

struct fevent {  
  __u32 pid;
  __u32 tid;  
  char file[DNAME_INLINE_LEN];
  char task[TASK_COMM_LEN];  
  enum op op;
};


struct event {
	char file[DNAME_INLINE_LEN];
	char task[TASK_COMM_LEN];
	__u64 delta_ns;
	pid_t tgid;
	/* private */
	//void *dentry;
	__u64 dentry;
};

//__attribute__((unused))는 C나 C++에서 사용되는 컴파일러 지시자입니다. 
//이 지시자는 컴파일러에게 해당 변수나 함수가 사용되지 않을 것임을 알려줍니다. 
//주로 개발자가 의도적으로 변수나 함수를 선언하지만 사용하지 않을 때 사용됩니다. 
//이렇게 하면 컴파일러는 해당 변수나 함수를 사용하지 않은 것으로 간주하고, 
//이에 따른 경고 메시지를 표시하지 않게 됩니다.
const struct event *unused __attribute__((unused));

struct {
  __uint(type, BPF_MAP_TYPE_RINGBUF);
  __uint(max_entries, 1 << 24);
} fevents SEC(".maps");

struct {
	__uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
	__uint(key_size, sizeof(u32));
	__uint(value_size, sizeof(u32));
} events SEC(".maps");



struct renamedata___x {
	struct user_namespace *old_mnt_userns;
	struct new_mnt_idmap *new_mnt_idmap;
} __attribute__((preserve_access_index));

static __always_inline bool renamedata_has_old_mnt_userns_field(void)
{
	if (bpf_core_field_exists(struct renamedata___x, old_mnt_userns))
		return true;
	return false;
}

static __always_inline bool renamedata_has_new_mnt_idmap_field(void)
{
	if (bpf_core_field_exists(struct renamedata___x, new_mnt_idmap))
		return true;
	return false;
}

const volatile pid_t targ_tgid = 0;

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 8192);
	__type(key, struct dentry *);
	__type(value, u64);
} start SEC(".maps");



struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 8192);
	__type(key, u32); /* tid */
	__type(value, struct event);
} currevent SEC(".maps");

static __always_inline int probe_create(struct dentry *dentry)
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

static int submit_fevent(enum op op, struct dentry *dentry ) 
{  
	
	__u64 pid_tgid = bpf_get_current_pid_tgid();
  	__u32 pid = pid_tgid >> 32;
  	__u32 tid = (__u32)pid_tgid; 

	bpf_printk("DEBUG: submit_fevent===>> [%d]", pid);

	struct fevent *feventp;
	feventp = bpf_ringbuf_reserve(&fevents, sizeof(struct fevent), 0);
  	if (!feventp) {
	    return 0;
  	}


  	feventp->pid = pid;
  	feventp->tid = tid;    
  	feventp->op = op;
  	bpf_get_current_comm(&feventp->task, sizeof(feventp->task));
	if (dentry) {
		const u8 *qs_name_ptr =	BPF_CORE_READ(dentry, d_name.name);
		bpf_probe_read_kernel_str(&feventp->file, sizeof(feventp->file), qs_name_ptr);
		//bpf_probe_read_kernel_str(&fevent.file, sizeof(fevent.file), dentry->d_name.name);
	}
	bpf_ringbuf_submit(feventp, 0);
  
  return 0;
}



/**
 * In different kernel versions, function vfs_create() has two declarations,
 * and their parameter lists are as follows:
 *
 * int vfs_create(struct inode *dir, struct dentry *dentry, umode_t mode,
 *            bool want_excl);
 * int vfs_create(struct user_namespace *mnt_userns, struct inode *dir,
 *            struct dentry *dentry, umode_t mode, bool want_excl);
 * int vfs_create(struct mnt_idmap *idmap, struct inode *dir,
 *            struct dentry *dentry, umode_t mode, bool want_excl);
 */
SEC("kprobe/vfs_create")
int BPF_KPROBE(vfs_create, void *arg0, void *arg1, void *arg2){
//int  kprobe_vfs_create(struct pt_regs *ctx) {
	if (renamedata_has_old_mnt_userns_field() || renamedata_has_new_mnt_idmap_field()){
		probe_create(arg2);
		submit_fevent(CREATE, arg2);
	}
	else{
		probe_create(arg1);
		submit_fevent(CREATE, arg1);
	}

	int pid = bpf_get_current_pid_tgid() >> 32;
  	const char fmt_str[] = "VFS create [%d]";
  	bpf_trace_printk(fmt_str, sizeof(fmt_str), pid);


	return 0;
}


SEC("kprobe/vfs_open")
int BPF_KPROBE(vfs_open, struct path *path, struct file *file)
{
	struct dentry *dentry = BPF_CORE_READ(path, dentry);
	int fmode = BPF_CORE_READ(file, f_mode);

	if (!(fmode & FMODE_CREATED))
		return 0;

	int pid = bpf_get_current_pid_tgid() >> 32;
  	const char fmt_str[] = "VFS open [%d]";
  	bpf_trace_printk(fmt_str, sizeof(fmt_str), pid);

	submit_fevent(OPEN, dentry);

	return probe_create(dentry);
}

SEC("kprobe/security_inode_create")
int BPF_KPROBE(security_inode_create, struct inode *dir,
	     struct dentry *dentry)
{
	submit_fevent(INODE, dentry);

	return probe_create(dentry);
}

/**
 * In different kernel versions, function vfs_unlink() has two declarations,
 * and their parameter lists are as follows:
 *
 * int vfs_unlink(struct inode *dir, struct dentry *dentry,
 *        struct inode **delegated_inode);
 * int vfs_unlink(struct user_namespace *mnt_userns, struct inode *dir,
 *        struct dentry *dentry, struct inode **delegated_inode);
 * int vfs_unlink(struct mnt_idmap *idmap, struct inode *dir,
 *        struct dentry *dentry, struct inode **delegated_inode);
 */
SEC("kprobe/vfs_unlink")
int BPF_KPROBE(vfs_unlink, void *arg0, void *arg1, void *arg2)
{
	u64 id = bpf_get_current_pid_tgid();
	struct event event = {};
	const u8 *qs_name_ptr;
	u32 tgid = id >> 32;
	u32 tid = (u32)id;
	u64 *tsp, delta_ns;
	bool has_arg = renamedata_has_old_mnt_userns_field()
				|| renamedata_has_new_mnt_idmap_field();

	tsp = has_arg
		? bpf_map_lookup_elem(&start, &arg2)
		: bpf_map_lookup_elem(&start, &arg1);
	if (!tsp)
		return 0;   // missed entry

	delta_ns = bpf_ktime_get_ns() - *tsp;

	qs_name_ptr = has_arg
		? BPF_CORE_READ((struct dentry *)arg2, d_name.name)
		: BPF_CORE_READ((struct dentry *)arg1, d_name.name);    		

	bpf_probe_read_kernel_str(&event.file, sizeof(event.file), qs_name_ptr);
	bpf_get_current_comm(&event.task, sizeof(event.task));
	event.delta_ns = delta_ns;
	event.tgid = tgid;
	event.dentry = (__u64)(has_arg ? arg2 : arg1);

	bpf_map_update_elem(&currevent, &tid, &event, BPF_ANY);

	int pid = bpf_get_current_pid_tgid() >> 32;
  	const char fmt_str[] = "VFS ulink [%d]";
  	bpf_trace_printk(fmt_str, sizeof(fmt_str), pid);


	if (has_arg ){		
		submit_fevent(UNLINK, arg2);
	}else{
		submit_fevent(UNLINK, arg1);
	}

	return 0;
}

SEC("kretprobe/vfs_unlink")
int BPF_KRETPROBE(vfs_unlink_ret)
{
	u64 id = bpf_get_current_pid_tgid();
	u32 tid = (u32)id;
	int ret = PT_REGS_RC(ctx);
	struct event *event;

	event = bpf_map_lookup_elem(&currevent, &tid);
	if (!event)
		return 0;
	bpf_map_delete_elem(&currevent, &tid);

	/* skip failed unlink */
	if (ret)
		return 0;

	bpf_map_delete_elem(&start, &event->dentry);

	/* output */
	bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU,
			      event, sizeof(*event));
	return 0;
}

char LICENSE[] SEC("license") = "GPL";
