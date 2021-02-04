package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/moremorefun/mtool/log"

	"github.com/jmoiron/sqlx"
)

// 数据库数据类型
const (
	DbSQLGoTypeString  = 1
	DbSQLGoTypeInt64   = 2
	DbSQLGoTypeBytes   = 3
	DbSQLGoTypeFloat64 = 4
	DbSQLGoTypeTime    = 5
)

// DbSQLTypeToGoMap 类型转换关系
var DbSQLTypeToGoMap = map[string]int64{
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

// DbExeAble 数据库接口
type DbExeAble interface {
	Rebind(string) string
	Get(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Select(dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error

	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
}

// isShowSQL 是否显示执行的sql语句
var isShowSQL bool

// DbCreate 创建数据库链接
func DbCreate(dataSourceName string, showSQL bool) *sqlx.DB {
	isShowSQL = showSQL

	var err error
	var db *sqlx.DB

	db, err = sqlx.Connect("mysql", dataSourceName)
	if err != nil {
		log.Log.Fatalf("db connect error: %s", err.Error())
		return nil
	}

	count := runtime.NumCPU()*20 + 1
	db.SetMaxOpenConns(count)
	db.SetMaxIdleConns(count)
	db.SetConnMaxLifetime(1 * time.Hour)

	err = db.Ping()
	if err != nil {
		log.Log.Fatalf("db ping error: %s", err.Error())
		return nil
	}
	return db
}

// DbSetShowSQL 设置是否显示sql
func DbSetShowSQL(b bool) {
	isShowSQL = b
}

// DbExecuteLastIDNamedContent 执行sql语句并返回lastID
func DbExecuteLastIDNamedContent(ctx context.Context, tx DbExeAble, query string, argMap map[string]interface{}) (int64, error) {
	query, args, err := sqlx.Named(query, argMap)
	if err != nil {
		return 0, err
	}
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
	lastID, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastID, nil
}

// DbExecuteCountNamedContent 执行sql语句返回执行个数
func DbExecuteCountNamedContent(ctx context.Context, tx DbExeAble, query string, argMap map[string]interface{}) (int64, error) {
	query, args, err := sqlx.Named(query, argMap)
	if err != nil {
		return 0, err
	}
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

// DbGetNamedContent 执行sql查询并返回当个元素
func DbGetNamedContent(ctx context.Context, tx DbExeAble, dest interface{}, query string, argMap map[string]interface{}) (bool, error) {
	query, args, err := sqlx.Named(query, argMap)
	if err != nil {
		return false, err
	}
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return false, err
	}
	query = tx.Rebind(query)
	sqlLog(query, args)
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

// DbSelectNamedContent 执行sql查询并返回多行
func DbSelectNamedContent(ctx context.Context, tx DbExeAble, dest interface{}, query string, argMap map[string]interface{}) error {
	query, args, err := sqlx.Named(query, argMap)
	if err != nil {
		return err
	}
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return err
	}
	query = tx.Rebind(query)
	sqlLog(query, args)
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

// DbRowsNamedContent 执行sql查询并返回多行
func DbRowsNamedContent(ctx context.Context, tx DbExeAble, query string, argMap map[string]interface{}) ([]map[string]interface{}, error) {
	query, args, err := sqlx.Named(query, argMap)
	if err != nil {
		return nil, err
	}
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return nil, err
	}
	query = tx.Rebind(query)
	sqlLog(query, args)
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
		goType, ok := DbSQLTypeToGoMap[dbType]
		if !ok {
			return nil, fmt.Errorf("no db type: %s", dbType)
		}
		var tv reflect.Value
		switch goType {
		case DbSQLGoTypeString:
			tv = reflect.New(reflect.TypeOf(""))
		case DbSQLGoTypeInt64:
			tv = reflect.New(reflect.TypeOf(int64(0)))
		case DbSQLGoTypeBytes:
			tv = reflect.New(reflect.TypeOf([]byte{}))
		case DbSQLGoTypeFloat64:
			tv = reflect.New(reflect.TypeOf(float64(0)))
		case DbSQLGoTypeTime:
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

// DbTransaction 执行事物
func DbTransaction(ctx context.Context, db *sqlx.DB, f func(dbTx DbExeAble) error) error {
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
		log.Log.Debugf("exec sql:\n%s;\n%#v", query, args)
	}
}
