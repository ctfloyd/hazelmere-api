package database

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type txContextKey struct{}

// TransactionManager handles MongoDB transactions
type TransactionManager struct {
	client              *mongo.Client
	transactionsEnabled bool
}

func NewTransactionManager(client *mongo.Client, transactionsEnabled bool) *TransactionManager {
	return &TransactionManager{client: client, transactionsEnabled: transactionsEnabled}
}

// WithTransaction executes the given function within a transaction.
// If transactions are disabled, it simply executes the function directly.
// If a transaction is already in progress (detected via context), it reuses that transaction.
// Otherwise, it starts a new transaction.
// On error, the transaction is rolled back. On success, it is committed.
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if !tm.transactionsEnabled {
		return fn(ctx)
	}

	// Check if there's already a transaction in progress
	if tm.IsInTransaction(ctx) {
		// Reuse existing transaction - just run the function
		return fn(ctx)
	}

	// Start a new session and transaction
	session, err := tm.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(txCtx context.Context) (interface{}, error) {
		// Mark the context as being in a transaction
		txCtx = context.WithValue(txCtx, txContextKey{}, true)
		return nil, fn(txCtx)
	})

	return err
}

// IsInTransaction returns true if the context is currently within a transaction
func (tm *TransactionManager) IsInTransaction(ctx context.Context) bool {
	val := ctx.Value(txContextKey{})
	if val == nil {
		return false
	}
	inTx, ok := val.(bool)
	return ok && inTx
}
