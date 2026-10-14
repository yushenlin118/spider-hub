package main

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

// 🗄️ Store 资料库存取
type Store struct {
	pool *pgxpool.Pool
}

// NewStore 连线资料库并建立资料表
func NewStore(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("❌ 资料库连线设定错误: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("❌ 无法连线资料库: %w", err)
	}
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		pool.Close()
		return nil, fmt.Errorf("❌ 建立资料表失败: %w", err)
	}
	return &Store{pool: pool}, nil
}

// Close 关闭连线池
func (s *Store) Close() { s.pool.Close() }

// 📋 lastRun 某个任务最近一次执行纪录
type lastRun struct {
	StartedAt  time.Time
	Status     string
	LoginError bool
}

// LastRun 取得任务最近一次执行纪录，从未执行过时回传 nil
func (s *Store) LastRun(ctx context.Context, job string) (*lastRun, error) {
	var r lastRun
	err := s.pool.QueryRow(ctx,
		`SELECT started_at, status, login_error FROM public.coinglass_collector_run
		 WHERE job = $1 ORDER BY started_at DESC LIMIT 1`, job).
		Scan(&r.StartedAt, &r.Status, &r.LoginError)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("❌ 查询执行纪录失败: %w", err)
	}
	return &r, nil
}

// RecordRun 写入执行纪录
func (s *Store) RecordRun(ctx context.Context, job string, started time.Time, rows int, runErr error, loginErr bool) error {
	status, msg := "ok", (*string)(nil)
	if runErr != nil {
		status = "error"
		m := runErr.Error()
		msg = &m
	}
	_, err := s.pool.Exec(ctx,
		`INSERT INTO public.coinglass_collector_run (job, started_at, finished_at, status, rows, error, login_error)
		 VALUES ($1, $2, now(), $3, $4, $5, $6)`,
		job, started, status, rows, msg, loginErr)
	if err != nil {
		return fmt.Errorf("❌ 写入执行纪录失败: %w", err)
	}
	return nil
}

// LatestCoinFlow 某条流入流出序列最新一根K线的时间，无资料时回传 nil（需要回补全量历史）
func (s *Store) LatestCoinFlow(ctx context.Context, market, symbol, interval string) (*time.Time, error) {
	var latest *time.Time
	err := s.pool.QueryRow(ctx,
		`SELECT max(bar_time) FROM public.coinglass_coin_flow WHERE market = $1 AND symbol = $2 AND interval = $3`,
		market, symbol, interval).Scan(&latest)
	return latest, err
}

// HasRows 资料表是否已有资料
// 📝 table 不含 coinglass_ 前缀，如 fear_greed_daily
// ⚠️ table 只能传程序内的常数，不能来自外部输入
func (s *Store) HasRows(ctx context.Context, table string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM public.coinglass_`+table+`)`).Scan(&exists)
	return exists, err
}

// 📦 execBatch 在单一交易中执行一批写入，回传笔数
func (s *Store) execBatch(ctx context.Context, b *pgx.Batch) (int, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("❌ 开启交易失败: %w", err)
	}
	defer tx.Rollback(ctx)

	n := b.Len()
	if err := tx.SendBatch(ctx, b).Close(); err != nil {
		return 0, fmt.Errorf("❌ 批量写入失败: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("❌ 提交交易失败: %w", err)
	}
	return n, nil
}

// 💥 insertLiqSnapshot 写入一笔清算地图快照与其档位
func (s *Store) insertLiqSnapshot(ctx context.Context, snap liqSnapshot) (int, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("❌ 开启交易失败: %w", err)
	}
	defer tx.Rollback(ctx)

	var id int64
	err = tx.QueryRow(ctx,
		`INSERT INTO public.coinglass_liq_map_snapshot (map_type, query_symbol, range_days, last_price, range_low, range_high)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		snap.MapType, snap.QuerySymbol, snap.RangeDays, snap.LastPrice, snap.RangeLow, snap.RangeHigh).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("❌ 写入清算地图快照失败: %w", err)
	}

	rows := make([][]any, len(snap.Levels))
	for i, l := range snap.Levels {
		rows[i] = []any{id, l.Exchange, l.InstrumentID, l.Bucket, l.Price, l.AmountUSD, l.Leverage, l.Tag}
	}
	n, err := tx.CopyFrom(ctx, pgx.Identifier{"public", "coinglass_liq_map_level"},
		[]string{"snapshot_id", "exchange", "instrument_id", "bucket", "price", "amount_usd", "leverage", "tag"},
		pgx.CopyFromRows(rows))
	if err != nil {
		return 0, fmt.Errorf("❌ 写入清算地图档位失败: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("❌ 提交交易失败: %w", err)
	}
	return int(n), nil
}
