官方centos 6安装命令已经废弃，本仓库这里是个叉子保留了
```
bash <(curl -L -s https://install.direct/go.sh)
```
那么代替的安装命令就是
```
bash <(curl -L -s https://raw.githubusercontent.com/1265578519/v2ray-core/main/release/install-release.sh)
yum -y install wget;wget --no-check-certificate https://raw.githubusercontent.com/1265578519/v2ray-core/main/release/v2ray-start.sh -O v2ray-start.sh;sh ./v2ray-start.sh
```
经过修改可以在centos 6中正常安装使用，会自动安装服务和替换配置文件为ws模式并且启用AEAD加密
这是ws模式例子，导入后修改服务器ip地址即可
```
vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIui/meaYr3dz5qih5byP5L6L5a2QIiwNCiAgImFkZCI6ICIxMTkuMjguNi4zMyIsDQogICJwb3J0IjogIjg4ODAiLA0KICAiaWQiOiAiZGE1YzViMmQtMDZjYS00MzYzLTliNGQtM2E5ZTIzY2UyOTFkIiwNCiAgImFpZCI6ICIwIiwNCiAgIm5ldCI6ICJ3cyIsDQogICJ0eXBlIjogIm5vbmUiLA0KICAiaG9zdCI6ICIiLA0KICAicGF0aCI6ICIvIiwNCiAgInRscyI6ICIiDQp9
```
客户端下载（游戏软件用netch，浏览器用v2rayn）
https://github.com/1265578519/v2ray-core/releases

此时，已经可以正常使用ws模式，安装完毕后如果想切换更快的模式参考下方命令

需要速度更快的kcp模式（如无需求请勿使用，因为基于udp协议传输，无视下方中所有内容）
```
wget https://raw.githubusercontent.com/1265578519/v2ray-core/main/release/kcpconfig.json -O /etc/v2ray/config.json
service v2ray restart
cat /var/log/v2ray/error.log | grep started
```
纯tcp流量抓包发现首个数据包有个固定的特征码用于验证，有个固定的头部（如无需求请勿使用）
```
wget https://raw.githubusercontent.com/1265578519/v2ray-core/main/release/tcpconfig.json -O /etc/v2ray/config.json
service v2ray restart
cat /var/log/v2ray/error.log | grep started
```
ws模式多端口，同时开放多个端口随时切换使用
```
wget https://raw.githubusercontent.com/1265578519/v2ray-core/main/release/duowsconfig.json -O /etc/v2ray/config.json
service v2ray restart
cat /var/log/v2ray/error.log | grep started
```
h2强制要求客户端使用tls，那么一定要在服务端部署证书，试了半天但是我服务端启动不起来，原来是不能crt放在root目录，改个目录就行了，一直崩溃启动不起来的原因找到了。。
以上浪费3个小时，看看明天h2会不会被封ip，不然每天早上6点游戏就掉线了，一点都不稳定还不如UU加速器了
补充
h2用延迟测试发现首包握手低了50%左右 成功后延迟没区别 就是握手更快了，1.9.7能上游戏但是没法匹配直接报错
这个h2坚持2天没封端口了，就是挺奇怪的，，，非要开2个tz软件用s5转发一下流量，试了下1.9.7虽然能连上到大厅但是没办法进行匹配游戏，用1.6.9转发到另一个tz软件就正常玩了，这样用就正常。。。
```
wget https://raw.githubusercontent.com/1265578519/v2ray-core/main/release/h2config.json -O /etc/v2ray/config.json
wget https://raw.githubusercontent.com/1265578519/v2ray-core/main/release/oss.zip -O /root/oss.zip
unzip -o oss.zip
\cp -f oss.crt /etc/v2ray/oss.crt
\cp -f oss.key /etc/v2ray/oss.key
service v2ray restart
cat /var/log/v2ray/error.log | grep started
```

客户端注意同时修改端口号和模式

卸载方法
```
service v2ray stop
chkconfig v2ray off
rm -rf /etc/init.d/v2ray
rm -rf /etc/v2ray #(配置文件)
rm -rf /usr/bin/v2ray #(程序)
rm -rf /var/log/v2ray #(日志)
```
