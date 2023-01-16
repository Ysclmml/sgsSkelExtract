### 使用方法:
1. 执行可执行文件需要指定三国杀路径, 默认为当前目录名为sgs
2. 执行完成之后, 会生成skins和final两个文件夹. skins里面放入了未进行命名的原始动皮数据. 
final存放了可供十周年UI使用的动皮数据以及动皮相关的静皮和技能特效,技能语音.自行查看即可.
3. 当前路径下还会生成一个xxx_新.xlsx的文件, 当有新版本的三国杀出来后, 重新执行一遍程序, 会生成当前文件. 它会与原来老的文件进行比对, 当有新的动皮出现的时候
会生成空的一行. 自行补充到老的文件即可.
4. 当前路径下还会生成一个dynamicSkin_new.js文件. 这个会与老的dynamicSkin.js文件进行比对, 会生成新的十周年动皮参数, 自行复制到十周年ui即可.

config.yml文件有生成的默认动皮的待机攻击和背景参数, 可以自行调整.
#### linux:
```shell
./linux-run -apkPath=三国杀路径
```
#### windows:
```shell
.\window-run.exe -apkPath D:\Game\三国杀\手杀\sgs412
```

### 编译
需要go环境1.18以上. 
```shell
# linux 
go build -o linux-run .

# windows
go build -o window-run.exe . 
```

