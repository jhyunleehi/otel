//go:build ignore

/* SPDX-License-Identifier: (LGPL-2.1 OR BSD-2-Clause) */
/* Copyright (c) 2021 Hengqi Chen */
#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_tracing.h>
#include "stat.h"

#define MAX_ENTRIES	10240

#define PATH_MAX	4096
#define TASK_COMM_LEN	16

enum op {
	READ,
	WRITE,
};

struct file_id {
	__u64 inode;
	__u32 dev;
	__u32 rdev;
	__u32 pid;
	__u32 tid;
};

struct file_stat {
	__u64 reads;
	__u64 read_bytes;
	__u64 writes;
	__u64 write_bytes;
	__u32 pid;
	__u32 tid;
	char filename[PATH_MAX];
	char path[PATH_MAX];
	char comm[TASK_COMM_LEN];
	char type;
};


const volatile pid_t target_pid = 0;
const volatile bool regular_file_only = true;
static struct file_stat zero_value = {};

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, MAX_ENTRIES);
	__type(key, struct file_id);
	__type(value, struct file_stat);
} entries SEC(".maps");

static void get_file_name(struct file *file, char *buf, size_t size)
{
	struct qstr dname;

	dname = BPF_CORE_READ(file, f_path.dentry, d_name);
	bpf_probe_read_kernel(buf, size, dname.name);
}

static int get_file_path(struct file *file, char *buf, size_t size)
{	
	bpf_printk("DEBUG: ----1----> get_file_path [%s]", buf);
	unsigned char * filename;
	filename = BPF_CORE_READ(file, f_path.dentry,d_iname);
	int ret = bpf_probe_read_kernel(buf, sizeof(buf), filename);
    if (ret < 0) {
		bpf_printk("DEBUG: ----2----> get_file_path [%s]", buf);
		return 0;	
	}  	
	
	bpf_printk("DEBUG: =>>3>>>>get_file_path [%s]", buf);
	return 0;
}

static int probe_entry(struct pt_regs *ctx, struct file *file, size_t count, enum op op)
{
	__u64 pid_tgid = bpf_get_current_pid_tgid();
	__u32 pid = pid_tgid >> 32;
	__u32 tid = (__u32)pid_tgid;
	int mode;
	struct file_id key = {};
	struct file_stat *valuep;

	if (target_pid && target_pid != pid)
		return 0;

	mode = BPF_CORE_READ(file, f_inode, i_mode);
	if (regular_file_only && !S_ISREG(mode))
		return 0;

	key.dev = BPF_CORE_READ(file, f_inode, i_sb, s_dev);
	key.rdev = BPF_CORE_READ(file, f_inode, i_rdev);
	key.inode = BPF_CORE_READ(file, f_inode, i_ino);
	key.pid = pid;
	key.tid = tid;
	valuep = bpf_map_lookup_elem(&entries, &key);
		
	if (!valuep) {
		bpf_map_update_elem(&entries, &key, &zero_value, BPF_ANY);
		valuep = bpf_map_lookup_elem(&entries, &key);
		if (!valuep)
			return 0;
		valuep->pid = pid;
		valuep->tid = tid;
		bpf_get_current_comm(&valuep->comm, sizeof(valuep->comm));
		get_file_name(file, valuep->filename, sizeof(valuep->filename));		
		
		if (S_ISREG(mode)) {
			valuep->type = 'R';
		} else if (S_ISSOCK(mode)) {
			valuep->type = 'S';
		} else {
			valuep->type = 'O';
		}
	}
	if (op == READ) {
		valuep->reads++;
		valuep->read_bytes += count;
	} else {	/* op == WRITE */
		valuep->writes++;
		valuep->write_bytes += count;
	}

	get_file_path(file, valuep->path, sizeof(valuep->path));
	
	return 0;
};

//static char gpath[PATH_MAX];

SEC("kprobe/vfs_read")
int BPF_KPROBE(vfs_read_entry, struct file *file, char *buf, size_t count, loff_t *pos)
{
	//bpf_printk("DEBUG: kprobe/vfs_read");

	int pid = bpf_get_current_pid_tgid() >> 32;
  	const char fmt_str[] = "DEBUG: kprobe/vfs_read [%d]";
  	bpf_trace_printk(fmt_str, sizeof(fmt_str), pid);

	
	// get_file_path(file, gpath, sizeof(gpath));
	// bpf_printk("DEBUG: ----1----> path [%s]", gpath);
	
	return probe_entry(ctx, file, count, READ);
}

SEC("kprobe/vfs_write")
int BPF_KPROBE(vfs_write_entry, struct file *file, const char *buf, size_t count, loff_t *pos)
{
	//bpf_printk("DEBUG: kprobe/vfs_write");

	int pid = bpf_get_current_pid_tgid() >> 32;
  	const char fmt_str[] = "DEBUG: kprobe/vfs_write [%d]";
  	bpf_trace_printk(fmt_str, sizeof(fmt_str), pid);

	return probe_entry(ctx, file, count, WRITE);
}

char LICENSE[] SEC("license") = "Dual BSD/GPL";
