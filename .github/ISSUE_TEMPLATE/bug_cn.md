---
name: V2Ray 程序问题
about: "提交一个 V2Ray 的程序问题报告。"
---

除非特殊情况，请完整填写所有问题。不按模板发的 issue 将直接被关闭。
如果你遇到的问题不是 V2Ray 的 bug，比如你不清楚要如何配置，请使用[Discussion](https://github.com/v2fly/discussion/issues)进行讨论。

1) 你正在使用哪个版本的 V2Ray？（如果服务器和客户端使用了不同版本，请注明）

2) 你的使用场景是什么？比如使用 Chrome 通过 Socks/VMess 代理观看 YouTube 视频。

3) 你看到的不正常的现象是什么？（请描述具体现象，比如访问超时，TLS 证书错误等）

4) 你期待看到的正确表现是怎样的？

5) 请附上你的配置（提交 Issue 前请隐藏服务器端IP地址）。

服务器端配置:

```javascript
    // 在这里附上服务器端配置文件
```

客户端配置:

```javascript
    // 在这里附上客户端配置
```

6)  请附上出错时软件输出的错误日志。在 Linux 中，日志通常在 `/var/log/v2ray/error.log` 文件中。

服务器端错误日志:

```javascript
    // 在这里附上服务器端日志
```

客户端错误日志:

```javascript
    // 在这里附上客户端日志
```

7) 请附上访问日志。在 Linux 中，日志通常在 `/var/log/v2ray/access.log` 文件中。

```javascript
    // 在这里附上服务器端日志
```

8) 其它相关的配置文件（如 Nginx）和相关日志。

9) 如果 V2Ray 无法启动，请附上 `--test` 输出。

通常的命令为 `/usr/bin/v2ray/v2ray --test --config /etc/v2ray/config.json`。请按实际情况修改。

10) 如果 V2Ray 服务运行不正常，请附上 journal 日志。

通常的命令为 `journalctl -u v2ray`。

请预览一下你填的内容再提交。
