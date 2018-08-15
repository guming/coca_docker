init project
##### 学习目的，简单实现docker基础功能
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