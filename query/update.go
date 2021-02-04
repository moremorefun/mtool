package query

import (
	"bytes"
	"context"
	"fmt"

	"github.com/moremorefun/mtool/mysql"
)

type updateData struct {
	table       string
	updateParts []SQLAble
	whereParts  []SQLAble
}

// Update 创建更新
func Update() *updateData {
	var q updateData
	return &q
}

// Table 表名
func (q *updateData) Table(table string) *updateData {
	q.table = table
	return q
}

// Update 更新内容
func (q *updateData) Update(updateParts ...SQLAble) *updateData {
	q.updateParts = append(q.updateParts, updateParts...)
	return q
}

// Where 条件
func (q *updateData) Where(whereParts ...SQLAble) *updateData {
	q.whereParts = append(q.whereParts, whereParts...)
	return q
}

// ToSQL 生成sql
func (q *updateData) ToSQL() (string, map[string]interface{}, error) {
	var err error
	var buf bytes.Buffer
	arg := map[string]interface{}{}

	buf.WriteString("UPDATE\n    ")
	if len(q.table) == 0 {
		return "", nil, fmt.Errorf("update no table")
	}
	buf.WriteString(q.table)
	buf.WriteString("\nSET")
	if len(q.updateParts) == 0 {
		return "", nil, fmt.Errorf("update set empty")
	}
	lastUpdateIndex := len(q.updateParts) - 1
	for i, updatePart := range q.updateParts {
		buf.WriteString("\n    ")
		buf, arg, err = updatePart.AppendToQuery(buf, arg)
		if err != nil {
			return "", nil, err
		}
		if i != lastUpdateIndex {
			buf.WriteString(",")
		}
	}
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
func (q *updateData) DoExecuteCount(ctx context.Context, tx mysql.DbExeAble) (int64, error) {
	query, arg, err := q.ToSQL()
	if err != nil {
		return 0, err
	}
	return mysql.DbExecuteCountNamedContent(
		ctx,
		tx,
		query,
		arg,
	)
}
