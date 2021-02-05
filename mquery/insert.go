package mquery

import (
	"bytes"
	"context"
	"fmt"

	"github.com/moremorefun/mtool/mdb"
)

type insertData struct {
	isIgnore       bool
	into           string
	columns        []string
	values         []interface{}
	duplicateParts []SQLAble
}

// Insert 创建搜索
func Insert() *insertData {
	var q insertData
	return &q
}

// Ignore 忽略
func (q *insertData) Ignore() *insertData {
	q.isIgnore = true
	return q
}

// Into 表名
func (q *insertData) Into(into string) *insertData {
	q.into = into
	return q
}

// Columns 列
func (q *insertData) Columns(columns ...string) *insertData {
	q.columns = columns
	return q
}

// Values 值
func (q *insertData) Values(values ...interface{}) *insertData {
	q.values = append(q.values, values...)
	return q
}

// Duplicates 替换
func (q *insertData) Duplicates(duplicates ...SQLAble) *insertData {
	q.duplicateParts = append(q.duplicateParts, duplicates...)
	return q
}

// ToSQL 生成sql
func (q *insertData) ToSQL() (string, map[string]interface{}, error) {
	var err error

	var buf bytes.Buffer
	arg := map[string]interface{}{}
	buf.WriteString("INSERT")
	if q.isIgnore {
		buf.WriteString(" IGNORE")
	}
	buf.WriteString(" INTO ")
	if len(q.into) == 0 {
		return "", nil, fmt.Errorf("insert no into")
	}
	buf.WriteString(q.into)
	if len(q.columns) == 0 {
		return "", nil, fmt.Errorf("insert no columns")
	}
	buf.WriteString(" (")
	lastColumnIndex := len(q.columns) - 1
	for i, column := range q.columns {
		buf.WriteString("\n    ")
		buf.WriteString(column)
		if i != lastColumnIndex {
			buf.WriteString(",")
		}
	}
	buf.WriteString("\n) VALUES")
	if len(q.values) == 0 {
		return "", nil, fmt.Errorf("insert values empty")
	}
	lastValueIndex := len(q.values) - 1
	for i, value := range q.values {
		k := fmt.Sprintf("value%d", i)
		buf.WriteString("\n(:")
		buf.WriteString(k)
		buf.WriteString(")")
		if i != lastValueIndex {
			buf.WriteString(",")
		}
		arg[k] = value
	}
	if len(q.duplicateParts) > 0 {
		buf.WriteString("\nON DUPLICATE KEY UPDATE")
		lastDuplicateIndex := len(q.duplicateParts) - 1
		for i, duplicate := range q.duplicateParts {
			buf.WriteString("\n    ")
			buf, arg, err = duplicate.AppendToQuery(buf, arg)
			if err != nil {
				return "", nil, err
			}
			if i != lastDuplicateIndex {
				buf.WriteString(",")
			}
		}
	}
	return buf.String(), arg, nil
}

// DoExecuteLastID 获取最后一个插入id
func (q *insertData) DoExecuteLastID(ctx context.Context, tx mdb.ExecuteAble) (int64, error) {
	query, arg, err := q.ToSQL()
	if err != nil {
		return 0, err
	}
	return mdb.ExecuteLastIDContent(
		ctx,
		tx,
		query,
		arg,
	)
}

// DoExecuteCount 获取执行行数
func (q *insertData) DoExecuteCount(ctx context.Context, tx mdb.ExecuteAble) (int64, error) {
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
