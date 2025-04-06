## Shua!
支持多线程，自定义休息时间的一个小工具，用于刷流量。

## 下载
### Github Release
从 [Github Release](https://github.com/ZeroWolf233/shua/releases) 下载最新版。

## 使用
### 可选flag
| 名称 | 默认值                                             | 说明            |
|----|-------------------------------------------------|---------------|
| u  | https://s3.pysio.online/pcl2-ce/PCL2_CE_x64.exe | 请求内容的地址       |
| w  | 4                                               | 创建多少个工作进程(线程) |
| i  | 0s                                              | 每次请求后的休息时长    |

### 使用示例
```bash
./shua -u https://lf5-j1gamecdn-cn.dailygn.com/obj/lf-game-lf/gdl_app_2682/1233880772355.mp4 -w 128 -i 3s
```