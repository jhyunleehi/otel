## how to get uid string 

```c
#include <linux/bpf.h>
#include <linux/pid.h>
#include <linux/fs.h>
#include <linux/sched.h>

SEC("lsm/inode_permission")
int bpf_inode_permission(struct inode *inode, int mask) {
    struct task_struct *task = (struct task_struct *)bpf_get_current_task();
    
    // Retrieve UID and GID
    kuid_t uid = task->cred->uid;
    kgid_t gid = task->cred->gid;

    // Convert UID and GID to string representations
    char uid_str[16]; // Adjust size according to your needs
    char gid_str[16];
    snprintf(uid_str, sizeof(uid_str), "%d", from_kuid(&init_user_ns, uid));
    snprintf(gid_str, sizeof(gid_str), "%d", from_kgid(&init_user_ns, gid));

    // Do something with UID and GID strings
    printk("UID: %s, GID: %s\n", uid_str, gid_str);

    return 0;
}

char _license[] SEC("license") = "GPL";

```

## user uid string by golang 

```go
mport (
    "fmt"
    "os"
    "os/user"
)

func main() {
    // Get current user's UID
    uid := os.Getuid()

    // Get current user's GID
    gid := os.Getgid()

    // Print UID and GID
    fmt.Printf("UID: %d\n", uid)
    fmt.Printf("GID: %d\n", gid)

    // Get username from UID
    u, err := user.LookupId(fmt.Sprint(uid))
    if err == nil {
        fmt.Printf("Username: %s\n", u.Username)
    } else {
        fmt.Printf("Error getting username: %v\n", err)
    }
}
```