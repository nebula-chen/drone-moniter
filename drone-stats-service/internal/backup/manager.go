package backup

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// 备份管理器：负责导出 MySQL 为 SQL 文件，以及导出 Influx 数据为 line-protocol 文件。
// 导出的 MySQL 文件可以通过管道导入到 mysql 容器：
// cat mysql_...sql | docker exec -i drone-mysql sh -c 'mysql -uroot -proot123456'
// 导出的 Influx 为 line-protocol 文件，可以通过 influx write 导入：
// docker cp backups/influx_... influxdb:/tmp/influx_lp && docker exec -i influxdb influx write --bucket <bucket> --file /tmp/influx_lp/points.lp --org <org> --token <token>

type Manager struct {
	MySQLDB       *sql.DB
	InfluxClient  influxdb2.Client
	InfluxQuery   api.QueryAPI
	InfluxOrg     string
	InfluxBucket  string
	BackupDir     string
	RetentionDays int
}

// NewManager 创建备份管理器
func NewManager(mysqlDB *sql.DB, influxClient influxdb2.Client, influxOrg, influxBucket, backupDir string, retention int) *Manager {
	return &Manager{
		MySQLDB:       mysqlDB,
		InfluxClient:  influxClient,
		InfluxQuery:   influxClient.QueryAPI(influxOrg),
		InfluxOrg:     influxOrg,
		InfluxBucket:  influxBucket,
		BackupDir:     backupDir,
		RetentionDays: retention,
	}
}

// BackupOnce 执行一次完整备份（MySQL SQL + Influx line-protocol）
func (m *Manager) BackupOnce(ctx context.Context) error {
	ts := time.Now().Format("20060102_150405")
	targetDir := filepath.Join(m.BackupDir, fmt.Sprintf("backup_%s", ts))
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	// 1) 导出 MySQL
	sqlFile := filepath.Join(targetDir, fmt.Sprintf("mysql_%s.sql", ts))
	if err := m.dumpMySQL(sqlFile); err != nil {
		return fmt.Errorf("mysql dump failed: %w", err)
	}

	// 2) 导出 Influx 为 line-protocol
	lpFile := filepath.Join(targetDir, fmt.Sprintf("influx_%s.lp", ts))
	if err := m.exportInfluxLP(ctx, lpFile); err != nil {
		return fmt.Errorf("influx export failed: %w", err)
	}

	// 3) 清理过期备份
	if err := m.cleanupOld(); err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	return nil
}

