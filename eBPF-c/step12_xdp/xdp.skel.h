/* SPDX-License-Identifier: (LGPL-2.1 OR BSD-2-Clause) */

/* THIS FILE IS AUTOGENERATED BY BPFTOOL! */
#ifndef __XDP_SKEL_H__
#define __XDP_SKEL_H__

#include <errno.h>
#include <stdlib.h>
#include <bpf/libbpf.h>

struct xdp {
	struct bpf_object_skeleton *skeleton;
	struct bpf_object *obj;
	struct {
		struct bpf_map *xdp_stats_map;
	} maps;
	struct {
		struct bpf_program *xdp_prog_func;
	} progs;
	struct {
		struct bpf_link *xdp_prog_func;
	} links;

#ifdef __cplusplus
	static inline struct xdp *open(const struct bpf_object_open_opts *opts = nullptr);
	static inline struct xdp *open_and_load();
	static inline int load(struct xdp *skel);
	static inline int attach(struct xdp *skel);
	static inline void detach(struct xdp *skel);
	static inline void destroy(struct xdp *skel);
	static inline const void *elf_bytes(size_t *sz);
#endif /* __cplusplus */
};

static void
xdp__destroy(struct xdp *obj)
{
	if (!obj)
		return;
	if (obj->skeleton)
		bpf_object__destroy_skeleton(obj->skeleton);
	free(obj);
}

static inline int
xdp__create_skeleton(struct xdp *obj);

static inline struct xdp *
xdp__open_opts(const struct bpf_object_open_opts *opts)
{
	struct xdp *obj;
	int err;

	obj = (struct xdp *)calloc(1, sizeof(*obj));
	if (!obj) {
		errno = ENOMEM;
		return NULL;
	}

	err = xdp__create_skeleton(obj);
	if (err)
		goto err_out;

	err = bpf_object__open_skeleton(obj->skeleton, opts);
	if (err)
		goto err_out;

	return obj;
err_out:
	xdp__destroy(obj);
	errno = -err;
	return NULL;
}

static inline struct xdp *
xdp__open(void)
{
	return xdp__open_opts(NULL);
}

static inline int
xdp__load(struct xdp *obj)
{
	return bpf_object__load_skeleton(obj->skeleton);
}

static inline struct xdp *
xdp__open_and_load(void)
{
	struct xdp *obj;
	int err;

	obj = xdp__open();
	if (!obj)
		return NULL;
	err = xdp__load(obj);
	if (err) {
		xdp__destroy(obj);
		errno = -err;
		return NULL;
	}
	return obj;
}

static inline int
xdp__attach(struct xdp *obj)
{
	return bpf_object__attach_skeleton(obj->skeleton);
}

static inline void
xdp__detach(struct xdp *obj)
{
	bpf_object__detach_skeleton(obj->skeleton);
}

static inline const void *xdp__elf_bytes(size_t *sz);

static inline int
xdp__create_skeleton(struct xdp *obj)
{
	struct bpf_object_skeleton *s;
	int err;

	s = (struct bpf_object_skeleton *)calloc(1, sizeof(*s));
	if (!s)	{
		err = -ENOMEM;
		goto err;
	}

	s->sz = sizeof(*s);
	s->name = "xdp";
	s->obj = &obj->obj;

	/* maps */
	s->map_cnt = 1;
	s->map_skel_sz = sizeof(*s->maps);
	s->maps = (struct bpf_map_skeleton *)calloc(s->map_cnt, s->map_skel_sz);
	if (!s->maps) {
		err = -ENOMEM;
		goto err;
	}

	s->maps[0].name = "xdp_stats_map";
	s->maps[0].map = &obj->maps.xdp_stats_map;

	/* programs */
	s->prog_cnt = 1;
	s->prog_skel_sz = sizeof(*s->progs);
	s->progs = (struct bpf_prog_skeleton *)calloc(s->prog_cnt, s->prog_skel_sz);
	if (!s->progs) {
		err = -ENOMEM;
		goto err;
	}

	s->progs[0].name = "xdp_prog_func";
	s->progs[0].prog = &obj->progs.xdp_prog_func;
	s->progs[0].link = &obj->links.xdp_prog_func;

	s->data = xdp__elf_bytes(&s->data_sz);

	obj->skeleton = s;
	return 0;
err:
	bpf_object__destroy_skeleton(s);
	return err;
}

