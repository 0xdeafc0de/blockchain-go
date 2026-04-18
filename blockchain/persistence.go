package blockchain

import (
	"encoding/json"
	"fmt"
	"os"
)

type blockchainSnapshot struct {
	Blocks []*Block `json:"blocks"`
}

// SaveToFile writes the current blockchain state to disk in JSON format.
func (bc *Blockchain) SaveToFile(path string) error {
	if bc == nil {
		return fmt.Errorf("blockchain is nil")
	}

	snapshot := blockchainSnapshot{Blocks: bc.blocks}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal blockchain: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write blockchain file: %w", err)
	}

	return nil
}

// LoadBlockchainFromFile loads a blockchain from a JSON snapshot on disk.
func LoadBlockchainFromFile(path string) (*Blockchain, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read blockchain file: %w", err)
	}

	var snapshot blockchainSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("unmarshal blockchain: %w", err)
	}

	if len(snapshot.Blocks) == 0 {
		return nil, fmt.Errorf("blockchain file contains no blocks")
	}

	return &Blockchain{blocks: snapshot.Blocks}, nil
}
