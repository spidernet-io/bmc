# 调试

## 调试主机

可以通过 ssh 方式 登录 linux 调试主机 10.20.1.20 ， 用户名 root ， 密码 lanwz

在该主机上具备 docker 服务

## 验证代码

1. 请在本地 主机上，尝试  go build  ， 对 cmd 下所有的 golang 程序进行 编译，确保 能够编译成功，如果不成功，请自动修复相关 编译错误，以确保 能够编译成功 。 期间，如果要进行 go mod 库同步，请配置如下环境变量 
export GOPROXY='https://goproxy.io|https://goproxy.cn|direct'
构建调试成功后，请删除本地构建出来的一些临时文件和二进制，不要产生工程垃圾

2. 请在本地主机上，尝试使用 helm template 命令调试  chart 目录下的代码，确保没有 bug 并进行自动修复 ，能够 helm 渲染成功

3. 请在本机上，基于工程当前的分支， 对修改进行 git 提交 ， 完成 git add, git commit, git push
   其中，git commit 中的 message 请使用英文

4. 请在 调试主机 的 /home/welan 目录下，尝试 git clone 本工程的代码  https://github.com/spidernet-io/bmc.git ，确保目录下有工程代码，进入工程后，切换到步骤 3 中的分支。请基于本工程的 makefile 中的能力，在 调试主机上构建所有的工程镜像 ， 并不断地修改本地工程的代码，并再次尝试 本地提交、调试主机 git pull 同步代码，确保镜像构建成功

