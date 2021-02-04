package query

import "bytes"

// SQLAble sql语句生成接口
type SQLAble interface {
	AppendToQuery(bytes.Buffer, map[string]interface{}) (bytes.Buffer, map[string]interface{}, error)
}
