package blockchain

import (
	"fmt"
	"time"
)

type Transaction struct {
	Sender   string
	Receiver string
	Amount   int
	Data     string
}

// NewBlock creates and returns a new block using provided transactions and previous block hash.
func NewBlock(transactions []*Transaction, prevHash []byte, height int) *Block {
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

	return block
}

// NewGenesisBlock returns the first block of the blockchain, also known as the genesis block.
func NewGenesisBlock() *Block {
	nb := NewBlock([]*Transaction{
		{Sender: "genesis", Receiver: "satoshi", Amount: 100, Data: "Genesis Block"},
	}, []byte{}, 0)
	return nb
}
