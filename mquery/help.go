package mquery

import (
	"context"
	"fmt"
	"strings"

	"github.com/moremorefun/mtool/mutils"

	"github.com/moremorefun/mtool/mdb"

	jsoniter "github.com/json-iterator/go"
)

// FormatMapKey 格式化字段名到key
func FormatMapKey(oldKey string) string {
	oldKey = strings.ReplaceAll(oldKey, "`", "")
	lastIndex := strings.LastIndex(oldKey, ".")
	if lastIndex != -1 {
		oldKey = oldKey[lastIndex+1:]
	}
	return oldKey
}

// GetValuesFromRows 获取values
func GetValuesFromRows(rows []map[string]interface{}, key string) ([]interface{}, error) {
	key = FormatMapKey(key)
	var values []interface{}
	for _, row := range rows {
		v, ok := row[key]
		if !ok {
			return nil, fmt.Errorf("no key: %s", key)
		}
		if !mutils.IsInSlice(values, v) {
			values = append(values, v)
		}
	}
	return values, nil
}

// GetValuesFromMap 获取values
func GetValuesFromMap(m map[string]map[string]interface{}, key string) ([]interface{}, error) {
	key = FormatMapKey(key)
	var values []interface{}
	for _, row := range m {
		v, ok := row[key]
		if !ok {
			return nil, fmt.Errorf("no key: %s", key)
		}
		if !mutils.IsInSlice(values, v) {
			values = append(values, v)
		}
	}
	return values, nil
}

// GetValuesFromMapRows 获取values
func GetValuesFromMapRows(ms map[string][]map[string]interface{}, key string) ([]interface{}, error) {
	key = FormatMapKey(key)
	var values []interface{}
	for _, rows := range ms {
		for _, row := range rows {
			v, ok := row[key]
			if !ok {
				return nil, fmt.Errorf("no key: %s", key)
			}
			if !mutils.IsInSlice(values, v) {
				values = append(values, v)
			}
		}
	}
	return values, nil
}

// SelectRows2One 获取关联map
func SelectRows2One(ctx context.Context, tx mdb.ExecuteAble, sourceRows []map[string]interface{}, sourceKey, targetTableName, targetKey string, targetColumns []string) (map[string]map[string]interface{}, []interface{}, error) {
	keyValues, err := GetValuesFromRows(sourceRows, sourceKey)
	if err != nil {
		return nil, nil, err
	}
	targetMap, err := SelectKeys2One(ctx, tx, keyValues, targetTableName, targetKey, targetColumns)
	return targetMap, keyValues, err
}

// SelectRows2Many 获取关联map
func SelectRows2Many(ctx context.Context, tx mdb.ExecuteAble, sourceRows []map[string]interface{}, sourceKey, targetTableName, targetKey string, targetColumns []string) (map[string][]map[string]interface{}, []interface{}, error) {
	keyValues, err := GetValuesFromRows(sourceRows, sourceKey)
	if err != nil {
		return nil, nil, err
	}
	targetMap, err := SelectKeys2Many(ctx, tx, keyValues, targetTableName, targetKey, targetColumns)
	return targetMap, keyValues, err
}

// SelectKeys2One 获取关联map
func SelectKeys2One(ctx context.Context, tx mdb.ExecuteAble, keyValues []interface{}, targetTableName, targetKey string, targetColumns []string) (map[string]map[string]interface{}, error) {
	if len(keyValues) == 0 {
		return nil, nil
	}
	if len(targetColumns) != 0 {
		if !mutils.IsStringInSlice(targetColumns, targetKey) {
			targetColumns = append(targetColumns, targetKey)
		}
	}
	targetRows, err := Select().
		ColumnsString(targetColumns...).
		FromString(targetTableName).
		Where(ConvertEqMake(targetKey, keyValues)).
		Rows(
			ctx,
			tx,
		)
	if err != nil {
		return nil, err
	}
	mapTargetKey := FormatMapKey(targetKey)
	targetMap := map[string]map[string]interface{}{}
	for _, targetRow := range targetRows {
		kv, ok := targetRow[mapTargetKey]
		if !ok {
			return nil, fmt.Errorf("no target key: %s", mapTargetKey)
		}
		k := fmt.Sprintf("%v", kv)
		targetMap[k] = targetRow
	}
	return targetMap, nil
}

// SelectKeys2Many 获取关联map
func SelectKeys2Many(ctx context.Context, tx mdb.ExecuteAble, keyValues []interface{}, targetTableName, targetKey string, targetColumns []string) (map[string][]map[string]interface{}, error) {
	if len(keyValues) == 0 {
		return nil, nil
	}
	if len(targetColumns) != 0 {
		if !mutils.IsStringInSlice(targetColumns, targetKey) {
			targetColumns = append(targetColumns, targetKey)
		}
	}
	targetRows, err := Select().
		ColumnsString(targetColumns...).
		FromString(targetTableName).
		Where(ConvertEqMake(targetKey, keyValues)).
		Rows(
			ctx,
			tx,
		)
	if err != nil {
		return nil, err
	}
	mapTargetKey := FormatMapKey(targetKey)
	targetMap := map[string][]map[string]interface{}{}
	for _, targetRow := range targetRows {
		kv, ok := targetRow[mapTargetKey]
		if !ok {
			return nil, fmt.Errorf("no target key: %s", mapTargetKey)
		}
		k := fmt.Sprintf("%v", kv)
		targetMap[k] = append(targetMap[k], targetRow)
	}
	return targetMap, nil
}

// InterfaceToStruct 转换到struct
func InterfaceToStruct(inc interface{}, s interface{}) error {
	b, err := jsoniter.Marshal(inc)
	if err != nil {
		return err
	}
	err = jsoniter.Unmarshal(b, s)
	if err != nil {
		return err
	}
	return nil
}
