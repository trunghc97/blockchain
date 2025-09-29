package mempool

import (
	"sync"

	"orderer-cluster/proto"
)

// Mempool manages pending transactions
type Mempool struct {
	transactions map[string]*proto.Transaction // tx_id -> transaction
	mutex        sync.RWMutex
}

// NewMempool creates a new transaction mempool
func NewMempool() *Mempool {
	return &Mempool{
		transactions: make(map[string]*proto.Transaction),
	}
}

// AddTransaction adds a transaction to the mempool
func (m *Mempool) AddTransaction(tx *proto.Transaction) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.transactions[tx.TransactionId]; exists {
		return nil // Transaction already exists
	}

	m.transactions[tx.TransactionId] = tx
	return nil
}

// RemoveTransaction removes a transaction from the mempool
func (m *Mempool) RemoveTransaction(txID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.transactions, txID)
}

// RemoveTransactions removes multiple transactions from the mempool
func (m *Mempool) RemoveTransactions(txs []*proto.Transaction) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, tx := range txs {
		delete(m.transactions, tx.TransactionId)
	}
}

// GetPendingTransactions returns all pending transactions
func (m *Mempool) GetPendingTransactions() []*proto.Transaction {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	txs := make([]*proto.Transaction, 0, len(m.transactions))
	for _, tx := range m.transactions {
		txs = append(txs, tx)
	}

	return txs
}

// HasPendingTransactions checks if there are pending transactions
func (m *Mempool) HasPendingTransactions() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return len(m.transactions) > 0
}

// GetTransactionCount returns the number of pending transactions
func (m *Mempool) GetTransactionCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return len(m.transactions)
}
