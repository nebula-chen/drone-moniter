package export

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"drone-stats-service/internal/dao"
)

// 任务状态
const (
	TaskStatusPending = "pending"
	TaskStatusRunning = "running"
	TaskStatusDone    = "done"
	TaskStatusFailed  = "failed"
)

// Task 描述一个导出任务的元信息（持久化到磁盘）
type Task struct {
	ID         string    `json:"id"`
	Target     string    `json:"target"` // records | trajectory | both
	OrderID    string    `json:"orderId"`
	UasID      string    `json:"uasId"`
	StartTime  string    `json:"startTime"`
	EndTime    string    `json:"endTime"`
	Format     string    `json:"format"` // xlsx | csv
	Status     string    `json:"status"`
	ResultFile string    `json:"resultFile"` // 本地文件路径
	Error      string    `json:"error"`
	CreatedAt  time.Time `json:"createdAt"`
	FinishedAt time.Time `json:"finishedAt"`
}

// TaskManager 管理导出任务队列并执行
type TaskManager struct {
	mysql   *dao.MySQLDao
	influx  *dao.InfluxDao
	tasks   map[string]*Task
	mu      sync.RWMutex
	queue   chan string
	dir     string // 存储任务元数据及输出的根目录
	baseURL string // 用于生成下载 URL 的基路径（可为空，返回相对路径）
}

// NewTaskManager 创建 TaskManager，并启动后台 worker
// dir 为任务工作目录（如果为空，使用系统临时目录下 drone_export_tasks）
func NewTaskManager(mysql *dao.MySQLDao, influx *dao.InfluxDao, dir, baseURL string) (*TaskManager, error) {
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "drone_export_tasks")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	tm := &TaskManager{
		mysql:   mysql,
		influx:  influx,
		tasks:   make(map[string]*Task),
		queue:   make(chan string, 100),
		dir:     dir,
		baseURL: baseURL,
	}
	// load existing tasks metadata if any
	tm.loadTasksFromDisk()
	// 启动 worker
	go tm.worker()
	// 启动周期性清理（默认保留7天，每24小时执行一次）
	go tm.periodicCleaner(24*time.Hour, 7)
	return tm, nil
}

// CreateTask 新建并入队一个导出任务，返回 task id
func (tm *TaskManager) CreateTask(target, orderID, uasID, startTime, endTime, format string) (string, error) {
	if format == "" {
		format = "xlsx"
	}
	// 构造符合要求的 task id: 类型字母 + 时间戳(YYYYMMDDhhmmss) + 格式字母
	// 类型: records->R, trajectory->T, both->B
	// 格式: .xlsx->X, .csv->C
	var prefix string
	switch target {
	case "records":
		prefix = "R"
	case "trajectory":
		prefix = "T"
	case "both":
		prefix = "B"
	default:
		prefix = "U" // unknown
	}
	var fchar string
	if strings.ToLower(format) == "csv" {
		fchar = "C"
	} else {
		fchar = "X"
	}
	ts := time.Now().Format("20060102150405")
	id := prefix + ts + fchar
	// 如果碰巧冲突（同一秒创建了同样类型和格式的任务），追加序号保证唯一
	tm.mu.RLock()
	_, exists := tm.tasks[id]
	tm.mu.RUnlock()
	if exists {
		// 简单尝试加后缀 -1,-2 ...
		for i := 1; ; i++ {
			try := fmt.Sprintf("%s-%d", id, i)
			tm.mu.RLock()
			_, ex := tm.tasks[try]
			tm.mu.RUnlock()
			if !ex {
				id = try
				break
			}
		}
	}
	task := &Task{
		ID:        id,
		Target:    target,
		OrderID:   orderID,
		UasID:     uasID,
		StartTime: startTime,
		EndTime:   endTime,
		Format:    format,
		Status:    TaskStatusPending,
		CreatedAt: time.Now(),
	}
	tm.mu.Lock()
	tm.tasks[id] = task
	tm.mu.Unlock()
	if err := tm.saveTaskToDisk(task); err != nil {
		return "", err
	}
	// enqueue
	tm.queue <- id
	return id, nil
}

// GetTask 返回任务元信息
func (tm *TaskManager) GetTask(id string) (*Task, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	t, ok := tm.tasks[id]
	return t, ok
}

// worker 循环处理任务
func (tm *TaskManager) worker() {
	for id := range tm.queue {
		tm.runTask(id)
	}
}

