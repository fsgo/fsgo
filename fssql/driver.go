/*
 * Copyright(C) 2021 github.com/hidu  All Rights Reserved.
 * Author: hidu (duv123+git@baidu.com)
 * Date: 2021/3/19
 */

package fssql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
)

const DriverName = "fssql"

func init() {
	sql.Register(DriverName, &fDriver{})
}

var dbs = map[string]DBOnlyCtx{}

func RegisterDB(name string, db DBOnlyCtx) {
	dbs[name] = db
}

type fDriver struct{}

func (a *fDriver) Open(name string) (driver.Conn, error) {
	db, has := dbs[name]
	if !has {
		return nil, fmt.Errorf("should RegisterDB first")
	}
	dri := &driverConn{
		raw: db,
	}
	return dri, nil
}

var _ driver.Driver = (*fDriver)(nil)

type driverStmt struct {
	raw Stmt
}

func (a *driverStmt) Close() error {
	return a.raw.Close()
}

func (a *driverStmt) NumInput() int {
	return -1
}

func (a *driverStmt) Exec(args []driver.Value) (driver.Result, error) {
	return a.raw.Exec(trans(args)...)
}

func (a *driverStmt) Query(args []driver.Value) (driver.Rows, error) {
	rows, err := a.raw.Query(trans(args)...)
	if err != nil {
		return nil, err
	}
	return &driverRows{raw: rows}, nil
}

var _ driver.Stmt = (*driverStmt)(nil)

func trans(args []driver.Value) []interface{} {
	tmp := make([]interface{}, len(args))
	for i, v := range args {
		tmp[i] = v
	}
	return tmp
}

type driverRows struct {
	raw Rows
}

func (d *driverRows) Columns() []string {
	cs, _ := d.raw.Columns()
	return cs
}

func (d *driverRows) Close() error {
	return d.raw.Close()
}

func (d *driverRows) Next(dest []driver.Value) error {
	tmp := trans(dest)
	return d.raw.Scan(tmp...)
}

var _ driver.Rows = (*driverRows)(nil)

type driverConn struct {
	raw DBOnlyCtx
}

func (d *driverConn) Prepare(query string) (driver.Stmt, error) {
	return d.PrepareContext(context.Background(), query)
}

func (d *driverConn) Begin() (driver.Tx, error) {
	return d.BeginTx(context.Background(), driver.TxOptions{})
}

func (d *driverConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	st, err := d.raw.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &driverStmt{
		raw: st,
	}, nil
}

func (d *driverConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	o := &sql.TxOptions{
		Isolation: sql.IsolationLevel(opts.Isolation),
		ReadOnly:  opts.ReadOnly,
	}
	return d.raw.BeginTx(ctx, o)
}

func (d *driverConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	vs := make([]interface{}, len(args))
	for i, v := range args {
		vs[i] = v
	}
	rows, err := d.raw.QueryContext(ctx, query, vs...)
	if err != nil {
		return nil, err
	}
	return &driverRows{
		raw: rows,
	}, nil
}

func (d *driverConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	vs := make([]interface{}, len(args))
	for i, v := range args {
		vs[i] = v
	}
	return d.raw.ExecContext(ctx, query, vs...)
}

func (d *driverConn) Close() error {
	return d.raw.Close()
}

var _ dConn = (*driverConn)(nil)

type dConn interface {
	driver.Conn
	driver.ConnPrepareContext
	driver.ConnBeginTx
	driver.ConnPrepareContext
	driver.QueryerContext
	driver.ExecerContext
}
