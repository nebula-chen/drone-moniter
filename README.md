# drone-moniter
无人机监控系统后端接口

* 安装 nginx
```Shell
# 包管理器安装
sudo apt-get install nginx

配置文件目录：/etc/nginx/
默认网站根目录：/var/www/html/
日志文件：/var/log/nginx/

# 将无人机后端的nignx配置复制到 nginx 配置目录下
sudo cp ./nginx_drone.conf /etc/nginx/sites-enabled/
```

* 安装 docker-compose
```Shell
# 下载最新稳定版 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.36.1/docker-compose-$(uname -s)-$(uname -m)" -o docker-compose
sudo mv docker-compose /usr/local/bin/
# 赋予可执行权限
sudo chmod +x /usr/local/bin/docker-compose
# 验证安装
docker-compose --version
```

* 部署
```Shell
# 如果 index.html 有修改则需要同步至 nginx
# 把 index.html 移动到 nginx 默认网站根目录
sudo mkdir /var/www/html/api
sudo mkdir /var/www/html/stats
sudo cp ./drone-api/resources/index.html /var/www/html/api
sudo cp ./drone-api/resources/drone-icon.svg /var/www/html/api
sudo cp ./drone-stats-service/resources/index.html /var/www/html/stats/
sudo cp ./drone-stats-service/resources/script.js /var/www/html/stats/
sudo cp ./drone-stats-service/resources/styles.css /var/www/html/stats/
sudo nginx -t
sudo nginx -s reload
sudo systemctl start nginx

# 一键部署并启动服务
./deploy.sh

# 停止服务
./stop.sh
```

* frp 内网穿透
```
如需使用公网 ip 转发，请单独配置 frpc 进程
```
