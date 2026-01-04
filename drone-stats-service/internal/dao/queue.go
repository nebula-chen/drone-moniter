package dao

import (
	"encoding/json"
	"fmt"
	"time"

	"drone-stats-service/internal/model"

	bolt "go.etcd.io/bbolt"
)

// BoltDB 结构：使用名为 "points" 的 bucket，key 为 timestamp_nano_seq，value 为 JSON 编码的 []FlightTrackPoint

type Queue struct {
	db *bolt.DB
}

func NewQueue(path string) (*Queue, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists([]byte("points"))
		return e
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	return &Queue{db: db}, nil
}

func (q *Queue) Close() error {
	return q.db.Close()
}

// 将一批轨迹点入队（作为一条记录存储）
func (q *Queue) Enqueue(points []model.FlightTrackPoint) error {
	if len(points) == 0 {
		return nil
	}
	data, err := json.Marshal(points)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%d_%d", time.Now().UnixNano(), time.Now().Unix()%1000)
	return q.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("points"))
		return b.Put([]byte(key), data)
	})
}

// PeekBatch 返回最多 limit 个批次，结果为 map[key] -> points
func (q *Queue) PeekBatch(limit int) (map[string][]model.FlightTrackPoint, error) {
	out := make(map[string][]model.FlightTrackPoint)
	err := q.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("points"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		count := 0
		for k, v := c.First(); k != nil && count < limit; k, v = c.Next() {
			var pts []model.FlightTrackPoint
			if err := json.Unmarshal(v, &pts); err != nil {
				// 跳过解析失败的记录
				continue
			}
			out[string(k)] = pts
			count++
		}
		return nil
	})
	return out, err
}

// DeleteKeys 根据 keys 删除队列记录
func (q *Queue) DeleteKeys(keys []string) error {
	return q.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("points"))
		for _, k := range keys {
			if err := b.Delete([]byte(k)); err != nil {
				return err
			}
		}
		return nil
	})
}
