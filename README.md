# drone-moniter
无人机监控系统后端接口

* 安装 nginx
```Shell
# 包管理器安装
sudo apt-get install nginx

配置文件目录：/etc/nginx/
默认网站根目录：/var/www/html/
日志文件：/var/log/nginx/
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

```Shell

```
