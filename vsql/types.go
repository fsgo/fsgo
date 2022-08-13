// Copyright(C) 2021 github.com/hidu  All Rights Reserved.
// Author: hidu (duv123+git@baidu.com)
// Date: 2021/3/19

package vsql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"reflect"
	"time"
)

// DB interface off sql.DB
type DB interface {
	Begin() (Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	Close() error
	Conn(ctx context.Context) (Conn, error)
	Driver() driver.Driver
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Ping() error
	PingContext(ctx context.Context) error
	Prepare(query string) (Stmt, error)
	PrepareContext(ctx context.Context, query string) (Stmt, error)
	Query(query string, args ...any) (Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRow(query string, args ...any) Row
	QueryRowContext(ctx context.Context, query string, args ...any) Row
	SetConnMaxIdleTime(d time.Duration)
	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	Stats() sql.DBStats
}

type DBOnlyCtx interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	PrepareContext(ctx context.Context, query string) (Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PingContext(ctx context.Context) error
	Close() error
}

type DBRawOnlyCtx interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PingContext(ctx context.Context) error
	Close() error
}

var _ DBRawOnlyCtx = (*sql.DB)(nil)

// Tx interface off sql.Tx
type Tx interface {
	Commit() error
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Prepare(query string) (Stmt, error)
	PrepareContext(ctx context.Context, query string) (Stmt, error)
	Query(query string, args ...any) (Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRow(query string, args ...any) Row
	QueryRowContext(ctx context.Context, query string, args ...any) Row
	Rollback() error
	Stmt(stmt Stmt) Stmt
	StmtContext(ctx context.Context, stmt Stmt) Stmt
}

// Stmt interface off sql.Stmt
type Stmt interface {
	Close() error
	Exec(args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, args ...any) (sql.Result, error)
	Query(args ...any) (Rows, error)
	QueryContext(ctx context.Context, args ...any) (Rows, error)
	QueryRow(args ...any) Row
	QueryRowContext(ctx context.Context, args ...any) Row
}

// Conn interface off sql.Conn
type Conn interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	Close() error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PingContext(ctx context.Context) error
	PrepareContext(ctx context.Context, query string) (Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) Row
	Raw(f func(driverConn any) error) (err error)
}

// Rows interface off sql.Rows
type Rows interface {
	Close() error
	ColumnTypes() ([]ColumnType, error)
	Columns() ([]string, error)
	Err() error
	Next() bool
	NextResultSet() bool
	Scan(dest ...any) error
}

// Row interface off sql.Row
type Row interface {
	Err() error
	Scan(dest ...any) error
}

var _ Row = (*sql.Row)(nil)

// ColumnType interface off sql.ColumnType
type ColumnType interface {
	DatabaseTypeName() string
	DecimalSize() (precision, scale int64, ok bool)
	Length() (length int64, ok bool)
	Name() string
	Nullable() (nullable, ok bool)
	ScanType() reflect.Type
}

var _ ColumnType = (*sql.ColumnType)(nil)

// NewDB 将 *sql.DB 转换为一个 interface DB
func NewDB(d *sql.DB) DB {
	return &fDB{
		raw: d,
	}
}

var _ DB = (*fDB)(nil)

type fDB struct {
	raw *sql.DB
}

func (d *fDB) Begin() (Tx, error) {
	tx, err := d.raw.Begin()
	if err != nil {
		return nil, err
	}
	return NewTx(tx), nil
}

func (d *fDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	tx, err := d.raw.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return NewTx(tx), nil
}

func (d *fDB) Close() error {
	return d.raw.Close()
}

func (d *fDB) Conn(ctx context.Context) (Conn, error) {
	conn, err := d.raw.Conn(ctx)
	if err != nil {
		return nil, err
	}
	return NewConn(conn), nil
}

func (d *fDB) Driver() driver.Driver {
	return d.raw.Driver()
}

func (d *fDB) Exec(query string, args ...any) (sql.Result, error) {
	return d.raw.Exec(query, args...)
}

func (d *fDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.raw.ExecContext(ctx, query, args...)
}

func (d *fDB) Ping() error {
	return d.raw.Ping()
}

func (d *fDB) PingContext(ctx context.Context) error {
	return d.raw.PingContext(ctx)
}

func (d *fDB) Prepare(query string) (Stmt, error) {
	st, err := d.raw.Prepare(query)
	if err != nil {
		return nil, err
	}
	return NewStmt(st), nil
}

func (d *fDB) PrepareContext(ctx context.Context, query string) (Stmt, error) {
	st, err := d.raw.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return NewStmt(st), nil
}

func (d *fDB) Query(query string, args ...any) (Rows, error) {
	rows, err := d.raw.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return NewRows(rows), nil
}

func (d *fDB) QueryContext(ctx context.Context, query string, args ...any) (Rows, error) {
	rows, err := d.raw.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return NewRows(rows), nil
}

func (d *fDB) QueryRow(query string, args ...any) Row {
	return d.raw.QueryRow(query, args...)
}

func (d *fDB) QueryRowContext(ctx context.Context, query string, args ...any) Row {
	return d.raw.QueryRowContext(ctx, query, args...)
}

func (d *fDB) SetConnMaxIdleTime(t time.Duration) {
	d.raw.SetConnMaxIdleTime(t)
}

func (d *fDB) SetConnMaxLifetime(t time.Duration) {
	d.raw.SetConnMaxLifetime(t)
}

func (d *fDB) SetMaxIdleConns(n int) {
	d.raw.SetMaxIdleConns(n)
}

func (d *fDB) SetMaxOpenConns(n int) {
	d.raw.SetMaxOpenConns(n)
}

func (d *fDB) Stats() sql.DBStats {
	return d.raw.Stats()
}

func (d *fDB) Raw() *sql.DB {
	return d.raw
}

func NewTx(raw *sql.Tx) Tx {
	return &fTx{
		raw: raw,
	}
}

var _ Tx = (*fTx)(nil)

type fTx struct {
	raw *sql.Tx
}

func (f *fTx) Commit() error {
	return f.raw.Commit()
}

func (f *fTx) Exec(query string, args ...any) (sql.Result, error) {
	return f.raw.Exec(query, args...)
}

func (f *fTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.raw.ExecContext(ctx, query, args...)
}

func (f *fTx) Prepare(query string) (Stmt, error) {
	st, err := f.raw.Prepare(query)
	if err != nil {
		return nil, err
	}
	return NewStmt(st), nil
}

func (f *fTx) PrepareContext(ctx context.Context, query string) (Stmt, error) {
	st, err := f.raw.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return NewStmt(st), nil
}

func (f *fTx) Query(query string, args ...any) (Rows, error) {
	rows, err := f.raw.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return NewRows(rows), nil
}

func (f *fTx) QueryContext(ctx context.Context, query string, args ...any) (Rows, error) {
	rows, err := f.raw.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return NewRows(rows), nil
}

func (f *fTx) QueryRow(query string, args ...any) Row {
	return f.raw.QueryRow(query, args...)
}

func (f *fTx) QueryRowContext(ctx context.Context, query string, args ...any) Row {
	return f.raw.QueryRowContext(ctx, query, args...)
}

func (f *fTx) Rollback() error {
	return f.raw.Rollback()
}

func (f *fTx) Stmt(stmt Stmt) Stmt {
	raw := stmt.(interface{ Raw() *sql.Stmt }).Raw()
	return NewStmt(f.raw.Stmt(raw))
}

func (f *fTx) StmtContext(ctx context.Context, stmt Stmt) Stmt {
	raw := stmt.(interface{ Raw() *sql.Stmt }).Raw()
	return NewStmt(f.raw.StmtContext(ctx, raw))
}

func (f *fTx) Raw() *sql.Tx {
	return f.raw
}

func NewStmt(raw *sql.Stmt) Stmt {
	return &fStmt{
		raw: raw,
	}
}

var _ Stmt = (*fStmt)(nil)

type fStmt struct {
	raw *sql.Stmt
}

func (f *fStmt) Close() error {
	return f.raw.Close()
}

func (f *fStmt) Exec(args ...any) (sql.Result, error) {
	return f.raw.Exec(args...)
}

func (f *fStmt) ExecContext(ctx context.Context, args ...any) (sql.Result, error) {
	return f.raw.ExecContext(ctx, args...)
}

func (f *fStmt) Query(args ...any) (Rows, error) {
	rows, err := f.raw.Query(args...)
	if err != nil {
		return nil, err
	}
	return NewRows(rows), nil
}

func (f *fStmt) QueryContext(ctx context.Context, args ...any) (Rows, error) {
	rows, err := f.raw.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	return NewRows(rows), nil
}

func (f *fStmt) QueryRow(args ...any) Row {
	return f.raw.QueryRow(args...)
}

func (f *fStmt) QueryRowContext(ctx context.Context, args ...any) Row {
	return f.raw.QueryRowContext(ctx, args...)
}

func (f *fStmt) Raw() *sql.Stmt {
	return f.raw
}

func NewRows(raw *sql.Rows) Rows {
	return &fRows{
		raw: raw,
	}
}

var _ Rows = (*fRows)(nil)

type fRows struct {
	raw *sql.Rows
}

func (f *fRows) Close() error {
	return f.raw.Close()
}

func (f *fRows) ColumnTypes() ([]ColumnType, error) {
	cs, err := f.raw.ColumnTypes()
	if err != nil {
		return nil, err
	}
	vs := make([]ColumnType, len(cs))
	for i, v := range cs {
		vs[i] = v
	}
	return vs, nil
}

func (f *fRows) Columns() ([]string, error) {
	return f.raw.Columns()
}

func (f *fRows) Err() error {
	return f.raw.Err()
}

func (f *fRows) Next() bool {
	return f.raw.Next()
}

func (f *fRows) NextResultSet() bool {
	return f.raw.NextResultSet()
}

func (f *fRows) Scan(dest ...any) error {
	return f.raw.Scan(dest...)
}

func NewConn(raw *sql.Conn) Conn {
	return &fConn{
		raw: raw,
	}
}

var _ Conn = (*fConn)(nil)

type fConn struct {
	raw *sql.Conn
}

func (f *fConn) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	tx, err := f.raw.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return NewTx(tx), nil
}