static inline const void *xdp__elf_bytes(size_t *sz)
{
	static const char data[] __attribute__((__aligned__(8))) = "\
\x7f\x45\x4c\x46\x02\x01\x01\0\0\0\0\0\0\0\0\0\x01\0\xf7\0\x01\0\0\0\0\0\0\0\0\
\0\0\0\0\0\0\0\0\0\0\0\xb8\x0c\0\0\0\0\0\0\0\0\0\0\x40\0\0\0\0\0\x40\0\x0d\0\
\x01\0\x61\x12\x04\0\0\0\0\0\x61\x11\0\0\0\0\0\0\xbf\x13\0\0\0\0\0\0\x07\x03\0\
\0\x0e\0\0\0\x2d\x23\x1c\0\0\0\0\0\x69\x13\x0c\0\0\0\0\0\x55\x03\x1a\0\x08\0\0\
\0\xbf\x13\0\0\0\0\0\0\x07\x03\0\0\x22\0\0\0\x2d\x23\x17\0\0\0\0\0\xb7\x02\0\0\
\x0c\0\0\0\x0f\x21\0\0\0\0\0\0\x61\x11\x0e\0\0\0\0\0\x63\x1a\xfc\xff\0\0\0\0\
\xbf\xa2\0\0\0\0\0\0\x07\x02\0\0\xfc\xff\xff\xff\x18\x01\0\0\0\0\0\0\0\0\0\0\0\
\0\0\0\x85\0\0\0\x01\0\0\0\x55\0\x0b\0\0\0\0\0\xb7\x01\0\0\x01\0\0\0\x63\x1a\
\xf8\xff\0\0\0\0\xbf\xa2\0\0\0\0\0\0\x07\x02\0\0\xfc\xff\xff\xff\xbf\xa3\0\0\0\
\0\0\0\x07\x03\0\0\xf8\xff\xff\xff\x18\x01\0\0\0\0\0\0\0\0\0\0\0\0\0\0\xb7\x04\
\0\0\0\0\0\0\x85\0\0\0\x02\0\0\0\x05\0\x02\0\0\0\0\0\xb7\x01\0\0\x01\0\0\0\xc3\
\x10\0\0\0\0\0\0\xb7\0\0\0\x02\0\0\0\x95\0\0\0\0\0\0\0\x44\x75\x61\x6c\x20\x4d\
\x49\x54\x2f\x47\x50\x4c\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\
\0\0\0\0\0\0\0\0\0\x9f\xeb\x01\0\x18\0\0\0\0\0\0\0\x58\x03\0\0\x58\x03\0\0\x75\
\x03\0\0\0\0\0\0\0\0\0\x02\x03\0\0\0\x01\0\0\0\0\0\0\x01\x04\0\0\0\x20\0\0\x01\
\0\0\0\0\0\0\0\x03\0\0\0\0\x02\0\0\0\x04\0\0\0\x09\0\0\0\x05\0\0\0\0\0\0\x01\
\x04\0\0\0\x20\0\0\0\0\0\0\0\0\0\0\x02\x06\0\0\0\0\0\0\0\0\0\0\x03\0\0\0\0\x02\
\0\0\0\x04\0\0\0\x10\0\0\0\0\0\0\0\0\0\0\x02\x08\0\0\0\x19\0\0\0\0\0\0\x08\x09\
\0\0\0\x1f\0\0\0\0\0\0\x01\x04\0\0\0\x20\0\0\0\0\0\0\0\x04\0\0\x04\x20\0\0\0\
\x2c\0\0\0\x01\0\0\0\0\0\0\0\x31\0\0\0\x05\0\0\0\x40\0\0\0\x3d\0\0\0\x07\0\0\0\
\x80\0\0\0\x41\0\0\0\x07\0\0\0\xc0\0\0\0\x47\0\0\0\0\0\0\x0e\x0a\0\0\0\x01\0\0\
\0\0\0\0\0\0\0\0\x02\x0d\0\0\0\x55\0\0\0\x06\0\0\x04\x18\0\0\0\x5c\0\0\0\x08\0\
\0\0\0\0\0\0\x61\0\0\0\x08\0\0\0\x20\0\0\0\x6a\0\0\0\x08\0\0\0\x40\0\0\0\x74\0\
\0\0\x08\0\0\0\x60\0\0\0\x84\0\0\0\x08\0\0\0\x80\0\0\0\x93\0\0\0\x08\0\0\0\xa0\
\0\0\0\0\0\0\0\x01\0\0\x0d\x02\0\0\0\xa2\0\0\0\x0c\0\0\0\xa6\0\0\0\x01\0\0\x0c\
\x0e\0\0\0\x65\x01\0\0\x03\0\0\x04\x0e\0\0\0\x6c\x01\0\0\x12\0\0\0\0\0\0\0\x73\
\x01\0\0\x12\0\0\0\x30\0\0\0\x7c\x01\0\0\x13\0\0\0\x60\0\0\0\x84\x01\0\0\0\0\0\
\x01\x01\0\0\0\x08\0\0\0\0\0\0\0\0\0\0\x03\0\0\0\0\x11\0\0\0\x04\0\0\0\x06\0\0\
\0\x92\x01\0\0\0\0\0\x08\x14\0\0\0\x99\x01\0\0\0\0\0\x08\x15\0\0\0\x9f\x01\0\0\
\0\0\0\x01\x02\0\0\0\x10\0\0\0\xde\x01\0\0\x0a\0\0\x84\x14\0\0\0\xe4\x01\0\0\
\x17\0\0\0\0\0\0\x04\xe8\x01\0\0\x17\0\0\0\x04\0\0\x04\xf0\x01\0\0\x17\0\0\0\
\x08\0\0\0\xf4\x01\0\0\x13\0\0\0\x10\0\0\0\xfc\x01\0\0\x13\0\0\0\x20\0\0\0\xff\
\x01\0\0\x13\0\0\0\x30\0\0\0\x08\x02\0\0\x17\0\0\0\x40\0\0\0\x0c\x02\0\0\x17\0\
\0\0\x48\0\0\0\x15\x02\0\0\x18\0\0\0\x50\0\0\0\0\0\0\0\x19\0\0\0\x60\0\0\0\x1b\
\x02\0\0\0\0\0\x08\x11\0\0\0\x20\x02\0\0\0\0\0\x08\x14\0\0\0\0\0\0\0\x02\0\0\
\x05\x08\0\0\0\0\0\0\0\x1a\0\0\0\0\0\0\0\x28\x02\0\0\x1c\0\0\0\0\0\0\0\0\0\0\0\
\x02\0\0\x04\x08\0\0\0\x2e\x02\0\0\x1b\0\0\0\0\0\0\0\x34\x02\0\0\x1b\0\0\0\x20\
\0\0\0\x3a\x02\0\0\0\0\0\x08\x08\0\0\0\0\0\0\0\x02\0\0\x04\x08\0\0\0\x2e\x02\0\
\0\x1b\0\0\0\0\0\0\0\x34\x02\0\0\x1b\0\0\0\x20\0\0\0\x58\x03\0\0\0\0\0\x01\x01\
\0\0\0\x08\0\0\x01\0\0\0\0\0\0\0\x03\0\0\0\0\x1d\0\0\0\x04\0\0\0\x0d\0\0\0\x5d\
\x03\0\0\0\0\0\x0e\x1e\0\0\0\x01\0\0\0\x67\x03\0\0\x01\0\0\x0f\0\0\0\0\x0b\0\0\
\0\0\0\0\0\x20\0\0\0\x6d\x03\0\0\x01\0\0\x0f\0\0\0\0\x1f\0\0\0\0\0\0\0\x0d\0\0\
\0\0\x69\x6e\x74\0\x5f\x5f\x41\x52\x52\x41\x59\x5f\x53\x49\x5a\x45\x5f\x54\x59\
\x50\x45\x5f\x5f\0\x5f\x5f\x75\x33\x32\0\x75\x6e\x73\x69\x67\x6e\x65\x64\x20\
\x69\x6e\x74\0\x74\x79\x70\x65\0\x6d\x61\x78\x5f\x65\x6e\x74\x72\x69\x65\x73\0\
\x6b\x65\x79\0\x76\x61\x6c\x75\x65\0\x78\x64\x70\x5f\x73\x74\x61\x74\x73\x5f\
\x6d\x61\x70\0\x78\x64\x70\x5f\x6d\x64\0\x64\x61\x74\x61\0\x64\x61\x74\x61\x5f\
\x65\x6e\x64\0\x64\x61\x74\x61\x5f\x6d\x65\x74\x61\0\x69\x6e\x67\x72\x65\x73\
\x73\x5f\x69\x66\x69\x6e\x64\x65\x78\0\x72\x78\x5f\x71\x75\x65\x75\x65\x5f\x69\
\x6e\x64\x65\x78\0\x65\x67\x72\x65\x73\x73\x5f\x69\x66\x69\x6e\x64\x65\x78\0\
\x63\x74\x78\0\x78\x64\x70\x5f\x70\x72\x6f\x67\x5f\x66\x75\x6e\x63\0\x78\x64\
\x70\0\x30\x3a\x31\0\x2f\x72\x6f\x6f\x74\x2f\x67\x6f\x2f\x73\x72\x63\x2f\x65\
\x62\x70\x66\x2d\x67\x6f\x2f\x73\x74\x65\x70\x31\x32\x5f\x78\x64\x70\x2f\x78\
\x64\x70\x2e\x63\0\x09\x76\x6f\x69\x64\x20\x2a\x64\x61\x74\x61\x5f\x65\x6e\x64\
\x20\x3d\x20\x28\x76\x6f\x69\x64\x20\x2a\x29\x28\x6c\x6f\x6e\x67\x29\x63\x74\
\x78\x2d\x3e\x64\x61\x74\x61\x5f\x65\x6e\x64\x3b\0\x30\x3a\x30\0\x09\x76\x6f\
\x69\x64\x20\x2a\x64\x61\x74\x61\x20\x20\x20\x20\x20\x3d\x20\x28\x76\x6f\x69\
\x64\x20\x2a\x29\x28\x6c\x6f\x6e\x67\x29\x63\x74\x78\x2d\x3e\x64\x61\x74\x61\
\x3b\0\x09\x69\x66\x20\x28\x28\x76\x6f\x69\x64\x20\x2a\x29\x28\x65\x74\x68\x20\
\x2b\x20\x31\x29\x20\x3e\x20\x64\x61\x74\x61\x5f\x65\x6e\x64\x29\x20\x7b\0\x65\
\x74\x68\x68\x64\x72\0\x68\x5f\x64\x65\x73\x74\0\x68\x5f\x73\x6f\x75\x72\x63\
\x65\0\x68\x5f\x70\x72\x6f\x74\x6f\0\x75\x6e\x73\x69\x67\x6e\x65\x64\x20\x63\
\x68\x61\x72\0\x5f\x5f\x62\x65\x31\x36\0\x5f\x5f\x75\x31\x36\0\x75\x6e\x73\x69\
\x67\x6e\x65\x64\x20\x73\x68\x6f\x72\x74\0\x30\x3a\x32\0\x09\x69\x66\x20\x28\
\x65\x74\x68\x2d\x3e\x68\x5f\x70\x72\x6f\x74\x6f\x20\x21\x3d\x20\x62\x70\x66\
\x5f\x68\x74\x6f\x6e\x73\x28\x45\x54\x48\x5f\x50\x5f\x49\x50\x29\x29\x20\x7b\0\
\x69\x70\x68\x64\x72\0\x69\x68\x6c\0\x76\x65\x72\x73\x69\x6f\x6e\0\x74\x6f\x73\
\0\x74\x6f\x74\x5f\x6c\x65\x6e\0\x69\x64\0\x66\x72\x61\x67\x5f\x6f\x66\x66\0\
\x74\x74\x6c\0\x70\x72\x6f\x74\x6f\x63\x6f\x6c\0\x63\x68\x65\x63\x6b\0\x5f\x5f\
\x75\x38\0\x5f\x5f\x73\x75\x6d\x31\x36\0\x61\x64\x64\x72\x73\0\x73\x61\x64\x64\
\x72\0\x64\x61\x64\x64\x72\0\x5f\x5f\x62\x65\x33\x32\0\x30\x3a\x39\x3a\x30\x3a\
\x30\0\x09\x2a\x69\x70\x5f\x73\x72\x63\x5f\x61\x64\x64\x72\x20\x3d\x20\x28\x5f\
\x5f\x75\x33\x32\x29\x28\x69\x70\x2d\x3e\x73\x61\x64\x64\x72\x29\x3b\0\x09\x5f\
\x5f\x75\x33\x32\x20\x2a\x70\x6b\x74\x5f\x63\x6f\x75\x6e\x74\x20\x3d\x20\x62\
\x70\x66\x5f\x6d\x61\x70\x5f\x6c\x6f\x6f\x6b\x75\x70\x5f\x65\x6c\x65\x6d\x28\
\x26\x78\x64\x70\x5f\x73\x74\x61\x74\x73\x5f\x6d\x61\x70\x2c\x20\x26\x69\x70\
\x29\x3b\0\x09\x69\x66\x20\x28\x21\x70\x6b\x74\x5f\x63\x6f\x75\x6e\x74\x29\x20\
\x7b\0\x09\x09\x5f\x5f\x75\x33\x32\x20\x69\x6e\x69\x74\x5f\x70\x6b\x74\x5f\x63\
\x6f\x75\x6e\x74\x20\x3d\x20\x31\x3b\0\x09\x09\x62\x70\x66\x5f\x6d\x61\x70\x5f\
\x75\x70\x64\x61\x74\x65\x5f\x65\x6c\x65\x6d\x28\x26\x78\x64\x70\x5f\x73\x74\
\x61\x74\x73\x5f\x6d\x61\x70\x2c\x20\x26\x69\x70\x2c\x20\x26\x69\x6e\x69\x74\
\x5f\x70\x6b\x74\x5f\x63\x6f\x75\x6e\x74\x2c\x20\x42\x50\x46\x5f\x41\x4e\x59\
\x29\x3b\0\x09\x09\x5f\x5f\x73\x79\x6e\x63\x5f\x66\x65\x74\x63\x68\x5f\x61\x6e\
\x64\x5f\x61\x64\x64\x28\x70\x6b\x74\x5f\x63\x6f\x75\x6e\x74\x2c\x20\x31\x29\
\x3b\0\x09\x72\x65\x74\x75\x72\x6e\x20\x58\x44\x50\x5f\x50\x41\x53\x53\x3b\0\
\x63\x68\x61\x72\0\x5f\x5f\x6c\x69\x63\x65\x6e\x73\x65\0\x2e\x6d\x61\x70\x73\0\
\x6c\x69\x63\x65\x6e\x73\x65\0\0\0\0\x9f\xeb\x01\0\x20\0\0\0\0\0\0\0\x14\0\0\0\
\x14\0\0\0\x0c\x01\0\0\x20\x01\0\0\x4c\0\0\0\x08\0\0\0\xb4\0\0\0\x01\0\0\0\0\0\
\0\0\x0f\0\0\0\x10\0\0\0\xb4\0\0\0\x10\0\0\0\0\0\0\0\xbc\0\0\0\xe2\0\0\0\x26\
\x6c\0\0\x08\0\0\0\xbc\0\0\0\x15\x01\0\0\x26\x70\0\0\x10\0\0\0\xbc\0\0\0\x40\
\x01\0\0\x06\x80\0\0\x20\0\0\0\xbc\0\0\0\x40\x01\0\0\x06\x80\0\0\x28\0\0\0\xbc\
\0\0\0\xb2\x01\0\0\x0b\x90\0\0\x30\0\0\0\xbc\0\0\0\xb2\x01\0\0\x06\x90\0\0\x60\
\0\0\0\xbc\0\0\0\x49\x02\0\0\x1d\xc0\0\0\x68\0\0\0\xbc\0\0\0\x49\x02\0\0\x0f\
\xc0\0\0\x78\0\0\0\xbc\0\0\0\x49\x02\0\0\x1d\xc0\0\0\x80\0\0\0\xbc\0\0\0\x6d\
\x02\0\0\x15\xf0\0\0\x98\0\0\0\xbc\0\0\0\xab\x02\0\0\x06\xf4\0\0\xa8\0\0\0\xbc\
\0\0\0\xbe\x02\0\0\x09\xfc\0\0\xb8\0\0\0\xbc\0\0\0\0\0\0\0\0\0\0\0\xd0\0\0\0\
\xbc\0\0\0\xda\x02\0\0\x03\0\x01\0\0\x01\0\0\xbc\0\0\0\x20\x03\0\0\x03\x10\x01\
\0\x08\x01\0\0\xbc\0\0\0\x46\x03\0\0\x02\x24\x01\0\x10\0\0\0\xb4\0\0\0\x04\0\0\
\0\0\0\0\0\x0d\0\0\0\xb8\0\0\0\0\0\0\0\x08\0\0\0\x0d\0\0\0\x11\x01\0\0\0\0\0\0\
\x28\0\0\0\x10\0\0\0\xae\x01\0\0\0\0\0\0\x50\0\0\0\x16\0\0\0\x41\x02\0\0\0\0\0\
\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x03\0\x03\0\
\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x6f\0\0\0\0\0\x03\0\x08\x01\0\0\0\0\0\0\0\0\0\
\0\0\0\0\0\x76\0\0\0\0\0\x03\0\xf8\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x48\0\0\0\x12\
\0\x03\0\0\0\0\0\0\0\0\0\x18\x01\0\0\0\0\0\0\x22\0\0\0\x11\0\x06\0\0\0\0\0\0\0\
\0\0\x20\0\0\0\0\0\0\0\x3e\0\0\0\x11\0\x05\0\0\0\0\0\0\0\0\0\x0d\0\0\0\0\0\0\0\
\x80\0\0\0\0\0\0\0\x01\0\0\0\x05\0\0\0\xd0\0\0\0\0\0\0\0\x01\0\0\0\x05\0\0\0\
\x50\x03\0\0\0\0\0\0\x04\0\0\0\x05\0\0\0\x68\x03\0\0\0\0\0\0\x04\0\0\0\x06\0\0\
\0\x2c\0\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\x40\0\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\
\x50\0\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\x60\0\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\
\x70\0\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\x80\0\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\
\x90\0\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\xa0\0\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\
\xb0\0\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\xc0\0\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\
\xd0\0\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\xe0\0\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\
\xf0\0\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\0\x01\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\
\x10\x01\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\x20\x01\0\0\0\0\0\0\x04\0\0\0\x01\0\0\
\0\x30\x01\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\x4c\x01\0\0\0\0\0\0\x04\0\0\0\x01\0\
\0\0\x5c\x01\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\x6c\x01\0\0\0\0\0\0\x04\0\0\0\x01\
\0\0\0\x7c\x01\0\0\0\0\0\0\x04\0\0\0\x01\0\0\0\x0d\x0f\x0e\0\x2e\x74\x65\x78\
\x74\0\x2e\x72\x65\x6c\x2e\x42\x54\x46\x2e\x65\x78\x74\0\x2e\x6d\x61\x70\x73\0\
\x2e\x72\x65\x6c\x78\x64\x70\0\x78\x64\x70\x5f\x73\x74\x61\x74\x73\x5f\x6d\x61\
\x70\0\x2e\x6c\x6c\x76\x6d\x5f\x61\x64\x64\x72\x73\x69\x67\0\x5f\x5f\x6c\x69\
\x63\x65\x6e\x73\x65\0\x78\x64\x70\x5f\x70\x72\x6f\x67\x5f\x66\x75\x6e\x63\0\
\x2e\x73\x74\x72\x74\x61\x62\0\x2e\x73\x79\x6d\x74\x61\x62\0\x2e\x72\x65\x6c\
\x2e\x42\x54\x46\0\x4c\x42\x42\x30\x5f\x36\0\x4c\x42\x42\x30\x5f\x35\0\0\0\0\0\
\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\
\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x56\0\0\0\x03\0\0\0\0\0\0\0\0\0\0\0\
\0\0\0\0\0\0\0\0\x3b\x0c\0\0\0\0\0\0\x7d\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x01\0\0\
\0\0\0\0\0\0\0\0\0\0\0\0\0\x01\0\0\0\x01\0\0\0\x06\0\0\0\0\0\0\0\0\0\0\0\0\0\0\
\0\x40\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x04\0\0\0\0\0\0\0\0\0\0\0\
\0\0\0\0\x1e\0\0\0\x01\0\0\0\x06\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x40\0\0\0\0\0\0\
\0\x18\x01\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x08\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x1a\0\
\0\0\x09\0\0\0\x40\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\xa8\x0a\0\0\0\0\0\0\x20\0\0\0\
\0\0\0\0\x0c\0\0\0\x03\0\0\0\x08\0\0\0\0\0\0\0\x10\0\0\0\0\0\0\0\x40\0\0\0\x01\
\0\0\0\x03\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x58\x01\0\0\0\0\0\0\x0d\0\0\0\0\0\0\0\
\0\0\0\0\0\0\0\0\x01\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x14\0\0\0\x01\0\0\0\x03\0\0\
\0\0\0\0\0\0\0\0\0\0\0\0\0\x68\x01\0\0\0\0\0\0\x20\0\0\0\0\0\0\0\0\0\0\0\0\0\0\
\0\x08\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x6a\0\0\0\x01\0\0\0\0\0\0\0\0\0\0\0\0\0\0\
\0\0\0\0\0\x88\x01\0\0\0\0\0\0\xe5\x06\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x04\0\0\0\0\
\0\0\0\0\0\0\0\0\0\0\0\x66\0\0\0\x09\0\0\0\x40\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\
\xc8\x0a\0\0\0\0\0\0\x20\0\0\0\0\0\0\0\x0c\0\0\0\x07\0\0\0\x08\0\0\0\0\0\0\0\
\x10\0\0\0\0\0\0\0\x0b\0\0\0\x01\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x70\x08\
\0\0\0\0\0\0\x8c\x01\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x04\0\0\0\0\0\0\0\0\0\0\0\0\0\
\0\0\x07\0\0\0\x09\0\0\0\x40\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\xe8\x0a\0\0\0\0\0\0\
\x50\x01\0\0\0\0\0\0\x0c\0\0\0\x09\0\0\0\x08\0\0\0\0\0\0\0\x10\0\0\0\0\0\0\0\
\x30\0\0\0\x03\x4c\xff\x6f\0\0\0\x80\0\0\0\0\0\0\0\0\0\0\0\0\x38\x0c\0\0\0\0\0\
\0\x03\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x01\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x5e\0\0\
\0\x02\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\x0a\0\0\0\0\0\0\xa8\0\0\0\0\0\0\
\0\x01\0\0\0\x04\0\0\0\x08\0\0\0\0\0\0\0\x18\0\0\0\0\0\0\0";

	*sz = sizeof(data) - 1;
	return (const void *)data;
}

#ifdef __cplusplus
struct xdp *xdp::open(const struct bpf_object_open_opts *opts) { return xdp__open_opts(opts); }
struct xdp *xdp::open_and_load() { return xdp__open_and_load(); }
int xdp::load(struct xdp *skel) { return xdp__load(skel); }
int xdp::attach(struct xdp *skel) { return xdp__attach(skel); }
void xdp::detach(struct xdp *skel) { xdp__detach(skel); }
void xdp::destroy(struct xdp *skel) { xdp__destroy(skel); }
const void *xdp::elf_bytes(size_t *sz) { return xdp__elf_bytes(sz); }
#endif /* __cplusplus */

__attribute__((unused)) static void
xdp__assert(struct xdp *s __attribute__((unused)))
{
#ifdef __cplusplus
#define _Static_assert static_assert
#endif
#ifdef __cplusplus
#undef _Static_assert
#endif
}

#endif /* __XDP_SKEL_H__ */
