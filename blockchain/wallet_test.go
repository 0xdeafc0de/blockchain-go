package blockchain

import "testing"

func TestWalletAddressAndTransactionSignature(t *testing.T) {
	wallet, err := NewWallet()
	if err != nil {
		t.Fatalf("expected no error creating wallet: %v", err)
	}
	if wallet.Address() == "" {
		t.Fatal("expected wallet address to be non-empty")
	}

	transaction := NewTransaction(wallet.Address(), "receiver", 25, "signed transfer")
	if err := wallet.SignTransaction(transaction); err != nil {
		t.Fatalf("expected no error signing transaction: %v", err)
	}

	if !transaction.VerifySignature() {
		t.Fatal("expected signature verification to succeed")
	}

	transaction.Amount = 26
	if transaction.VerifySignature() {
		t.Fatal("expected signature verification to fail after tampering")
	}
}

func TestSignedTransactionValidation(t *testing.T) {
	wallet, err := NewWallet()
	if err != nil {
		t.Fatalf("expected no error creating wallet: %v", err)
	}

	transaction := NewTransaction(wallet.Address(), "receiver", 25, "signed transfer")
	if err := wallet.SignTransaction(transaction); err != nil {
		t.Fatalf("expected no error signing transaction: %v", err)
	}

	if err := transaction.Validate(); err != nil {
		t.Fatalf("expected signed transaction to validate, got %v", err)
	}
}
