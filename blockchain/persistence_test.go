package blockchain

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadBlockchain(t *testing.T) {
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

	if got, want := len(loaded.Blocks()), len(bc.Blocks()); got != want {
		t.Fatalf("expected %d blocks after load, got %d", want, got)
	}

	for index := range bc.Blocks() {
		original := bc.Blocks()[index]
		restored := loaded.Blocks()[index]

		if original.Height != restored.Height {
			t.Fatalf("block %d height mismatch: want %d got %d", index, original.Height, restored.Height)
		}
		if string(original.Hash) != string(restored.Hash) {
			t.Fatalf("block %d hash mismatch", index)
		}
		if string(original.PrevBlockHash) != string(restored.PrevBlockHash) {
			t.Fatalf("block %d previous hash mismatch", index)
		}
	}
}

func TestLoadBlockchainFromFileRejectsInvalidData(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "chain.json")

	if err := os.WriteFile(filePath, []byte("not-json"), 0o644); err != nil {
		t.Fatalf("expected no error writing temp file: %v", err)
	}

	if _, err := LoadBlockchainFromFile(filePath); err == nil {
		t.Fatal("expected load error, got nil")
	}
}
