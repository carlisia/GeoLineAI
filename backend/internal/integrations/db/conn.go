package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"runtime"
	"strings"

	_ "modernc.org/sqlite"
)

// DB is one SQLite file reached through two pools. Every write goes through write,
// which holds a single connection, since SQLite admits one writer at a time; reads
// go through read, opened read-only so a write sent there fails instead of queueing,
// and sized for the concurrent reads WAL mode allows beside the writer.
type DB struct {
	read  *sql.DB
	write *sql.DB
}

// Close closes both pools.
func (d DB) Close() error {
	var errs []error
	for _, p := range []*sql.DB{d.read, d.write} {
		if p != nil {
			errs = append(errs, p.Close())
		}
	}
	return errors.Join(errs...)
}

// busyTimeout is how long, in milliseconds, a connection waits for a lock another
// holds before failing.
const busyTimeout = 5000

// connect opens both pools on the file at path. Connection settings travel in the
// DSN, because the driver applies a DSN's pragmas to every connection it opens, while
// a PRAGMA statement run on a pool would reach only one of its connections.
func connect(ctx context.Context, path string) (DB, error) {
	if path == "" || path == ":memory:" || strings.HasPrefix(path, "file:") {
		return DB{}, fmt.Errorf("database path %q must name a file", path)
	}
	// An immediate transaction takes the write lock when it begins, so a save that
	// reads before it writes waits its turn rather than failing halfway.
	write, err := pool(ctx, path, url.Values{
		"_txlock": {"immediate"},
		"_pragma": {fmt.Sprintf("busy_timeout(%d)", busyTimeout), "foreign_keys(1)", "journal_mode(WAL)"},
	}, 1)
	if err != nil {
		return DB{}, fmt.Errorf("open write pool: %w", err)
	}
	read, err := pool(ctx, path, url.Values{
		"mode":    {"ro"},
		"_pragma": {fmt.Sprintf("busy_timeout(%d)", busyTimeout)},
	}, max(4, runtime.GOMAXPROCS(0)))
	if err != nil {
		_ = write.Close()
		return DB{}, fmt.Errorf("open read pool: %w", err)
	}
	return DB{read: read, write: write}, nil
}

// pool opens and pings a pool of at most size connections, all kept idle between
// uses so a burst does not reopen them and rerun their pragmas.
func pool(ctx context.Context, path string, params url.Values, size int) (*sql.DB, error) {
	// The path is a file name, not a URI: escape what SQLite's URI parser would read
	// as an escape, a query or a fragment, so it opens exactly the file named.
	name := strings.NewReplacer("%", "%25", "?", "%3F", "#", "%23").Replace(path)
	p, err := sql.Open("sqlite", "file:"+name+"?"+params.Encode())
	if err != nil {
		return nil, err
	}
	p.SetMaxOpenConns(size)
	p.SetMaxIdleConns(size)
	if err := p.PingContext(ctx); err != nil {
		_ = p.Close()
		return nil, err
	}
	return p, nil
}
