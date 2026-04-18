package blockchain

import (
	"path/filepath"
	"testing"
)

func TestValidateChainAcceptsValidBlockchain(t *testing.T) {
	bc := NewBlockchain()
	if err := bc.AddBlock([]*Transaction{NewTransaction("Alice", "Bob", 10)}); err != nil {
		t.Fatalf("expected no error adding block: %v", err)
	}
	if err := bc.AddBlock([]*Transaction{NewTransaction("Bob", "Charlie", 5)}); err != nil {
		t.Fatalf("expected no error adding block: %v", err)
	}

	if err := bc.ValidateChain(); err != nil {
		t.Fatalf("expected valid chain, got %v", err)
	}
}

func TestValidateChainRejectsCorruptedLink(t *testing.T) {
	bc := NewBlockchain()
	if err := bc.AddBlock([]*Transaction{NewTransaction("Alice", "Bob", 10)}); err != nil {
		t.Fatalf("expected no error adding block: %v", err)
	}

	bc.blocks[1].PrevBlockHash = []byte("broken-link")

	if err := bc.ValidateChain(); err == nil {
		t.Fatal("expected chain validation error, got nil")
	}
}

func TestLoadBlockchainFromFileRejectsCorruptedChain(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "chain.json")

	bc := NewBlockchain()
	if err := bc.AddBlock([]*Transaction{NewTransaction("Alice", "Bob", 10)}); err != nil {
		t.Fatalf("expected no error adding block: %v", err)
	}
	if err := bc.SaveToFile(filePath); err != nil {
		t.Fatalf("expected no error saving blockchain: %v", err)
	}

	loaded, err := LoadBlockchainFromFile(filePath)
	if err != nil {
		t.Fatalf("expected no error loading blockchain: %v", err)
	}
	loaded.blocks[1].PrevBlockHash = []byte("corrupted")

	corruptedPath := filepath.Join(tmpDir, "corrupted-chain.json")
	if err := loaded.SaveToFile(corruptedPath); err != nil {
		t.Fatalf("expected no error saving corrupted blockchain: %v", err)
	}

	if _, err := LoadBlockchainFromFile(corruptedPath); err == nil {
		t.Fatal("expected validation error loading corrupted chain, got nil")
	}
}
