/*
 * Note that this header file contains a subset of kernel 
 * definitions needed for the tcprtt_sockops example.
 */
#ifndef BPF_SOCKOPS_H
#define BPF_SOCKOPS_H

/*
 * Copy of TCP states.
 * See: https://elixir.bootlin.com/linux/latest/source/include/uapi/linux/bpf.h#L6347.
 */

/*
 * Copy of sock_ops operations.
 * See: https://elixir.bootlin.com/linux/latest/source/include/uapi/linux/bpf.h#L6233.
 */
/*
 * Copy of definitions for bpf_sock_ops_cb_flags.
 * See: https://elixir.bootlin.com/linux/latest/source/include/uapi/linux/bpf.h#L6178.
 */


/*
 * Copy of bpf.h's bpf_sock_ops with minimal subset 
 * of fields used by the tcprtt_sockops example.
 * See: https://elixir.bootlin.com/linux/latest/source/include/uapi/linux/bpf.h#L6101.
 */
#endif