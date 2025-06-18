
if ! docker network inspect my-network &>/dev/null; then
  docker network create --driver bridge my-network
  echo "网络 my-network 创建成功"
else
  echo "网络 my-network 已存在，跳过创建"
fi

cd ./influxDB
mkdir -p ./influxdb2

path=$(pwd -P)
echo "进入目录 ${path}"
if ! docker-compose up --build -d; then
  echo "错误: docker-compose 启动 influxdb 失败!" >&2
  exit 1
fi
echo "离开目录 ${path}"

cd ../drone-api
path=$(pwd -P)
echo "进入目录 ${path}"
if ! docker-compose up --build -d; then
  echo "错误: docker-compose 启动 drone-api 失败!" >&2
  exit 1
fi
echo "离开目录 ${path}"

cd ../drone-stats-service
path=$(pwd -P)
echo "进入目录 ${path}"
if ! docker-compose up --build -d; then
  echo "错误: docker-compose 启动 drone-stats-service 失败!" >&2
  exit 1
fi
echo "离开目录 ${path}"

echo "Done!!!"
