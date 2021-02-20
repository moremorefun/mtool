package mquery

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var globalIndex int64

// ErrInValueLenZero in 条件数据长度为0
var ErrInValueLenZero = errors.New("sql in values len 0")

// getK 获取key
func getK(old string) string {
	globalIndex++
	old = strings.ReplaceAll(old, ".", "_")
	old = strings.ReplaceAll(old, "`", "_")
	var buf bytes.Buffer
	buf.WriteString(old)
	buf.WriteString("_")
	buf.WriteString(strconv.FormatInt(globalIndex, 10))
	return buf.String()
}

// ConvertRaw 原样生成
type ConvertRaw string

// AppendToQuery 写入sql,填充arg
func (o ConvertRaw) AppendToQuery(buf bytes.Buffer, arg map[string]interface{}) (bytes.Buffer, map[string]interface{}, error) {
	_, err := buf.WriteString(string(o))
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	return buf, arg, nil
}

// ConvertKv kv结构
type ConvertKv struct {
	K string
	V interface{}
}

// ConvertKvStr kv字符串
type ConvertKvStr struct {
	K string
	V string
}

// ConvertFuncColAs 字符串
type ConvertFuncColAs struct {
	Func string
	Col  string
	As   string
}

// ConvertEq k=:k or k IN (:k)
type ConvertEq ConvertKv

// ConvertEqMake 生成
func ConvertEqMake(k string, v interface{}) ConvertEq {
	return ConvertEq{
		K: k,
		V: v,
	}
}

// AppendToQuery 写入sql,填充arg
func (o ConvertEq) AppendToQuery(buf bytes.Buffer, arg map[string]interface{}) (bytes.Buffer, map[string]interface{}, error) {
	k := getK(o.K)

	buf.WriteString(o.K)
	rt := reflect.TypeOf(o.V)
	switch rt.Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(o.V)
		if s.Len() == 0 {
			return bytes.Buffer{}, nil, ErrInValueLenZero
		}
		buf.WriteString(" IN (:")
		buf.WriteString(k)
		buf.WriteString(")")
	default:
		buf.WriteString("=:")
		buf.WriteString(k)
	}
	arg[k] = o.V
	return buf, arg, nil
}

// ConvertAdd k=k+:k
type ConvertAdd ConvertKv

// ConvertAddMake 生成
func ConvertAddMake(k string, v interface{}) ConvertAdd {
	return ConvertAdd{
		K: k,
		V: v,
	}
}

// AppendToQuery 写入sql,填充arg
func (o ConvertAdd) AppendToQuery(buf bytes.Buffer, arg map[string]interface{}) (bytes.Buffer, map[string]interface{}, error) {
	k := getK(o.K)

	_, err := buf.WriteString(o.K)
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString("=")
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString(o.K)
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString("+:")
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString(k)
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	arg[k] = o.V
	return buf, arg, nil
}

// ConvertGt k>:k
type ConvertGt ConvertKv

// ConvertGtMake 生成
func ConvertGtMake(k string, v interface{}) ConvertGt {
	return ConvertGt{
		K: k,
		V: v,
	}
}

// AppendToQuery 写入sql,填充arg
func (o ConvertGt) AppendToQuery(buf bytes.Buffer, arg map[string]interface{}) (bytes.Buffer, map[string]interface{}, error) {
	k := getK(o.K)

	_, err := buf.WriteString(o.K)
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString(">:")
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString(k)
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	arg[k] = o.V
	return buf, arg, nil
}

// ConvertLt k<:k
type ConvertLt ConvertKv

// ConvertLtMake 生成
func ConvertLtMake(k string, v interface{}) ConvertLt {
	return ConvertLt{
		K: k,
		V: v,
	}
}

// AppendToQuery 写入sql,填充arg
func (o ConvertLt) AppendToQuery(buf bytes.Buffer, arg map[string]interface{}) (bytes.Buffer, map[string]interface{}, error) {
	k := getK(o.K)

	_, err := buf.WriteString(o.K)
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString("<:")
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString(k)
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	arg[k] = o.V
	return buf, arg, nil
}

// ConvertEqRaw k=v
type ConvertEqRaw ConvertKvStr

// ConvertEqRawMake 生成
func ConvertEqRawMake(k, v string) ConvertEqRaw {
	return ConvertEqRaw{
		K: k,
		V: v,
	}
}

// AppendToQuery 写入sql,填充arg
func (o ConvertEqRaw) AppendToQuery(buf bytes.Buffer, arg map[string]interface{}) (bytes.Buffer, map[string]interface{}, error) {
	_, err := buf.WriteString(o.K)
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString("=")
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString(o.V)
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	return buf, arg, nil
}

// ConvertDesc k DESC
type ConvertDesc string

// AppendToQuery 写入sql,填充arg
func (o ConvertDesc) AppendToQuery(buf bytes.Buffer, arg map[string]interface{}) (bytes.Buffer, map[string]interface{}, error) {
	_, err := buf.WriteString(string(o))
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString(" DESC")
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	return buf, arg, nil
}

// ConvertValues k=VALUES(k)
type ConvertValues string

// AppendToQuery 写入sql,填充arg
func (o ConvertValues) AppendToQuery(buf bytes.Buffer, arg map[string]interface{}) (bytes.Buffer, map[string]interface{}, error) {
	_, err := buf.WriteString(string(o))
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString("=VALUES(")
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString(string(o))
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString(")")
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	return buf, arg, nil
}

// ConvertOr 或条件
type ConvertOr struct {
	Left  SQLAble
	Right SQLAble
}

// ConvertOrMake 生成
func ConvertOrMake(left, right SQLAble) ConvertOr {
	return ConvertOr{
		Left:  left,
		Right: right,
	}
}

// AppendToQuery 写入sql,填充arg
func (o ConvertOr) AppendToQuery(buf bytes.Buffer, arg map[string]interface{}) (bytes.Buffer, map[string]interface{}, error) {
	var err error
	if o.Left == nil || o.Right == nil {
		return bytes.Buffer{}, nil, fmt.Errorf("or empty")
	}
	buf.WriteString("(")
	buf, arg, err = o.Left.AppendToQuery(buf, arg)
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	buf.WriteString(" or ")
	buf, arg, err = o.Right.AppendToQuery(buf, arg)
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	buf.WriteString(" )")
	return buf, arg, nil
}

// ConvertFuncAs func(col) AS as
type ConvertFuncAs ConvertFuncColAs

// ConvertFuncAsMake 生成
func ConvertFuncAsMake(f, col, as string) ConvertFuncAs {
	return ConvertFuncAs{
		Func: f,
		Col:  col,
		As:   as,
	}
}

// AppendToQuery 写入sql,填充arg
func (o ConvertFuncAs) AppendToQuery(buf bytes.Buffer, arg map[string]interface{}) (bytes.Buffer, map[string]interface{}, error) {
	_, err := buf.WriteString(o.Func)
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString("(")
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString(o.Col)
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString(") AS ")
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	_, err = buf.WriteString(o.As)
	if err != nil {
		return bytes.Buffer{}, nil, err
	}
	return buf, arg, nil
}
