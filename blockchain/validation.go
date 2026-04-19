package blockchain

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// ValidateChain checks the entire blockchain for structural and cryptographic integrity.
func (bc *Blockchain) ValidateChain() error {
	if bc == nil {
		return fmt.Errorf("blockchain is nil")
	}
	if len(bc.blocks) == 0 {
		return fmt.Errorf("blockchain has no blocks")
	}

	for index, block := range bc.blocks {
		if block == nil {
			return fmt.Errorf("block %d is nil", index)
		}
		if err := ValidateTransactions(block.Transactions); err != nil {
			return fmt.Errorf("block %d invalid transactions: %w", index, err)
		}

		expectedMerkleRoot := CalculateMerkleRoot(block.Transactions)
		if !bytes.Equal(block.MerkleRootHash, expectedMerkleRoot) {
			return fmt.Errorf("block %d merkle root mismatch", index)
		}

		pow := NewProofOfWork(block)
		if !pow.Validate() {
			return fmt.Errorf("block %d proof of work is invalid", index)
		}

		expectedHash := sha256.Sum256(pow.prepareData(block.Nonce))
		if !bytes.Equal(block.Hash, expectedHash[:]) {
			return fmt.Errorf("block %d hash mismatch", index)
		}

		if index == 0 {
			if block.Height != 0 {
				return fmt.Errorf("genesis block height must be zero")
			}
			if len(block.PrevBlockHash) != 0 {
				return fmt.Errorf("genesis block must not reference a previous hash")
			}
			continue
		}

		previousBlock := bc.blocks[index-1]
		if block.Height != previousBlock.Height+1 {
			return fmt.Errorf("block %d height mismatch", index)
		}
		if !bytes.Equal(block.PrevBlockHash, previousBlock.Hash) {
			return fmt.Errorf("block %d previous hash mismatch", index)
		}
	}

	return nil
}

// ReplaceChain swaps the current chain for a validated candidate chain.
func (bc *Blockchain) ReplaceChain(blocks []*Block) error {
	candidate := &Blockchain{blocks: blocks}
	if err := candidate.ValidateChain(); err != nil {
		return err
	}

	bc.blocks = blocks
	return nil
}
