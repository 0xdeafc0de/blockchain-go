package main

import (
	//	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	//	"strconv"
	"time"
)

type Transaction struct {
	Sender    string
	Receiver  string
	Amount    float64
	Timestamp int64
}

type BlockHeader struct {
	Version       int
	Height        int
	PrevBlockHash []byte
	Timestamp     int64
	MerkleRoot    []byte
	Nonce         int
	Bits          int
	BlockReward   float64
}

type BlockBody struct {
	Transactions []*Transaction
}

type Block struct {
	Header BlockHeader
	Body   BlockBody
	Hash   []byte
}

func calculateMerkleRoot(transactions []*Transaction) []byte {
	if len(transactions) == 0 {
		return []byte{}
	}

	var hashes [][]byte
	for _, tx := range transactions {
		txBytes, _ := json.Marshal(tx)
		hash := sha256.Sum256(txBytes)
		hashes = append(hashes, hash[:])
	}

	for len(hashes) > 1 {
		var newHashes [][]byte
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				combined := append(hashes[i], hashes[i+1]...)
				hash := sha256.Sum256(combined)
				newHashes = append(newHashes, hash[:])
			} else {
				newHashes = append(newHashes, hashes[i]) // odd node carried forward
			}
		}
		hashes = newHashes
	}
	return hashes[0]
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int) *Block {
	timestamp := time.Now().Unix()
	merkleRoot := calculateMerkleRoot(transactions)

	header := BlockHeader{
		Version:       1,
		Height:        height,
		PrevBlockHash: prevBlockHash,
		Timestamp:     timestamp,
		MerkleRoot:    merkleRoot,
		Nonce:         0,
		Bits:          1,
		BlockReward:   6.25,
	}

	body := BlockBody{Transactions: transactions}

	block := &Block{
		Header: header,
		Body:   body,
		Hash:   []byte{},
	}

	block.Hash = block.calculateHash()
	return block
}

func (b *Block) calculateHash() []byte {
	headerBytes, _ := json.Marshal(b.Header)
	hash := sha256.Sum256(headerBytes)
	return hash[:]
}

type Blockchain struct {
	blocks []*Block
}

func (bc *Blockchain) AddBlock(transactions []*Transaction) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(transactions, prevBlock.Hash, len(bc.blocks))
	bc.blocks = append(bc.blocks, newBlock)
}

func NewGenesisBlock() *Block {
	genesisTx := &Transaction{
		Sender:    "genesis",
		Receiver:  "miner",
		Amount:    50,
		Timestamp: time.Now().Unix(),
	}
	return NewBlock([]*Transaction{genesisTx}, []byte{}, 0)
}

func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewGenesisBlock()}}
}

func main() {
	bc := NewBlockchain()

	// Add sample blocks
	bc.AddBlock([]*Transaction{
		{Sender: "Alice", Receiver: "Bob", Amount: 10, Timestamp: time.Now().Unix()},
		{Sender: "Bob", Receiver: "Charlie", Amount: 5, Timestamp: time.Now().Unix()},
	})

	bc.AddBlock([]*Transaction{
		{Sender: "Charlie", Receiver: "Dave", Amount: 3, Timestamp: time.Now().Unix()},
	})

	for i, block := range bc.blocks {
		fmt.Printf("\nBlock #%d\n", i)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Prev: %x\n", block.Header.PrevBlockHash)
		fmt.Printf("MerkleRoot: %x\n", block.Header.MerkleRoot)
		fmt.Printf("Transactions:\n")
		for _, tx := range block.Body.Transactions {
			fmt.Printf("  From: %s To: %s Amount: %.2f\n", tx.Sender, tx.Receiver, tx.Amount)
		}
	}
}
