package mquery

import (
	"bytes"
	"fmt"

	"github.com/gin-gonic/gin"
)

// join类型
const (
	JoinTypeInner = 1
	JoinTypeLeft  = 2
)

type joinData struct {
	joinType int64
	table    SQLAble
	onParts  []SQLAble
}

// Join 链接
func Join(joinType int64) *joinData {
	j := joinData{
		joinType: joinType,
	}
	return &j
}

// Table 表名
func (q *joinData) Table(from SQLAble) *joinData {
	q.table = from
	return q
}

// FromString 表名
func (q *joinData) TableString(from string) *joinData {
	q.table = ConvertRaw(from)
	return q
}

// On 链接条件
func (q *joinData) On(cond ...SQLAble) *joinData {
	q.onParts = append(q.onParts, cond...)
	return q
}

// ToSQL 生成sql
func (q *joinData) AppendToQuery(buf bytes.Buffer, arg gin.H) (bytes.Buffer, gin.H, error) {
	var err error
	switch q.joinType {
	case JoinTypeInner:
		_, err := buf.WriteString("INNER JOIN ")
		if err != nil {
			return bytes.Buffer{}, nil, err
		}
	case JoinTypeLeft:
		_, err := buf.WriteString("LEFT JOIN ")
		if err != nil {
			return bytes.Buffer{}, nil, err
		}
	default:
		return bytes.Buffer{}, nil, fmt.Errorf("no joinData type: %d", q.joinType)
	}
	if q.table == nil {
		return bytes.Buffer{}, nil, fmt.Errorf("join no table")
	}
	buf, arg, err = q.table.AppendToQuery(buf, arg)
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	if len(q.onParts) == 0 {
		return bytes.Buffer{}, nil, fmt.Errorf("join no on")
	}
	buf.WriteString(" ON (")
	for i, on := range q.onParts {
		buf.WriteString("\n    ")
		if i != 0 {
			buf.WriteString("AND ")
		}
		buf, arg, err = on.AppendToQuery(buf, arg)
		if err != nil {
			return bytes.Buffer{}, nil, err
		}
	}
	buf.WriteString("\n)")
	return buf, arg, nil
}