// runTask 执行导出并更新任务状态
func (tm *TaskManager) runTask(id string) {
	tm.mu.Lock()
	task, ok := tm.tasks[id]
	if !ok {
		tm.mu.Unlock()
		return
	}
	task.Status = TaskStatusRunning
	tm.saveTaskToDisk(task)
	tm.mu.Unlock()

	// 生成任务目录
	taskDir := filepath.Join(tm.dir, id)
	_ = os.MkdirAll(taskDir, 0o755)
	recExt := "xlsx"
	trajExt := "xlsx"
	if task.Format == "csv" {
		recExt = "csv"
		trajExt = "csv"
	}
	recordFile := filepath.Join(taskDir, "flightRecord."+recExt)
	trajFile := filepath.Join(taskDir, "flightTrajectory."+trajExt)

	// parse times
	var err error
	st, _ := time.Parse(time.RFC3339, task.StartTime)
	ed, _ := time.Parse(time.RFC3339, task.EndTime)

	// helper local functions
	exportRecords := func() error {
		if tm.mysql == nil {
			return fmt.Errorf("mysql not configured")
		}
		// 使用流式导出以减少内存占用
		if task.Format == "csv" {
			return tm.mysql.ExportFlightRecordsToCSVStream(task.OrderID, task.UasID, st.Format("2006-01-02 15:04:05"), ed.Format("2006-01-02 15:04:05"), recordFile)
		}
		return tm.mysql.ExportFlightRecordsToExcelStream(task.OrderID, task.UasID, st.Format("2006-01-02 15:04:05"), ed.Format("2006-01-02 15:04:05"), recordFile)
	}

	exportTrajectory := func() error {
		if tm.mysql == nil {
			return fmt.Errorf("mysql not configured")
		}
		// 使用流式导出轨迹点
		if task.Format == "csv" {
			return tm.mysql.ExportTrackPointsToCSVStream(task.StartTime, task.EndTime, task.OrderID, trajFile)
		}
		return tm.mysql.ExportTrackPointsToExcelStream(task.StartTime, task.EndTime, task.OrderID, trajFile)
	}

	switch task.Target {
	case "records":
		err = exportRecords()
		if err == nil {
			task.ResultFile = recordFile
		}
	case "trajectory":
		err = exportTrajectory()
		if err == nil {
			task.ResultFile = trajFile
		}
	case "both":
		// generate both and zip
		if err = exportRecords(); err == nil {
			if err = exportTrajectory(); err == nil {
				zipPath := filepath.Join(taskDir, "flight_export.zip")
				if e := CreateZip([]string{recordFile, trajFile}, zipPath); e != nil {
					err = e
				} else {
					task.ResultFile = zipPath
				}
			}
		}
	default:
		err = fmt.Errorf("unknown target %s", task.Target)
	}

	tm.mu.Lock()
	if err != nil {
		task.Status = TaskStatusFailed
		task.Error = err.Error()
	} else {
		task.Status = TaskStatusDone
	}
	task.FinishedAt = time.Now()
	_ = tm.saveTaskToDisk(task)
	tm.mu.Unlock()
}

// storage helpers
func (tm *TaskManager) taskMetaPath(id string) string {
	return filepath.Join(tm.dir, id+".json")
}

func (tm *TaskManager) saveTaskToDisk(t *Task) error {
	p := tm.taskMetaPath(t.ID)
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

func (tm *TaskManager) loadTasksFromDisk() {
	files, err := os.ReadDir(tm.dir)
	if err != nil {
		return
	}
	for _, fi := range files {
		if fi.IsDir() || filepath.Ext(fi.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(tm.dir, fi.Name()))
		if err != nil {
			continue
		}
		var t Task
		if err := json.Unmarshal(data, &t); err != nil {
			continue
		}
		tm.tasks[t.ID] = &t
		// 若之前是 running 或 pending，重新入队 pending
		if t.Status == TaskStatusPending || t.Status == TaskStatusRunning {
			tm.queue <- t.ID
		}
	}
}

// periodicCleaner 定时清理超过 retainDays 的任务及其文件
func (tm *TaskManager) periodicCleaner(interval time.Duration, retainDays int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		tm.CleanOlderThan(retainDays)
	}
}

// CleanOlderThan 删除 finish 时间早于 retainDays 天的任务和产物
func (tm *TaskManager) CleanOlderThan(retainDays int) {
	cutoff := time.Now().Add(-time.Duration(retainDays) * 24 * time.Hour)
	tm.mu.Lock()
	defer tm.mu.Unlock()
	for id, t := range tm.tasks {
		if (t.Status == TaskStatusDone || t.Status == TaskStatusFailed) && !t.FinishedAt.IsZero() && t.FinishedAt.Before(cutoff) {
			// 删除文件夹和元数据
			taskDir := filepath.Join(tm.dir, id)
			_ = os.RemoveAll(taskDir)
			_ = os.Remove(tm.taskMetaPath(id))
			delete(tm.tasks, id)
		}
	}
}

// StatusURL 返回任务状态查询的完整或相对 URL
func (tm *TaskManager) StatusURL(id string) string {
	return "/record/exportStatus?id=" + id
}

// DownloadURL 返回任务下载的完整或相对 URL
func (tm *TaskManager) DownloadURL(id string) string {
	return "/record/exportDownload?id=" + id
}
