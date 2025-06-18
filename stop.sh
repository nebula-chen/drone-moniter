# 停止 drone-api 服务
cd ./drone-api
path=$(pwd -P)
echo "进入目录 ${path}"
docker-compose down
echo "离开目录 ${path}"

# 停止 drone-stats-service 服务
cd ../drone-stats-service
path=$(pwd -P)
echo "进入目录 ${path}"
docker-compose down
echo "离开目录 ${path}"

# 停止 influxDB 服务
cd ../influxDB
path=$(pwd -P)
echo "进入目录 ${path}"
docker-compose down
echo "离开目录 ${path}"

if docker network inspect my-network &>/dev/null; then
  docker network rm my-network
  echo "网络 my-network 已删除"
fi

echo "ALL Service Stop!!!"