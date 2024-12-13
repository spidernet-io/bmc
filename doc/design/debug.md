# 调试

## 提交代码


## 调试主机

可以通过 ssh 方式 登录调试主机 10.20.1.20 ， 用户名 root ， 密码 lanwz

在该主机上具备 docker 服务

## 同步代码

1. 请在本机上，工程当前的修改进行 git 提交

2. 请在 调试主机 的 /home/welan 目录下，git clone 本工程的代码  https://github.com/spidernet-io/bmc.git

3. 请基于本工程的 makefile 中的能力，在 调试主机 上构建所有的工程镜像 ， 并不断地修改本地工程的代码，并再次尝试 本体提交、调试主机 同步代码，确保镜像构建成功





