package migrate

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/exc-works/migrate/internal/dialect"
)

func TestTransactionFailurePaths(t *testing.T) {
	for _, tc := range []struct {
		name                                string
		panicValue                          any
		stepErr, errorRollback, errorCommit error
	}{
		{name: "success"},
		{name: "execution", stepErr: errors.New("step failed")},
		{name: "rollback", stepErr: errors.New("step failed"), errorRollback: errors.New("rollback failed")},
		{name: "commit", errorCommit: errors.New("commit failed")},
		{name: "panic", panicValue: "private panic detail"},
		{name: "panic and rollback", panicValue: "private panic detail", errorRollback: errors.New("rollback failed")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			s := &Service{ctx: context.Background(), cfg: Config{DB: db, Dialect: dialect.PostgresDialect{}}}
			mock.ExpectBegin()
			if tc.stepErr != nil || tc.panicValue != nil {
				mock.ExpectRollback().WillReturnError(tc.errorRollback)
			} else {
				mock.ExpectCommit().WillReturnError(tc.errorCommit)
			}
			var escaped any
			func() {
				defer func() { escaped = recover() }()
				err = s.withTx(func(execContextRunner) error {
					if tc.panicValue != nil {
						panic(tc.panicValue)
					}
					return tc.stepErr
				})
			}()
			if escaped != nil {
				t.Errorf("panic escaped the execution boundary")
			}
			if tc.stepErr != nil && !errors.Is(err, tc.stepErr) {
				t.Errorf("lost execution error: %v", err)
			}
			if tc.errorRollback != nil && !errors.Is(err, tc.errorRollback) {
				t.Errorf("lost rollback error: %v", err)
			}
			if tc.errorCommit != nil && !errors.Is(err, tc.errorCommit) {
				t.Errorf("lost commit error: %v", err)
			}
			if tc.panicValue != nil && (err == nil || strings.Contains(err.Error(), "private panic detail")) {
				t.Errorf("panic must produce a sanitized error: %v", err)
			}
			if tc.stepErr == nil && tc.panicValue == nil && tc.errorCommit == nil && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestTransactionBeginFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	want := errors.New("begin failed")
	mock.ExpectBegin().WillReturnError(want)
	s := &Service{ctx: context.Background(), cfg: Config{DB: db, Dialect: dialect.PostgresDialect{}}}
	if err = s.withTx(func(execContextRunner) error { t.Fatal("callback ran after begin failure"); return nil }); !errors.Is(err, want) {
		t.Fatalf("%v", err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestNonTransactionalPanicIsSanitized(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := &Service{ctx: context.Background(), cfg: Config{DB: db, Dialect: dialect.ClickHouseDialect{}}}
	err = s.withTx(func(execContextRunner) error { panic("private panic detail") })
	if err == nil || strings.Contains(err.Error(), "private panic detail") {
		t.Fatalf("%v", err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
