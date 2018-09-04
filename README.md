init project
##### 学习目的，简单实现docker基础功能
- centos7
- Linux version 4.17.13-1.el7.x86_64 (mockbuild@buildbox.spaceduck.org) (gcc version 4.8.5 20150623 (Red Hat 4.8.5-28) (GCC))
- branch dev-1.0 简单实现容器
- branch dev-1.1 添加cgroup
- branch dev-2.0 添加imagelayer(aufs)
- branch dev-2.1 添加detach and add commands like ps,logs,exec,stop,rm 


##### dependences
- yum install kernel-ml-aufs
- refer https://github.com/bnied/kernel-ml-aufs

##### troubleshooting
- XFS and AUFS 不兼容问题
- mount -t tmpfs -o size=200M tmpfs /tmp

- open the ip_forward
- echo '1' > /proc/sys/net/ipv4/ip_forward