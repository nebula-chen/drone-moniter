let droneMap; // 全局变量以便于在事件处理程序中访问

window.addEventListener('load', function() {
    if (typeof AMap === 'undefined') {
        console.error('高德地图API加载失败');
        return;
    }

    class DroneMap3D {
        constructor(containerId) {
            // 创建3D地图实例，设置为中山大学深圳校区（光明校区）位置
            this.map = new AMap.Map(containerId, {
                zoom: 15.5,
                center: [113.953099, 22.800721],
                viewMode: '3D',
                pitch: 0,
                rotation: 0,
                mapStyle: 'amap://styles/darkblue',
                features: ['bg', 'building', 'point', 'road'],
                buildingAnimation: true,
                expandZoomRange: true,
                zooms: [3, 20]
            });

            // 添加3D建筑物图层
            const buildings = new AMap.Buildings({
                zIndex: 130,
                merge: false,
                sort: false,
                zooms: [17, 20]
            });
            this.map.add(buildings);

            // 添加地图控件
            this.map.addControl(new AMap.Scale());
            this.map.addControl(new AMap.ToolBar({
                position: 'RB'
            }));
            this.map.addControl(new AMap.MapType({
                defaultType: 0,
                position: 'RB'
            }));

            // 初始化旋转角度
            this.currentRotation = -15;
            
            // 存储所有无人机数据的集合
            this.droneCollection = new Map();
            
            // 存储已收到的recordId集合，用于统计飞行次数
            this.flightCodeSet = new Set();
            
            // 存储飞行路线数据，按recordId分组
            this.flightPaths = new Map();
            
            // 存储飞行路线对象，用于绘制和更新路线
            this.pathPolylines = new Map();

            // 新增-存储飞行路线对象，用于3D可视化
            this.flightPath3DLines = new Map();
            this.customCoords = this.map.customCoords; // 获取地图的自定义坐标转换工具
            
            // 当前选中的无人机ID
            this.selectedDroneId = null;
            
            // 设置轨迹显示状态 - 默认为显示
            this.showPaths = true;

            // 新增：轨迹绘制模式，true为3D轨迹，false为2D轨迹
            this.use3DPath = false;
            
            // 添加按钮事件监听
            this.initButtonControl();

            // 等待地图加载完成
            this.map.on('complete', () => {
                // 新增-初始化3D飞行轨迹图层
                this.initFlightPath3DLayer();

                document.getElementById('loading').style.display = 'none';

                this.showRecentTracksAnimated();

                this.connectWebSocket();
            });
        }

        // 替换键盘控制为按钮控制
        initButtonControl() {
            // 俯视图按钮
            document.getElementById('resetRotation').addEventListener('click', () => {
                this.currentRotation = 0;
                this.map.setRotation(this.currentRotation);
                this.map.setPitch(0);
                this.map.setZoom(15.5);

                // 清除所有轨迹
                this.clearAllPaths();

                // 切换轨迹绘制为2D
                this.use3DPath = false;

            });

            // 轴测图按钮
            document.getElementById('axonometric').addEventListener('click', () => {
                this.currentRotation = 0;
                this.map.setRotation(this.currentRotation);
                this.map.setPitch(45);
                this.map.setZoom(17);

                // 清除所有轨迹
                this.clearAllPaths();

                // 切换轨迹绘制为3D
                this.use3DPath = true;
            });

            // // 高度曲线图按钮
            // document.getElementById('altitudeCurve').addEventListener('click', () => {
            //     const currentZoom = this.map.getZoom();
            //     const newZoom = Math.min(currentZoom + 0.5, 20);  // 最大缩放级别限制在20
            //     this.map.setZoom(newZoom);
            // });
            
            // 添加飞行路径显示/隐藏控制
            // const togglePathsBtn = document.getElementById('togglePaths');
            
            // togglePathsBtn.addEventListener('click', () => {
            //     // 切换显示状态
            //     this.showPaths = !this.showPaths;
                
            //     // 更新按钮文本和样式
            //     if (this.showPaths) {
            //         togglePathsBtn.textContent = '隐藏轨迹';
            //         togglePathsBtn.classList.add('active');
            //     } else {
            //         togglePathsBtn.textContent = '显示轨迹';
            //         togglePathsBtn.classList.remove('active');
            //     }
                
            //     // 更新所有路径的可见性
            //     this.toggleAllPathsVisibility(this.showPaths);
            // });
        }

        simulateDroneData() {
            // 由于现在从后端接收多架无人机的数据，此方法不再需要
            // 这里可以增加一些默认数据，仅在开发环境下使用
            console.log('准备接收无人机数据...');
        }
        
        connectWebSocket() {
            // 创建WebSocket连接
            this.socket = new WebSocket('ws://172.25.74.79:19999/api/ws');
            
            // 更新连接状态为"连接中"
            const statusEl = document.getElementById('connection-status');
            statusEl.className = 'connection-status connecting';
            statusEl.textContent = '连接状态: 连接中...';
            
            // 连接建立时的处理
            this.socket.onopen = () => {
                console.log('WebSocket连接已建立');
                statusEl.className = 'connection-status connected';
                statusEl.textContent = '连接状态: 已连接';
                
                // 设置心跳检测，每30秒发送一次心跳
                this.heartbeatInterval = setInterval(() => {
                    if (this.socket.readyState === WebSocket.OPEN) {
                        this.socket.send(JSON.stringify({ type: 'heartbeat' }));
                    }
                }, 30000);
            };
            
            // 接收到消息时的处理
            this.socket.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    // 检查是否是心跳响应
                    if (data.type === 'heartbeat') {
                        console.log('收到心跳响应');
                        return;
                    }
                    this.processDroneData(data);
                } catch (error) {
                    console.error('解析WebSocket数据错误:', error);
                }
            };
            
            // 连接关闭时的处理
            this.socket.onclose = () => {
                console.log('WebSocket连接已关闭，尝试重新连接...');
                statusEl.className = 'connection-status disconnected';
                statusEl.textContent = '连接状态: 已断开，正在重连...';
                
                // 清除心跳定时器
                if (this.heartbeatInterval) {
                    clearInterval(this.heartbeatInterval);
                }
                
                // 2秒后尝试重新连接
                setTimeout(() => this.connectWebSocket(), 2000);
            };
            
            // 连接错误时的处理
            this.socket.onerror = (error) => {
                console.error('WebSocket连接错误:', error);
                statusEl.className = 'connection-status disconnected';
                statusEl.textContent = '连接状态: 连接错误';
            };
        }
        
        // 添加点击事件逻辑
        on3DClick(event) {
            // 根据渲染canvas的坐标计算鼠标的标准化设备坐标
            const rect = this.fp_renderer.domElement.getBoundingClientRect();
            const mouse = new THREE.Vector2();
            mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
            mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;

            const raycaster = new THREE.Raycaster();
            raycaster.setFromCamera(mouse, this.fp_camera);
            // 获取所有3D无人机标记
            const markers = this.droneMarkerGroup ? this.droneMarkerGroup.children : [];
            const intersects = raycaster.intersectObjects(markers);
            if (intersects.length > 0) {
                // 找到最近的被点击对象
                const sprite = intersects[0].object;
                const droneId = sprite.userData.droneId;
                this.selectedDroneId = droneId;
                if (this.droneCollection.has(droneId)) {
                    const droneData = this.droneCollection.get(droneId);
                    this.updateInfoPanel(droneData);
                }
                return true; // 命中无人机标记
            }
            return false; // 未命中
        }
        
        // 新增：初始化3D飞行轨迹图层
        initFlightPath3DLayer() {
            const self = this;
            // 创建一个用于存放3D轨迹线的组
            this.flightPathGroup = new THREE.Group();
            // 创建GLCustomLayer
            this.flightPathLayer = new AMap.GLCustomLayer({
                zIndex: 5,
                init: (gl) => {
                    self.fp_renderer = new THREE.WebGLRenderer({
                        context: gl,
                        alpha: true, // 开启透明背景
                    });
                    self.fp_renderer.autoClear = false;
                    self.fp_renderer.setClearColor(0x000000, 0);
                    self.fp_scene = new THREE.Scene();
                    self.fp_scene.add(self.flightPathGroup);    // 添加3D轨迹组到场景中
                    self.droneMarkerGroup = new THREE.Group();  // 创建无人机标记组
                    self.fp_scene.add(self.droneMarkerGroup);   // 添加无人机标记组到场景中
                    self.fp_camera = new THREE.PerspectiveCamera(60, window.innerWidth / window.innerHeight, 0.1, 1 << 28);

                    // 让相机正上方俯视地面
                    self.fp_camera.position.set(0, 0, 2000); // z 值大一些，视野更高
                    self.fp_camera.up.set(0, 1, 0);
                    self.fp_camera.lookAt(0, 0, 0);
                    // 修改：只在命中无人机标记时阻止事件冒泡
                    self.fp_renderer.domElement.addEventListener('pointerdown', (event) => {
                        const hit = this.on3DClick(event);
                        if (hit) {
                            event.stopPropagation();
                            event.preventDefault();
                        }
                        // 没命中时不阻止，事件会传递到高德地图
                    });
                },
                render: () => {
                    self.fp_renderer.resetState();
                    // 同步相机参数（使用地图的customCoords工具获取当前相机参数）
                    this.customCoords.setCenter([113.953099, 22.800721]);
                    const { near, far, fov, up, lookAt, position } = this.customCoords.getCameraParams();
                    self.fp_camera.near = near;
                    self.fp_camera.far = far;
                    self.fp_camera.fov = fov;
                    self.fp_camera.position.set(...position);
                    self.fp_camera.up.set(...up);
                    self.fp_camera.lookAt(...lookAt);
                    self.fp_camera.updateProjectionMatrix();
                    
                    self.fp_renderer.render(self.fp_scene, self.fp_camera);
                    self.fp_renderer.resetState();
                }
            });
            this.map.add(this.flightPathLayer);
        }

        processDroneData(data) {
            // 根据API文档中的数据结构进行处理
            console.log('收到无人机数据:', data);
            
            // 检查数据是否有效
            if (!data || typeof data !== 'object') {
                console.error('接收到无效的无人机数据');
                return;
            }
            
            // 注意：后端的经纬度是整数，需要转换为浮点数并除以适当的因子
            const droneData = {
                id: data.flightCode,
                recordId: data.flightCode,                  // 无人机ID作为唯一标识
                longitude: data.longitude / 10000000.0, // 将整数转换为度数
                latitude: data.latitude / 10000000.0,   // 将经度转换为度数
                altitude: data.altitude / 10,         // 海拔高度
                height: data.height / 10,             // 对地高度
                heading: data.course,                // 无人机朝向角度
                speed: data.VS,                      // 计算速度
                uavType: data.uavType,               // 无人机类型
                orderID: data.orderID,               // 记录ID
                timeStamp: data.timeStamp,           // 时间戳
                SOC: data.SOC                        // 初始化模拟电池电量
            };
            
            // 检查是否是新的flightCode
            if (data.flightCode && !this.flightCodeSet.has(data.flightCode)) {
                // 清除之前所有的飞行路径，只保留当前正在执行的
                this.clearAllPaths();
                
                // 添加新的flightCode到集合
                this.flightCodeSet.add(data.flightCode);
                
                // 初始化新的飞行路径数组
                this.flightPaths.set(data.flightCode, []);
                
                // 更新统计数据
                this.updateStatsPanel();
            }
            
            // 添加当前位置到飞行路径
            if (data.flightCode) {
                const path = this.flightPaths.get(data.flightCode) || [];
                const newPoint = [droneData.longitude, droneData.latitude, droneData.altitude];
                
                // 检查是否有效点位，避免添加重复点或异常点
                const lastPoint = path.length > 0 ? path[path.length - 1] : null;
                if (!lastPoint || 
                    (lastPoint[0] !== newPoint[0] || 
                    lastPoint[1] !== newPoint[1] ||
                    lastPoint[2] !== newPoint[2])) {
                    path.push(newPoint);
                    this.flightPaths.set(data.flightCode, path);
                    
                    const color = this.getColorByRecordId(recordId);
                    this.updateFlightPath(data.flightCode, path, color);   // 更新2D路径
                    // this.updateFlightPath3D(data.flightCode, path);   // 更新3D路径
                }
            }
            
            // 将无人机数据存入集合
            this.droneCollection.set(droneData.id, droneData);
            
            this.updateOrCreateDroneMarker(droneData);   // 更新2D无人机标记
            // this.updateOrCreateDroneMarker3D(droneData);    // 更新3D无人机标记
            
            // 如果是当前选中的无人机，更新信息面板
            if (this.selectedDroneId === droneData.id) {
                this.updateInfoPanel(droneData);
            }
            
            // 更新统计数据
            this.updateStatsPanel();
        }

        // 修改-将颜色生成提取为类方法
        getColorByRecordId = (id) => {
            // 简单的哈希算法生成颜色
            let hash = 0;
            for (let i = 0; i < id.length; i++) {
                hash = id.charCodeAt(i) + ((hash << 5) - hash);
            }
            const hue = Math.abs(hash % 360);
            return `hsl(${hue}, 100%, 50%)`;
        };
        
        // 更新2D飞行路径
        updateFlightPath(recordId, path, color) {
            if (this.use3DPath) {
                this.updateFlightPath3D(recordId, path);
                return;
            }
            // 原2D绘制逻辑
            if (this.pathPolylines.has(recordId)) {
                // 更新已有路径线
                const polyline = this.pathPolylines.get(recordId);
                polyline.setPath(path);
            } else {
                // 创建新的路径线
                
                const polyline = new AMap.Polyline({
                    path: path,
                    strokeColor: color,
                    strokeWeight: 4,
                    strokeOpacity: 0.8,
                    zIndex: 100,
                    strokeStyle: 'solid',
                    strokeDasharray: [10, 5],
                    visible: this.showPaths // 根据当前设置决定是否可见
                });
                
                this.map.add(polyline);
                
                // 如果当前是隐藏轨迹模式，则立即隐藏这条新轨迹
                if (!this.showPaths) {
                    polyline.hide();
                }
                
                this.pathPolylines.set(recordId, polyline);
            }
        }
        
        // 新增：利用THREE更新3D轨迹（与当前recordId对应）
        updateFlightPath3D(recordId, path) {
            // 使用customCoords工具将每个[lng, lat]点转换为3D坐标，并使用altitude值进行Z轴偏移
            const points = path.map(pt => {
                const [lng, lat, altitude] = pt;
                // customCoords.lngLatsToCoords 接受二维数组
                const coord = this.customCoords.lngLatsToCoords([[lng, lat]])[0];
                // 将转换后的坐标第三个元素赋予高度（若需要，可以添加放大因子）
                coord[2] = altitude;
                return new THREE.Vector3(coord[0], coord[1], coord[2]);
            });
            const geometry = new THREE.BufferGeometry().setFromPoints(points);
            const material = new THREE.LineBasicMaterial({ 
                color: new THREE.Color(this.getColorByRecordId(recordId)),  // 使用recordId生成颜色
            });
            if (this.flightPath3DLines.has(recordId)) {
                // 更新已有3D轨迹线的几何数据
                const line = this.flightPath3DLines.get(recordId);
                line.geometry.dispose();    // 释放旧的几何体资源
                line.geometry = geometry;   // 更新几何体

                this.flightPathLayer.show(); // 间接触发3D渲染
            } else {
                // 新建轨迹线对象
                const line = new THREE.Line(geometry, material);
                this.flightPathGroup.add(line);
                this.flightPath3DLines.set(recordId, line);
            }
        }
        
        // 新增：获取最近3条飞行记录并动画绘制轨迹
        async showRecentTracksAnimated() {
            try {
                const res = await fetch('/record/recentTracks?n=3');
                const data = await res.json();
                if (!data.Track || data.Track.length === 0) return;
                // 按 flightCode 分组
                const tracksByRecord = {};
                data.Track.forEach(pt => {
                    const recordId = pt.flightCode;
                    if (!tracksByRecord[recordId]) tracksByRecord[recordId] = [];
                    // 保留原始点及flightStatus
                    tracksByRecord[recordId].push({
                        lng: pt.longitude / 1e7,
                        lat: pt.latitude / 1e7,
                        alt: pt.altitude,
                        flightStatus: pt.flightStatus
                    });
                });
                // 对每条轨迹依次动画绘制
                const recordIds = Object.keys(tracksByRecord);
                let idx = 0;
                const drawNextTrack = () => {
                    if (idx >= recordIds.length) {
                        // 所有轨迹绘制完毕，延时清除并重新开始
                        setTimeout(() => {
                            this.clearAllPaths();
                            setTimeout(() => {
                                this.showRecentTracksAnimated();
                            }, 500); // 清除后稍作延迟再重播
                        }, 1000); // 所有轨迹展示完后停留1.5秒
                        return;
                    }
                    const points = tracksByRecord[recordIds[idx]];
                    const color = this.getColorByRecordId(recordIds[idx]);
                    let path = [];
                    let i = 0;
                    const animate = () => {
                        if (i < points.length) {
                            // 如果当前点为TakeOff，重置path，避免首尾连线
                            if (points[i].flightStatus === 'TakeOff') {
                                path = [];
                            }
                            path.push([points[i].lng, points[i].lat, points[i].alt]);
                            this.updateFlightPath(recordIds[idx], path, color);
                            i++;
                            setTimeout(animate, 1000);
                        } else {
                            idx++;
                            setTimeout(drawNextTrack, 2000);
                        }
                    };
                    animate();
                };
                drawNextTrack();
            } catch (e) {
                console.error('获取最近轨迹失败', e);
            }
        }
        
        // 切换所有路径的可见性
        toggleAllPathsVisibility(visible) {
            if (this.pathPolylines) {
                this.pathPolylines.forEach(polyline => {
                    if (visible) {
                        polyline.show();
                    } else {
                        polyline.hide();
                    }
                });
            }
        }
        
        // 清除所有飞行路径
        clearAllPaths() {
            if (this.pathPolylines) {
                this.pathPolylines.forEach(polyline => {
                    this.map.remove(polyline);
                });
                this.pathPolylines.clear();
            }
            this.flightPaths.clear();

            // 新增-清除3D飞行路径
            if (this.flightPathGroup) {
                while (this.flightPathGroup.children.length) {
                    this.flightPathGroup.remove(this.flightPathGroup.children[0]);
                }
                this.flightPath3DLines.clear();
            }
        }
        
        // 更新或创建无人机标记
        updateOrCreateDroneMarker(data) {
            if (!this.droneMarkers) {
                this.droneMarkers = new Map();
            }
            
            const { id, longitude, latitude, heading } = data;
            
            if (this.droneMarkers.has(id)) {
                // 更新已有标记位置
                const marker = this.droneMarkers.get(id);
                marker.setPosition([longitude, latitude]);
                marker.setAngle(heading);
            } else {
                // 创建新的无人机标记
                const droneIcon = new AMap.Icon({
                    size: new AMap.Size(32, 32),
                    image: './drone-icon.svg',
                    imageSize: new AMap.Size(32, 32)
                });
                
                const marker = new AMap.Marker({
                    position: [longitude, latitude],
                    icon: droneIcon,
                    offset: new AMap.Pixel(-16, -16),
                    autoRotation: true,
                    angle: heading,
                    anchor: 'center',
                    zIndex: 150,
                    extData: { droneId: id }
                });
                
                // 添加点击事件，显示对应无人机信息
                marker.on('click', (e) => {
                    const droneId = e.target.getExtData().droneId;
                    this.selectedDroneId = droneId;
                    if (this.droneCollection.has(droneId)) {
                        const droneData = this.droneCollection.get(droneId);
                        this.updateInfoPanel(droneData);
                    }
                });
                
                this.map.add(marker);
                this.droneMarkers.set(id, marker);
            }
        }

        // 修改：更新或创建无人机标记（使用Three.js在3D层展示图标）
        updateOrCreateDroneMarker3D(data) {
            if (!this.droneMarkers3D) {
                this.droneMarkers3D = new Map();
            }
            
            const { id, longitude, latitude, heading, altitude } = data;
            
            // 利用customCoords将经纬度转换为3D坐标
            const coord = this.map.customCoords.lngLatsToCoords([[longitude, latitude]])[0];
            // 使用数据中的高度，并加上5单位的偏移，使标记稍微悬浮
            coord[2] = altitude;
            
            if (this.droneMarkers3D.has(id)) {
                // 更新已有标记
                const marker = this.droneMarkers3D.get(id);
                marker.position.set(coord[0], coord[1], coord[2]);
                marker.material.rotation = heading * Math.PI / 180;
            } else {
                // 创建新的无人机标记，使用THREE.Sprite展示图标
                const texture = new THREE.TextureLoader().load('./drone-icon.svg');
                const material = new THREE.SpriteMaterial({
                    map: texture,
                    transparent: true
                });
                const sprite = new THREE.Sprite(material);
                // 根据实际图标尺寸调整scale（此处单位与3D场景相关）
                sprite.scale.set(32, 32, 1);
                sprite.position.set(coord[0], coord[1], coord[2]);
                sprite.material.rotation = heading * Math.PI / 180;
                
                // 可在此处添加鼠标交互事件，但需要借助射线检测
                sprite.userData.droneId = id;
                
                // 将3D标记添加到无人机标记组中
                if (this.droneMarkerGroup) {
                    this.droneMarkerGroup.add(sprite);
                }
                this.droneMarkers3D.set(id, sprite);
            }
        }
        
        updateStatsPanel() {
            // 更新统计面板 - 显示无人机数量和飞行次数
            document.getElementById('drone-count').textContent = this.droneCollection.size;
            document.getElementById('flight-count').textContent = this.flightCodeSet.size;
        }
        
        updateInfoPanel(data) {
            // 更新信息面板 - 显示选中无人机的详细信息
            const infoPanel = document.getElementById('info-panel');
            const selectionHint = document.getElementById('selection-hint');
            const droneInfo = document.getElementById('drone-info');
            
            // 显示无人机信息，隐藏选择提示
            infoPanel.classList.remove('no-selection');
            selectionHint.style.display = 'none';
            droneInfo.style.display = 'block';
            
            // 更新信息面板内容
            document.getElementById('speed-info').textContent = data.speed.toFixed(2);
            document.getElementById('altitude-info').textContent = data.altitude.toFixed(2);
            document.getElementById('height-info').textContent = data.height.toFixed(2);
            document.getElementById('longitude').textContent = data.longitude.toFixed(6);
            document.getElementById('latitude').textContent = data.latitude.toFixed(6);
            document.getElementById('uav-id').textContent = data.id || '-';
            document.getElementById('soc').textContent = data.SOC + '%' || '0%';
            
            // 显示更多无人机相关信息
            // 检查是否已经添加了额外信息项
            // if (!document.getElementById('uav-type-info')) {
            //     // 添加无人机类型
            //     const typeInfo = document.createElement('p');
            //     typeInfo.innerHTML = `无人机类型: <span id="uav-type-info">-</span>`;
            //     droneInfo.appendChild(typeInfo);
                
            //     // 添加飞行时间
            //     const flightTimeInfo = document.createElement('p');
            //     flightTimeInfo.innerHTML = `飞行时间: <span id="flight-time-info">00:00:00</span>`;
            //     droneInfo.appendChild(flightTimeInfo);
                
            //     // 添加总飞行距离
            //     const distanceInfo = document.createElement('p');
            //     distanceInfo.innerHTML = `飞行距离: <span id="distance-info">0 m</span>`;
            //     droneInfo.appendChild(distanceInfo);
                
            //     // 添加电池电量
            //     const batteryInfo = document.createElement('p');
            //     batteryInfo.innerHTML = `电池电量: <span id="battery-info">100%</span>`;
            //     droneInfo.appendChild(batteryInfo);
            // }
            
            // 更新额外信息
            // document.getElementById('uav-type-info').textContent = data.uavType || '-';
            
            // // 格式化飞行时间
            // const hours = Math.floor(data.flightTime / 3600);
            // const minutes = Math.floor((data.flightTime % 3600) / 60);
            // const seconds = data.flightTime % 60;
            // const formattedTime = 
            //     String(hours).padStart(2, '0') + ':' + 
            //     String(minutes).padStart(2, '0') + ':' + 
            //     String(seconds).padStart(2, '0');
            // document.getElementById('flight-time-info').textContent = formattedTime;
            
            // // 格式化飞行距离
            // let distanceText = '';
            // if (data.totalDistance >= 1000) {
            //     distanceText = (data.totalDistance / 1000).toFixed(2) + ' km';
            // } else {
            //     distanceText = Math.round(data.totalDistance) + ' m';
            // }
            // document.getElementById('distance-info').textContent = distanceText;
            
            // // 更新电池电量
            // document.getElementById('battery-info').textContent = data.battery + '%';
        }
        
        // 清理方法，用于关闭WebSocket连接和清理资源
        cleanup() {
            if (this.socket) {
                this.socket.close();
            }
            
            if (this.heartbeatInterval) {
                clearInterval(this.heartbeatInterval);
            }
            
            // 清理所有无人机标记
            if (this.droneMarkers) {
                this.droneMarkers.forEach(marker => {
                    this.map.remove(marker);
                });
                this.droneMarkers.clear();
            }
            
            // 清理所有飞行路径
            if (this.pathPolylines) {
                this.pathPolylines.forEach(polyline => {
                    this.map.remove(polyline);
                });
                this.pathPolylines.clear();
            }
        }
    }

    // 创建地图实例
    droneMap = new DroneMap3D('map-scene');
    
    // 添加页面卸载前的清理
    window.addEventListener('beforeunload', () => {
        if (droneMap) {
            droneMap.cleanup();
        }
    });
});
