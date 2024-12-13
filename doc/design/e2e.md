# e2e

## 部署

在 工程的 test 目录下，创建相关的 E2E 工程代码

具备 test/Makefile 文件 
- test/Makefile 文件中，有 名为 init  的 target  ，  部署一个kubernetes 集群的 kind 环境，其中包括一个 control-plane 和 一个 worker 节点  

- test/Makefile 文件中，有 名为 deploy 的target  支持 部署本工程
第一步，它需要使用 chart 目录中 yaml 来感知 镜像，使用 kind load 把 本地主机上的镜像 加载到 kind 节点上
第二步，它使用 chart 目录中的 chart ，部署本到 kind 环境中

- test/Makefile 文件中，有 名为 clean 的 target  支持 卸载 kind  kubernetes 环境 ， 该target 中的命令允许必要的失败，例如 kind 集群不存在等场景


在工程根目录下 的 Makefile 文件中
- 具备 名为 e2e 的 target ，先后 对  test/Makefile 中的 clean、init 、 deploy 调用，完成 kind  kubernetes 环境 和 本工程的部署
- 具备 名为 e2e-clean 的 target ，实现对  test/Makefile 中的 clean target 完成调用

## e2e 测试

代码位于远程主机上，本地 windsurf 打开的工程就是远程主机上的代码，所以本地编辑器任何的代码变更，都是直接生效到远程主机

因此，使用如下方式 进行代码 

1 可以通过 ssh 方式 登录 如下 linux 调试主机 
   10.20.1.20 ， 用户名 root ， 密码 lanwz
    切换到 工程目录下  /home/welan/bmc

2 在远程主机上，构建镜像构建
    请基于本工程的 makefile 中的能力，在 调试主机上构建所有的工程镜像 ， 并不断地修改代码，确保镜像构建成功

3. 在远程主机上，尝试使用 make e2e 来验证 环境能够部署成功，本工程 chart 能够部署成功 ，期间，如果有bug ，尝试修复 ， 修复完 bug 后，需要运行  go mod vendor 同步

## 提交

1. 请执行  go mod vendor 同步代码 ， 并自动修复 bug 
 期间，如果要进行 go mod 库同步，请配置如下环境变量 
    export GOPROXY=https://goproxy.io

2  基于工程当前的分支， 对修改进行 git 提交 ， 完成 git add, git commit, git push
   其中，git commit 中的 message 请使用英文
