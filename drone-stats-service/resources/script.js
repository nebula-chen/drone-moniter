// 页面加载完成后执行
document.addEventListener('DOMContentLoaded', function() {
    // 初始化数据
    loadFlightRecords();
    // 不再默认加载飞行轨迹点数据
    // loadFlightPoints();

    // 添加导航切换事件
    document.getElementById('monitor-link').addEventListener('click', function(e) {
        e.preventDefault();
        showContent('monitor');
    });

    // 数据分析功能注释开始
    /*
    document.getElementById('analysis-link').addEventListener('click', function(e) {
        e.preventDefault();
        showContent('analysis');
    });
    */
    // 数据分析功能注释结束

    // 添加查询按钮事件
    document.getElementById('search-btn').addEventListener('click', function() {
        const uavId = document.querySelector('.input-group input').value.trim(); 
        const startTime = document.getElementById('start-time').value;
        const endTime = document.getElementById('end-time').value;
        // 如果输入框为空，则查询全部
        if (!uavId) uavId = ""; // 查询全部
        if (!startTime) startTime = "";
        if (!endTime) endTime = "";
        
        searchFlightRecords(uavId, startTime, endTime);
        
        // 隐藏飞行轨迹点数据区域
        hideFlightPoints();
    });
    
    // 添加重置按钮事件
    document.getElementById('reset-btn').addEventListener('click', function() {
        // 清空输入框
        document.querySelector('.input-group input').value = "";
        document.getElementById('start-time').value = "";
        document.getElementById('end-time').value = "";
        
        // 重新加载所有数据
        loadFlightRecords();
        
        // 隐藏飞行轨迹点数据区域
        hideFlightPoints();
    });
    
    // 添加关闭详情按钮事件
    document.getElementById('close-details-btn').addEventListener('click', function() {
        hideFlightPoints();
    });
    
    // 设置日期选择器默认值为当前日期
    const today = new Date();
    const oneWeekAgo = new Date(today);
    oneWeekAgo.setDate(today.getDate() - 7);
    
    function formatDateTimeLocal(date) {
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        const hour = String(date.getHours()).padStart(2, '0');
        const minute = String(date.getMinutes()).padStart(2, '0');
        return `${year}-${month}-${day}T${hour}:${minute}`;
    }

    document.getElementById('start-time').value = formatDateTimeLocal(oneWeekAgo);
    document.getElementById('end-time').value = formatDateTimeLocal(today);
});

// 隐藏飞行轨迹点数据区域
function hideFlightPoints() {
    document.getElementById('flight-points-container').style.display = 'none';
}

// 显示飞行轨迹点数据区域
function showFlightPoints() {
    document.getElementById('flight-points-container').style.display = 'block';
}

// 格式化日期为YYYY-MM-DD格式
function formatDate(date) {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
}

// 显示指定内容区域
function showContent(contentType) {
    // 更新导航链接状态
    document.querySelectorAll('.nav-link').forEach(link => {
        link.classList.remove('active');
    });
    document.getElementById(contentType + '-link').classList.add('active');

    // 更新内容区域显示
    document.querySelectorAll('.content-section').forEach(section => {
        section.style.display = 'none';
    });
    document.getElementById(contentType + '-content').style.display = 'block';
}

// 加载飞行记录数据
function loadFlightRecords() {
    fetch('/api/flight/query', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            flightCode: "", // 查询全部
            startTime: "",
            endTime: ""
        })
    })
    .then(res => res.json())
    .then(data => {
        renderFlightRecords(data.flightRecords || []);
    })
    .catch(() => {
        renderFlightRecords([]);
    });
}

// 根据条件搜索飞行记录
function searchFlightRecords(uavId, startTime, endTime) {
    function toBackendFormat(dt) {
        if (!dt) return "";
        return dt.replace('T', ' ') + ':00';
    }
    fetch('/api/flight/query', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            flightCode: uavId,
            startTime: toBackendFormat(startTime),
            endTime: toBackendFormat(endTime)
        })
    })
    .then(res => res.json())
    .then(data => {
        renderFlightRecords(data.flightRecords || []);
    })
    .catch(() => {
        renderFlightRecords([]);
    });
}

