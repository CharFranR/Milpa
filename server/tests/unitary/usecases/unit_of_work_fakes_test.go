package usecases_test

import (
	"context"

	port "milpa/domain/port/secondary"
)

// fakeUnitOfWork runs the unit of work against a hand-built scope.
//
// It deliberately does NOT simulate a rollback: a fake cannot prove atomicity,
// and pretending to would be a lie the compiler cannot catch. The atomicity
// claims are proven by the integration suite in tests/integration, which runs
// against a real PostgreSQL and a real rollback.
type fakeUnitOfWork struct {
	scope       port.TxScope
	withinTxErr error

	commits   int
	rollbacks int
}

func newFakeUnitOfWork(scope port.TxScope) *fakeUnitOfWork {
	return &fakeUnitOfWork{scope: scope}
}

func (f *fakeUnitOfWork) WithinTx(ctx context.Context, fn func(port.TxScope) error) error {
	if f.withinTxErr != nil {
		return f.withinTxErr
	}
	if err := fn(f.scope); err != nil {
		f.rollbacks++
		return err
	}
	f.commits++
	return nil
}

var _ port.UnitOfWork = (*fakeUnitOfWork)(nil)