func (f *fConn) Close() error {
	return f.raw.Close()
}

func (f *fConn) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.raw.ExecContext(ctx, query, args...)
}

func (f *fConn) PingContext(ctx context.Context) error {
	return f.raw.PingContext(ctx)
}

func (f *fConn) PrepareContext(ctx context.Context, query string) (Stmt, error) {
	st, err := f.raw.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return NewStmt(st), nil
}

func (f *fConn) QueryContext(ctx context.Context, query string, args ...any) (Rows, error) {
	rows, err := f.raw.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return NewRows(rows), nil
}

func (f *fConn) QueryRowContext(ctx context.Context, query string, args ...any) Row {
	return f.raw.QueryRowContext(ctx, query, args...)
}

func (f *fConn) Raw(fn func(driverConn any) error) (err error) {
	return f.raw.Raw(fn)
}

func (f *fConn) RawConn() *sql.Conn {
	return f.raw
}

func NewDBOnlyCtx(raw DBRawOnlyCtx) DBOnlyCtx {
	return &fDBOnlyCtx{
		raw: raw,
	}
}

var _ DBOnlyCtx = (*fDBOnlyCtx)(nil)

type fDBOnlyCtx struct {
	raw DBRawOnlyCtx
}

func (f *fDBOnlyCtx) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	tx, err := f.raw.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return NewTx(tx), nil
}

func (f *fDBOnlyCtx) PrepareContext(ctx context.Context, query string) (Stmt, error) {
	st, err := f.raw.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return NewStmt(st), nil
}

func (f *fDBOnlyCtx) QueryContext(ctx context.Context, query string, args ...any) (Rows, error) {
	rows, err := f.raw.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return NewRows(rows), nil
}

func (f *fDBOnlyCtx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.raw.ExecContext(ctx, query, args...)
}

func (f *fDBOnlyCtx) PingContext(ctx context.Context) error {
	return f.raw.PingContext(ctx)
}

func (f *fDBOnlyCtx) Close() error {
	return f.raw.Close()
}
