# 调试

## 调试主机

可以通过 ssh 方式 登录调试主机 10.20.1.20 ， 用户名 root ， 密码 lanwz

在该主机上具备 docker 服务

## 同步代码

1. 请在本机上，基于工程当前的分支， 对修改进行 git 提交 ， 完成 git add, git commit, git push
   其中，git commit 中的 message 请使用英文

2. 请在 调试主机 的 /home/welan 目录下，git clone 本工程的代码  https://github.com/spidernet-io/bmc.git ，并切换到步骤1 中的分支。请基于本工程的 makefile 中的能力，在 调试主机上构建所有的工程镜像 ， 并不断地修改本地工程的代码，并再次尝试 本地提交、调试主机 git pull 同步代码，确保镜像构建成功





