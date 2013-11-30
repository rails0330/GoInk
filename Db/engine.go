package Db

import (
	"database/sql"
	"reflect"
	"fmt"
)

type Engine struct {
	SqlDb *sql.DB
	statements map[string]*sql.Stmt
}

func (this *Engine) Prepare(name string, sql string) error {
	stmt, e := this.SqlDb.Prepare(sql)
	if e != nil {
		return e
	}
	this.statements[name] = stmt
	return nil
}

func (this *Engine) Exec(sql string, arg...interface {}) (*Result, error) {
	stmt := this.statements[sql]
	if stmt != nil {
		rs, e := stmt.Exec(arg...)
		if e != nil {
			return nil, e
		}
		return newResultFromSqlResult(rs), nil
	}
	rs, e := this.SqlDb.Exec(sql, arg...)
	if e != nil {
		return nil, e
	}
	return newResultFromSqlResult(rs), nil
}

func (this *Engine) Query(sql string, arg...interface {}) (*Result, error) {
	stmt := this.statements[sql]
	if stmt != nil {
		rows, e := stmt.Query(arg...)
		if e != nil {
			return nil, e
		}
		return newResultFromRows(rows), nil
	}
	rows, e := this.SqlDb.Query(sql, arg...)
	if e != nil {
		return nil, e
	}
	return newResultFromRows(rows), nil
}

func NewEngine(driver, dsn string, conn...int) (*Engine, error) {
	db, e := sql.Open(driver, dsn)
	if e != nil {
		return nil, e
	}
	if len(conn) > 0 {
		db.SetMaxIdleConns(conn[0])
	}
	if len(conn) > 1 {
		db.SetMaxOpenConns(conn[1])
	}
	engine := new(Engine)
	engine.SqlDb = db
	engine.statements = make(map[string]*sql.Stmt)
	return engine, nil
}

// map row data to string
func rowsToMap(rows *sql.Rows) ([]map[string]string, error) {
	cols, e := rows.Columns()
	if e != nil {
		return nil, e
	}
	tmpItf := make([]interface{}, len(cols))
	for i, _ := range tmpItf {
		var itr interface{}
		tmpItf[i] = &itr
	}
	// set returning result
	rs := make([]map[string]string, 0)
	for rows.Next() {
		rows.Scan(tmpItf...)
		rowRst := make(map[string]string)
		// make all column values to string and append to result
		for i, col := range tmpItf {
			str := fmt.Sprint(reflect.Indirect(reflect.ValueOf(col)).Interface())
			if str == "<nil>" {
				str = ""
			}
			rowRst[cols[i]] = str
		}
		rs = append(rs, rowRst)
	}
	return rs, nil
}

