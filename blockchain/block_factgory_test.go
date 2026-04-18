package blockchain

import "testing"

func TestTransactionValidate(t *testing.T) {
	testCases := []struct {
		name    string
		tx      *Transaction
		wantErr bool
	}{
		{
			name:    "valid transaction",
			tx:      NewTransaction("Alice", "Bob", 10, "pay Bob"),
			wantErr: false,
		},
		{
			name:    "missing sender",
			tx:      &Transaction{Receiver: "Bob", Amount: 10},
			wantErr: true,
		},
		{
			name:    "missing receiver",
			tx:      &Transaction{Sender: "Alice", Amount: 10},
			wantErr: true,
		},
		{
			name:    "non-positive amount",
			tx:      &Transaction{Sender: "Alice", Receiver: "Bob", Amount: 0},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.tx.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestValidateTransactions(t *testing.T) {
	t.Run("accepts valid transactions", func(t *testing.T) {
		txs := []*Transaction{
			NewTransaction("Alice", "Bob", 10),
			NewTransaction("Bob", "Charlie", 5),
		}

		if err := ValidateTransactions(txs); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("rejects invalid transactions", func(t *testing.T) {
		txs := []*Transaction{
			NewTransaction("Alice", "Bob", 10),
			{Sender: "", Receiver: "Charlie", Amount: 5},
		}

		if err := ValidateTransactions(txs); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestNewBlockRejectsInvalidTransactions(t *testing.T) {
	_, err := NewBlock([]*Transaction{{Sender: "Alice", Receiver: "Bob", Amount: 0}}, []byte{}, 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBlockchainAddBlock(t *testing.T) {
	bc := NewBlockchain()
	if len(bc.Blocks()) != 1 {
		t.Fatalf("expected genesis block, got %d blocks", len(bc.Blocks()))
	}

	if err := bc.AddBlock([]*Transaction{NewTransaction("Alice", "Bob", 10)}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(bc.Blocks()) != 2 {
		t.Fatalf("expected 2 blocks after append, got %d", len(bc.Blocks()))
	}
}
