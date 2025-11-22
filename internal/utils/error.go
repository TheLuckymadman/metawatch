package utils

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func CheckRetriablePgError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure:
			return true
		}
	}
	var nerr net.Error
	return errors.As(err, &nerr)
}

func WithRetry[T any](ctx context.Context, f func() (T, error)) (T, error) {
	var dummy T
	delays := [3]int{1, 3, 5}
	v, err := f()

	if err != nil {
		if CheckRetriablePgError(err) {
			for i := 0; i < len(delays); i++ {
				select {
				case <-ctx.Done():
					return dummy, fmt.Errorf("context failed: %w", ctx.Err())
				case <-time.After(time.Duration(delays[i]) * time.Second):
					v, err = f()
					if err == nil {
						return v, nil
					}
					if !CheckRetriablePgError(err) {
						return dummy, fmt.Errorf("failed due to an unretriable error: %w", err)
					}
				}
			}
			return dummy, fmt.Errorf("max retry count %d exceeded: %w", len(delays), err)
		}
		return dummy, fmt.Errorf("failed due to an unretriable error: %w", err)
	}
	return v, nil
}
