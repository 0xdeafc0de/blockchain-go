package blockchain

import (
	"fmt"
	"strings"
	"time"
)

type Transaction struct {
	Sender   string
	Receiver string
	Amount   int
	Data     string
}

// Validate ensures the transaction has the minimum required fields.
func (tx *Transaction) Validate() error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	if strings.TrimSpace(tx.Sender) == "" {
		return fmt.Errorf("transaction sender is required")
	}
	if strings.TrimSpace(tx.Receiver) == "" {
		return fmt.Errorf("transaction receiver is required")
	}
	if tx.Amount <= 0 {
		return fmt.Errorf("transaction amount must be greater than zero")
	}
	return nil
}

// ValidateTransactions checks a list of transactions before block creation.
func ValidateTransactions(transactions []*Transaction) error {
	for index, tx := range transactions {
		if err := tx.Validate(); err != nil {
			return fmt.Errorf("transaction %d invalid: %w", index, err)
		}
	}
	return nil
}

// NewBlock creates and returns a new block using provided transactions and previous block hash.
func NewBlock(transactions []*Transaction, prevHash []byte, height int) (*Block, error) {
	if err := ValidateTransactions(transactions); err != nil {
		return nil, err
	}

	block := &Block{
		Height:         height,
		Timestamp:      time.Now().Unix(),
		PrevBlockHash:  prevHash,
		MerkleRootHash: CalculateMerkleRoot(transactions),
		Bits:           targetBits,
		Version:        1,
		Reward:         50,
		Transactions:   transactions,
	}
	pow := NewProofOfWork(block)

	//do the mining work i.e. finding a nonce that gives a hash satisfying the target difficulty
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash

	// Now validate the mined block to verify that block’s nonce and hash meet the difficulty requirement
	if pow.Validate() {
		fmt.Println("Proof of Work: VALID")
	} else {
		fmt.Println("Proof of Work: INVALID")
	}

	return block, nil
}

// NewGenesisBlock returns the first block of the blockchain, also known as the genesis block.
func NewGenesisBlock() *Block {
	nb, err := NewBlock([]*Transaction{
		{Sender: "genesis", Receiver: "satoshi", Amount: 100, Data: "Genesis Block"},
	}, []byte{}, 0)
	if err != nil {
		panic(err)
	}
	return nb
}