// dumpMySQL 导出当前数据库的 schema + data 为可直接导入的 SQL 文件
func (m *Manager) dumpMySQL(outPath string) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// 基本 header
	f.WriteString("SET FOREIGN_KEY_CHECKS=0;\n")

	// 获取当前数据库名
	var dbName sql.NullString
	if err := m.MySQLDB.QueryRow("SELECT DATABASE()").Scan(&dbName); err != nil {
		return err
	}
	db := ""
	if dbName.Valid {
		db = dbName.String
	}
	if db != "" {
		f.WriteString(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`;\nUSE `%s`;\n", db, db))
	}

	// 获取表列表
	rows, err := m.MySQLDB.Query("SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE()")
	if err != nil {
		return err
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err == nil {
			tables = append(tables, t)
		}
	}

	// 导出每个表的 CREATE 和 INSERT
	for _, t := range tables {
		// SHOW CREATE TABLE
		var table string
		var createSQL string
		if err := m.MySQLDB.QueryRow(fmt.Sprintf("SHOW CREATE TABLE `%s`", t)).Scan(&table, &createSQL); err != nil {
			return err
		}
		f.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\n", t))
		f.WriteString(createSQL + ";\n\n")

		// 导出数据（批量 INSERT）
		q := fmt.Sprintf("SELECT * FROM `%s`", t)
		r, err := m.MySQLDB.Query(q)
		if err != nil {
			return err
		}
		cols, err := r.Columns()
		if err != nil {
			r.Close()
			return err
		}
		colCount := len(cols)
		batchSize := 500
		vals := make([]interface{}, colCount)
		ptrs := make([]interface{}, colCount)
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		inserts := 0
		var valueBuf strings.Builder
		for r.Next() {
			if err := r.Scan(ptrs...); err != nil {
				r.Close()
				return err
			}
			// build value tuple
			if inserts%batchSize == 0 {
				if inserts != 0 {
					// flush previous
					f.WriteString(fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s;\n", t, strings.Join(cols, ", "), valueBuf.String()))
					valueBuf.Reset()
				}
			} else {
				valueBuf.WriteString(",")
			}
			valueBuf.WriteString("(")
			for i := 0; i < colCount; i++ {
				if i > 0 {
					valueBuf.WriteString(",")
				}
				valueBuf.WriteString(escapeSQLValue(vals[i]))
			}
			valueBuf.WriteString(")")
			inserts++
		}
		if inserts > 0 {
			// flush remaining
			f.WriteString(fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s;\n\n", t, strings.Join(cols, ", "), valueBuf.String()))
		}
		r.Close()
	}

	f.WriteString("SET FOREIGN_KEY_CHECKS=1;\n")
	return nil
}

// escapeSQLValue 将单元格值转换为 SQL 字面量
func escapeSQLValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case []byte:
		// []byte 用作字符串
		s := string(val)
		return quoteStringForSQL(s)
	case string:
		return quoteStringForSQL(val)
	case time.Time:
		return quoteStringForSQL(val.Format("2006-01-02 15:04:05"))
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%v", val)
	case bool:
		if val {
			return "1"
		}
		return "0"
	default:
		// fallback to字符串表示并引用
		return quoteStringForSQL(fmt.Sprintf("%v", val))
	}
}

func quoteStringForSQL(s string) string {
	// 简单转义单引号
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	return fmt.Sprintf("'%s'", s)
}

// exportInfluxLP 导出 Influx 数据为 line-protocol 文件
func (m *Manager) exportInfluxLP(ctx context.Context, outPath string) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// flux 查询：从最早数据开始导出 bucket 全量数据
	flux := fmt.Sprintf(`from(bucket: "%s") |> range(start: 0)`, m.InfluxBucket)
	result, err := m.InfluxQuery.Query(ctx, flux)
	if err != nil {
		return err
	}
	defer result.Close()

	for result.Next() {
		rec := result.Record()
		// 构造 line-protocol：measurement,tags field=value timestamp
		measurement := rec.Measurement()
		if measurement == "" {
			measurement = "measurement"
		}
		// tags: record.Values() 包含标签字段，排除内部字段
		vals := rec.Values()
		// 收集 tags
		tags := make([]string, 0)
		for k, v := range vals {
			if strings.HasPrefix(k, "_") || k == "result" || k == "table" || k == "_value" || k == "_field" || k == "_time" {
				continue
			}
			// 仅当值为 string 时作为 tag
			if sv, ok := v.(string); ok {
				tags = append(tags, fmt.Sprintf("%s=%s", sanitizeTag(k), sanitizeTag(sv)))
			}
		}
		tagStr := ""
		if len(tags) > 0 {
			tagStr = "," + strings.Join(tags, ",")
		}
		// field
		field := rec.Field()
		value := rec.Value()
		fieldVal := formatLPField(value)
		ts := rec.Time().UnixNano()
		line := fmt.Sprintf("%s%s %s=%s %d\n", measurement, tagStr, sanitizeField(field), fieldVal, ts)
		if _, err := f.WriteString(line); err != nil {
			return err
		}
	}
	if result.Err() != nil {
		return result.Err()
	}
	return nil
}

func sanitizeTag(s string) string {
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "=", "\\=")
	s = strings.ReplaceAll(s, " ", "\\ ")
	return s
}

func sanitizeField(s string) string {
	s = strings.ReplaceAll(s, " ", "\\ ")
	return s
}

func formatLPField(v interface{}) string {
	switch val := v.(type) {
	case string:
		// string fields must be quoted
		esc := strings.ReplaceAll(val, "\"", "\\\"")
		return fmt.Sprintf("\"%s\"", esc)
	case int64, int32, int, int8, int16:
		return fmt.Sprintf("%di", val)
	case uint64, uint32, uint, uint8, uint16:
		return fmt.Sprintf("%di", val)
	case float64, float32:
		return fmt.Sprintf("%v", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		// fallback to quoted string
		s := fmt.Sprintf("%v", val)
		esc := strings.ReplaceAll(s, "\"", "\\\"")
		return fmt.Sprintf("\"%s\"", esc)
	}
}

// cleanupOld 删除超过保留期的备份目录
func (m *Manager) cleanupOld() error {
	if m.RetentionDays <= 0 {
		return nil
	}
	cutoff := time.Now().AddDate(0, 0, -m.RetentionDays)
	entries, err := os.ReadDir(m.BackupDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.RemoveAll(filepath.Join(m.BackupDir, e.Name()))
		}
	}
	return nil
}
