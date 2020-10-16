wget https://raw.githubusercontent.com/1265578519/v2ray-core/main/release/v2ray-start.txt -O /etc/init.d/v2ray
chmod a+x /etc/init.d/v2ray
chkconfig v2ray on
service v2ray start
systemctl stop firewalld;systemctl disable firewalld;service ip6tables stop;service iptables stop;chkconfig iptables off
wget https://raw.githubusercontent.com/1265578519/v2ray-core/main/release/wsconfig.json -O /etc/v2ray/config.json
service v2ray restart
cat /var/log/v2ray/error.log | grep started
