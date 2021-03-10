package mquery

import (
	"bytes"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/moremorefun/mtool/mdb"
)

type deleteData struct {
	table      string
	whereParts []SQLAble
}

// Delete 创建删除
func Delete() *deleteData {
	var q deleteData
	return &q
}

// Table 表名
func (q *deleteData) Table(table string) *deleteData {
	q.table = table
	return q
}

// Where 条件
func (q *deleteData) Where(whereParts ...SQLAble) *deleteData {
	q.whereParts = append(q.whereParts, whereParts...)
	return q
}

// ToSQL 生成sql
func (q *deleteData) ToSQL() (string, gin.H, error) {
	var err error
	var buf bytes.Buffer
	arg := gin.H{}

	buf.WriteString("DELETE\nFROM\n    ")
	if len(q.table) == 0 {
		return "", nil, fmt.Errorf("delete no table")
	}
	buf.WriteString(q.table)
	if len(q.whereParts) > 0 {
		buf.WriteString("\nWHERE")
		for i, where := range q.whereParts {
			buf.WriteString("\n    ")
			if i != 0 {
				buf.WriteString("AND ")
			}
			buf, arg, err = where.AppendToQuery(buf, arg)
			if err != nil {
				return "", nil, err
			}
		}
	}
	return buf.String(), arg, nil
}

// DoExecuteCount 获取执行行数
func (q *deleteData) DoExecuteCount(ctx context.Context, tx mdb.ExecuteAble) (int64, error) {
	query, arg, err := q.ToSQL()
	if err != nil {
		return 0, err
	}
	return mdb.ExecuteCountContent(
		ctx,
		tx,
		query,
		arg,
	)
}
