package mdb

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/moremorefun/mtool/mlog"

	"github.com/jmoiron/sqlx"

	// 导入mysql
	_ "github.com/go-sql-driver/mysql"
)

// 数据库数据类型
const (
	GoTypeString  = 1
	GoTypeInt64   = 2
	GoTypeBytes   = 3
	GoTypeFloat64 = 4
	GoTypeTime    = 5
)

// TypeMySQLToGoMap 类型转换关系
var TypeMySQLToGoMap = map[string]int64{
	"BIT":        1,
	"TEXT":       1,
	"BLOB":       3,
	"DATETIME":   5,
	"DOUBLE":     4,
	"ENUM":       1,
	"FLOAT":      4,
	"GEOMETRY":   1,
	"MEDIUMINT":  2,
	"JSON":       1,
	"INT":        2,
	"LONGTEXT":   1,
	"LONGBLOB":   3,
	"BIGINT":     2,
	"MEDIUMTEXT": 1,
	"MEDIUMBLOB": 3,
	"DATE":       5,
	"DECIMAL":    1,
	"SET":        1,
	"SMALLINT":   2,
	"BINARY":     3,
	"CHAR":       1,
	"TIME":       5,
	"TIMESTAMP":  5,
	"TINYINT":    2,
	"TINYTEXT":   1,
	"TINYBLOB":   3,
	"VARBINARY":  3,
	"VARCHAR":    1,
	"YEAR":       2,
}

// ExecuteAble 数据库接口
type ExecuteAble interface {
	Rebind(string) string

	Get(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Select(dest interface{}, query string, args ...interface{}) error

	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// isShowSQL 是否显示执行的sql语句
var isShowSQL bool

// Create 创建数据库链接
func Create(dataSourceName string, showSQL bool) *sqlx.DB {
	isShowSQL = showSQL

	var err error
	var db *sqlx.DB

	db, err = sqlx.Connect("mdb", dataSourceName)
	if err != nil {
		mlog.Log.Fatalf("db connect error: %s", err.Error())
		return nil
	}

	count := runtime.NumCPU()*20 + 1
	db.SetMaxOpenConns(count)
	db.SetMaxIdleConns(count)
	db.SetConnMaxLifetime(1 * time.Hour)

	err = db.Ping()
	if err != nil {
		mlog.Log.Fatalf("db ping error: %s", err.Error())
		return nil
	}
	return db
}

// SetShowSQL 设置是否显示sql
func SetShowSQL(b bool) {
	isShowSQL = b
}

// ExecuteLastIDContent 执行sql语句并返回lastID
func ExecuteLastIDContent(ctx context.Context, tx ExecuteAble, query string, argMap map[string]interface{}) (int64, error) {
	query, args, err := wrapSQL(query, argMap, tx)
	if err != nil {
		return 0, err
	}
	ret, err := tx.ExecContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return 0, err
	}
	lastID, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastID, nil
}

// ExecuteCountContent 执行sql语句返回执行个数
func ExecuteCountContent(ctx context.Context, tx ExecuteAble, query string, argMap map[string]interface{}) (int64, error) {
	query, args, err := wrapSQL(query, argMap, tx)
	if err != nil {
		return 0, err
	}
	ret, err := tx.ExecContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return 0, err
	}
	count, err := ret.RowsAffected()
	if err != nil {
		return 0, err
	}
	return count, nil
}

// ExecuteCountManyContent 返回sql语句并返回执行行数
func ExecuteCountManyContent(ctx context.Context, tx ExecuteAble, query string, n int, args ...interface{}) (int64, error) {
	var err error
	insertArgs := strings.Repeat("(?),", n)
	insertArgs = strings.TrimSuffix(insertArgs, ",")
	query = fmt.Sprintf(query, insertArgs)
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return 0, err
	}
	query = tx.Rebind(query)
	sqlLog(query, args)
	ret, err := tx.ExecContext(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return 0, err
	}
	count, err := ret.RowsAffected()
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetContent 执行sql查询并返回当个元素
func GetContent(ctx context.Context, tx ExecuteAble, dest interface{}, query string, argMap map[string]interface{}) (bool, error) {
	query, args, err := wrapSQL(query, argMap, tx)
	if err != nil {
		return false, err
	}
	err = tx.GetContext(
		ctx,
		dest,
		query,
		args...,
	)
	if err == sql.ErrNoRows {
		// 没有元素
		return false, nil
	}
	if err != nil {
		// 执行错误
		return false, err
	}
	return true, nil
}

