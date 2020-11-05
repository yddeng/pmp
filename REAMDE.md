
## Program management platform

master 主节点控制程序的启动，配置更新，文件更新。web端操作简单
对外web服务。对内控制程序启动

slave  程序启动器，守护程序，接收更新配置，通知进程重读配置

## slave 启动

连接到master 后，由 master 同步配置文件，程序。


### 通知配置更新

master 更新文件后，自动同步到各 slave 。 在master web端 手动选择程序是否需要重读配置。

slave 向目标进程发送 USER1 信号，各程序需监听该信号才能使用该功能。

### 程序守护

slave 启动程序，返回对应的 pid 。若 pid 进程宕机，slave 重启程序。


