package mquery

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"github.com/moremorefun/mtool/mmysql"
)

type selectData struct {
	columns      []SQLAble
	from         SQLAble
	joins        []SQLAble
	whereParts   []SQLAble
	groupBys     []SQLAble
	orderByParts []SQLAble
	offset       int64
	limit        int64
	isForUpdate  bool
	as           string
}

// Select 创建搜索
func Select() *selectData {
	var q selectData
	return &q
}

// Columns 字段
func (q *selectData) Columns(columns ...SQLAble) *selectData {
	q.columns = append(q.columns, columns...)
	return q
}

// ColumnsString 字段
func (q *selectData) ColumnsString(columns ...string) *selectData {
	for _, column := range columns {
		q.columns = append(q.columns, ConvertRaw(column))
	}
	return q
}

// ColumnsReset 字段
func (q *selectData) ColumnsReset() *selectData {
	q.columns = nil
	return q
}

// From 表名
func (q *selectData) From(from SQLAble) *selectData {
	q.from = from
	return q
}

// FromString 表名
func (q *selectData) FromString(from string) *selectData {
	q.from = ConvertRaw(from)
	return q
}

// Where 条件
func (q *selectData) Where(cond ...SQLAble) *selectData {
	q.whereParts = append(q.whereParts, cond...)
	return q
}

// GroupBys 分组
func (q *selectData) GroupBys(groupBys ...SQLAble) *selectData {
	q.groupBys = append(q.groupBys, groupBys...)
	return q
}

// GroupBysString 分组
func (q *selectData) GroupBysString(groupBys ...string) *selectData {
	for _, groupBy := range groupBys {
		q.groupBys = append(q.groupBys, ConvertRaw(groupBy))
	}
	return q
}

// OrderBys 排序
func (q *selectData) OrderBys(orders ...SQLAble) *selectData {
	q.orderByParts = append(q.orderByParts, orders...)
	return q
}

// OrderBysString 排序
func (q *selectData) OrderBysString(orders ...string) *selectData {
	for _, order := range orders {
		q.orderByParts = append(q.orderByParts, ConvertRaw(order))
	}
	return q
}

// Limit 限制
func (q *selectData) Limit(limit int64) *selectData {
	q.limit = limit
	return q
}

// Offset 偏移
func (q *selectData) Offset(offset int64) *selectData {
	q.offset = offset
	return q
}

// QueryJoin 链接
func (q *selectData) Join(join ...SQLAble) *selectData {
	q.joins = append(q.joins, join...)
	return q
}

// ForUpdate 加锁
func (q *selectData) ForUpdate() *selectData {
	q.isForUpdate = true
	return q
}

// As 设置为as
func (q *selectData) As(newName string) *selectData {
	q.as = newName
	return q
}

// AppendToQuery 添加输入
func (q *selectData) AppendToQuery(buf bytes.Buffer, arg map[string]interface{}) (bytes.Buffer, map[string]interface{}, error) {
	var err error
	if len(q.as) > 0 {
		buf.WriteString("(\n")
	}
	buf.WriteString("SELECT")
	if len(q.columns) == 0 {
		buf.WriteString("\n   *")
	} else {
		lastColumnIndex := len(q.columns) - 1
		for i, column := range q.columns {
			_, err = buf.WriteString("\n    ")
			if err != nil {
				return bytes.Buffer{}, nil, err
			}
			buf, arg, err = column.AppendToQuery(buf, arg)
			if err != nil {
				return bytes.Buffer{}, nil, err
			}
			if i != lastColumnIndex {
				buf.WriteString(",")
			}
		}
	}

	if q.from == nil {
		return bytes.Buffer{}, nil, fmt.Errorf("select no from")
	}
	buf.WriteString("\nFROM\n    ")
	buf, arg, err = q.from.AppendToQuery(buf, arg)
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	if len(q.joins) > 0 {
		for _, join := range q.joins {
			buf.WriteString("\n")
			buf, arg, err = join.AppendToQuery(buf, arg)
			if err != nil {
				return bytes.Buffer{}, nil, err
			}
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
				return bytes.Buffer{}, nil, err
			}
		}
	}
	if len(q.groupBys) > 0 {
		buf.WriteString("\nGROUP BY\n    ")
		for i, groupBy := range q.groupBys {
			if i != 0 {
				buf.WriteString(", ")
			}
			buf, arg, err = groupBy.AppendToQuery(buf, arg)
			if err != nil {
				return bytes.Buffer{}, nil, err
			}
		}
	}
	if len(q.orderByParts) > 0 {
		buf.WriteString("\nORDER BY\n    ")
		for i, orderByPart := range q.orderByParts {
			if i != 0 {
				buf.WriteString(", ")
			}
			buf, arg, err = orderByPart.AppendToQuery(buf, arg)
			if err != nil {
				return bytes.Buffer{}, nil, err
			}
		}
	}
	if q.limit > 0 {
		if q.offset > 0 {
			buf.WriteString("\nLIMIT ")
			buf.WriteString(strconv.FormatInt(q.offset, 10))
			buf.WriteString(", ")
			buf.WriteString(strconv.FormatInt(q.limit, 10))
		} else {
			buf.WriteString("\nLIMIT ")
			buf.WriteString(strconv.FormatInt(q.limit, 10))
		}
	}
	if q.isForUpdate {
		buf.WriteString("\nFOR UPDATE")
	}
	if len(q.as) > 0 {
		buf.WriteString("\n) AS ")
		buf.WriteString(q.as)
	}
	return buf, arg, nil
}

// ToSQL 生成sql
func (q *selectData) ToSQL() (string, map[string]interface{}, error) {
	var err error
	var buf bytes.Buffer
	arg := map[string]interface{}{}

	buf, arg, err = q.AppendToQuery(buf, arg)
	if err != nil {
		return "", nil, err
	}
	return buf.String(), arg, err
}

// DoGet 获取数据
func (q *selectData) DoGet(ctx context.Context, tx mmysql.DbExeAble, dest interface{}) (bool, error) {
	query, arg, err := q.Limit(1).ToSQL()
	if err == ErrInValueLenZero {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return mmysql.DbGetNamedContent(
		ctx,
		tx,
		dest,
		query,
		arg,
	)
}

// DoSelect 获取数据
func (q *selectData) DoSelect(ctx context.Context, tx mmysql.DbExeAble, dest interface{}) error {
	query, arg, err := q.ToSQL()
	if err == ErrInValueLenZero {
		return nil
	}
	if err != nil {
		return err
	}
	return mmysql.DbSelectNamedContent(
		ctx,
		tx,
		dest,
		query,
		arg,
	)
}

// RowInterface 获取数据
func (q *selectData) Row(ctx context.Context, tx mmysql.DbExeAble) (map[string]interface{}, error) {
	rows, err := q.Limit(1).Rows(
		ctx,
		tx,
	)
	if err == ErrInValueLenZero {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

// RowsInterface 获取数据
func (q *selectData) Rows(ctx context.Context, tx mmysql.DbExeAble) ([]map[string]interface{}, error) {
	query, arg, err := q.ToSQL()
	if err == ErrInValueLenZero {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mmysql.DbRowsNamedContent(ctx, tx, query, arg)
}