// SelectContent 执行sql查询并返回多行
func SelectContent(ctx context.Context, tx ExecuteAble, dest interface{}, query string, argMap map[string]interface{}) error {
	query, args, err := wrapSQL(query, argMap, tx)
	if err != nil {
		return err
	}
	err = tx.SelectContext(
		ctx,
		dest,
		query,
		args...,
	)
	if err == sql.ErrNoRows {
		// 没有元素
		return nil
	}
	if err != nil {
		// 执行错误
		return err
	}
	return nil
}

// RowsContent 执行sql查询并返回多行
func RowsContent(ctx context.Context, tx ExecuteAble, query string, argMap map[string]interface{}) ([]map[string]interface{}, error) {
	query, args, err := wrapSQL(query, argMap, tx)
	if err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(
		ctx,
		query,
		args...,
	)
	if err == sql.ErrNoRows {
		// 没有元素
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	cts, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	l := len(cts)
	columns := make([]reflect.Value, l)
	columnsPoint := make([]interface{}, l)
	for i, ct := range cts {
		dbType := ct.DatabaseTypeName()
		goType, ok := TypeMySQLToGoMap[dbType]
		if !ok {
			return nil, fmt.Errorf("no db type: %s", dbType)
		}
		var tv reflect.Value
		switch goType {
		case GoTypeString:
			tv = reflect.New(reflect.TypeOf(""))
		case GoTypeInt64:
			tv = reflect.New(reflect.TypeOf(int64(0)))
		case GoTypeBytes:
			tv = reflect.New(reflect.TypeOf([]byte{}))
		case GoTypeFloat64:
			tv = reflect.New(reflect.TypeOf(float64(0)))
		case GoTypeTime:
			tv = reflect.New(reflect.TypeOf(time.Time{}))
		default:
			return nil, fmt.Errorf("no go type: %d", goType)
		}
		e := tv.Elem()
		columns[i] = e
		columnsPoint[i] = e.Addr().Interface()
	}
	var mapRows []map[string]interface{}
	for rows.Next() {
		err := rows.Scan(columnsPoint...)
		if err != nil {
			return nil, err
		}
		rowMap := map[string]interface{}{}
		for i, v := range columns {
			colName := cts[i].Name()
			rowMap[colName] = v.Interface()
		}
		mapRows = append(mapRows, rowMap)
	}
	return mapRows, nil
}

// Transaction 执行事物
func Transaction(ctx context.Context, db *sqlx.DB, f func(dbTx ExecuteAble) error) error {
	isComment := false
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if !isComment {
			_ = tx.Rollback()
		}
	}()
	err = f(tx)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	isComment = true
	return nil
}

// wrapSQL 打包sql
func wrapSQL(query string, argMap map[string]interface{}, tx ExecuteAble) (string, []interface{}, error) {
	query, args, err := sqlx.Named(query, argMap)
	if err != nil {
		return "", nil, err
	}
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return "", nil, err
	}
	query = tx.Rebind(query)
	sqlLog(query, args)
	return query, args, nil
}

func sqlLog(query string, args []interface{}) {
	if isShowSQL {
		queryStr := query + ";"
		for _, arg := range args {
			_, ok := arg.(string)
			if ok {
				queryStr = strings.Replace(queryStr, "?", fmt.Sprintf(`"%s"`, arg), 1)
			} else {
				queryStr = strings.Replace(queryStr, "?", fmt.Sprintf(`%v`, arg), 1)
			}
		}
		mlog.Log.Debugf("exec sql:\n%s;\n%#v", query, args)
	}
}
