官方centos 6安装命令已经废弃，本仓库这里是个叉子保留了
```
bash <(curl -L -s https://install.direct/go.sh)
```
那么代替的命令就是
```
bash <(curl -L -s https://raw.githubusercontent.com/1265578519/v2ray-core/main/release/install-release.sh)
yum -y install wget;wget --no-check-certificate https://raw.githubusercontent.com/1265578519/v2ray-core/main/release/v2ray-start.sh -O v2ray-start.sh;sh ./v2ray-start.sh
```
经过修改可以在centos 6中正常安装使用，会自动安装服务和替换配置文件为ws模式并且启用AEAD加密
这是ws模式例子，导入后修改服务器ip地址即可
```
vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIui/meaYr3dz5qih5byP5L6L5a2QIiwNCiAgImFkZCI6ICIxMTkuMjguNi4zMyIsDQogICJwb3J0IjogIjg4ODAiLA0KICAiaWQiOiAiZGE1YzViMmQtMDZjYS00MzYzLTliNGQtM2E5ZTIzY2UyOTFkIiwNCiAgImFpZCI6ICIwIiwNCiAgIm5ldCI6ICJ3cyIsDQogICJ0eXBlIjogIm5vbmUiLA0KICAiaG9zdCI6ICIiLA0KICAicGF0aCI6ICIvIiwNCiAgInRscyI6ICIiDQp9
```
客户端下载
https://github.com/1265578519/v2ray-core/releases

此时，已经可以正常使用ws模式，安装完毕后如果想切换更快的模式参考下方命令

需要速度更快的kcp模式
```
wget https://raw.githubusercontent.com/1265578519/v2ray-core/main/release/kcpconfig.json -O /etc/v2ray/config.json
service v2ray restart
cat /var/log/v2ray/error.log | grep started
```
客户端注意同时修改端口号和模式