// 渲染飞行记录数据
function renderFlightRecords(data) {
    const tbody = document.getElementById('flight-records');
    tbody.innerHTML = '';

    if (data.length === 0) {
        const row = document.createElement('tr');
        row.innerHTML = `<td colspan="12" class="text-center">没有找到符合条件的记录</td>`;
        tbody.appendChild(row);
        return;
    }

    data.forEach(record => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${record.id}</td>
            <td>${record.uav_id}</td>
            <td>${record.start_time}</td>
            <td>${record.end_time}</td>
            <td>${record.start_lat}</td>
            <td>${record.start_lng}</td>
            <td>${record.end_lat}</td>
            <td>${record.end_lng}</td>
            <td>${record.distance}</td>
            <td>${record.battery_used}</td>
            <td>${record.created_at}</td>
            <td>
                <button class="btn btn-sm btn-outline-primary" onclick="viewFlightDetails(${record.id}, '${record.uav_id}')">查看详情</button>
            </td>
        `;
        tbody.appendChild(row);
    });
}

// 加载飞行轨迹点数据
function loadFlightPoints(recordId) {
    // 模拟从API获取数据
    const flightPointsData = [
        {
            id: 5326,
            flight_record_id: recordId,
            flight_status: 'Inflight',
            time_stamp: '2025-06-16 00:47:33',
            longitude: 113.9530990,
            latitude: 22.8007210,
            altitude: 0.0,
            soc: 100
        },
        {
            id: 5327,
            flight_record_id: recordId,
            flight_status: 'Inflight',
            time_stamp: '2025-06-16 00:47:34',
            longitude: 113.9530990,
            latitude: 22.8007210,
            altitude: 2.0,
            soc: 100
        },
        {
            id: 5328,
            flight_record_id: recordId,
            flight_status: 'Inflight',
            time_stamp: '2025-06-16 00:47:35',
            longitude: 113.9530990,
            latitude: 22.8007210,
            altitude: 4.0,
            soc: 100
        },
        {
            id: 5329,
            flight_record_id: recordId,
            flight_status: 'Inflight',
            time_stamp: '2025-06-16 00:47:36',
            longitude: 113.9530990,
            latitude: 22.8007210,
            altitude: 5.0,
            soc: 100
        },
        {
            id: 5330,
            flight_record_id: recordId,
            flight_status: 'Inflight',
            time_stamp: '2025-06-16 00:47:37',
            longitude: 113.9530990,
            latitude: 22.8007210,
            altitude: 25.0,
            soc: 100
        }
    ];

    renderFlightPoints(flightPointsData);
}

// 渲染飞行轨迹点数据
function renderFlightPoints(data) {
    const tbody = document.getElementById('flight-points');
    tbody.innerHTML = '';

    if (data.length === 0) {
        const row = document.createElement('tr');
        row.innerHTML = `<td colspan="8" class="text-center">没有找到轨迹点数据</td>`;
        tbody.appendChild(row);
        return;
    }

    data.forEach(point => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${point.id}</td>
            <td>${point.flight_record_id}</td>
            <td>${point.flight_status}</td>
            <td>${point.time_stamp}</td>
            <td>${point.longitude}</td>
            <td>${point.latitude}</td>
            <td>${point.altitude}</td>
            <td>${point.soc}</td>
        `;
        tbody.appendChild(row);
    });
}

// 根据无人机ID筛选记录
function filterRecordsByUavId(uavId) {
    // 模拟从API获取筛选后的数据
    const flightRecordsData = [
        {
            id: 11,
            uav_id: 'uav1',
            start_time: '2025-06-16 08:47:33',
            end_time: '2025-06-16 09:27:50',
            start_lat: 22.0000000,
            start_lng: 113.0000000,
            end_lat: 22.0000000,
            end_lng: 113.0000000,
            distance: 9469.83,
            battery_used: 59,
            created_at: '2025-06-17 17:00:58'
        }
    ];

    renderFlightRecords(flightRecordsData);
}

// 查看飞行详情
function viewFlightDetails(recordId, uavId) {
    // 设置选中的飞行记录ID显示
    document.getElementById('selected-flight-id').textContent = `ID: ${recordId} (${uavId})`;
    
    // 显示飞行轨迹点数据区域
    showFlightPoints();
    
    // 根据记录ID加载对应的轨迹点
    loadFlightPoints(recordId);
    
    // 滚动到轨迹点数据区域
    document.getElementById('flight-points-container').scrollIntoView({ behavior: 'smooth' });
}

// 添加导出按钮事件
document.querySelector('.btn-outline-primary').addEventListener('click', function() {
    const uavId = document.querySelector('.input-group input').value.trim();
    const startTime = document.getElementById('start-time').value;
    const endTime = document.getElementById('end-time').value;
    let url = `/api/flight/export?uavId=${encodeURIComponent(uavId)}&startTime=${encodeURIComponent(startTime)}&endTime=${encodeURIComponent(endTime)}`;
    window.open(url, '_blank');
});