package main

import (
	"fmt"
	"github.com/0xdeafc0de/blockchain-go/blockchain"
)

func main() {
	bc := blockchain.NewBlockchain()

	tx1 := blockchain.NewTransaction("Alice", "Bob", 10, "Alice pays Bob")
	if err := bc.AddBlock([]*blockchain.Transaction{tx1}); err != nil {
		panic(err)
	}

	tx2 := blockchain.NewTransaction("Bob", "Charlie", 5, "Bob pays Charlie")
	if err := bc.AddBlock([]*blockchain.Transaction{tx2}); err != nil {
		panic(err)
	}

	tx3 := blockchain.NewTransaction("Charlie", "Dave", 100, "Charlie pays Dave")
	if err := bc.AddBlock([]*blockchain.Transaction{tx3}); err != nil {
		panic(err)
	}

	tx4 := blockchain.NewTransaction("Alice", "Dave", 200, "Alice pays Dave")
	if err := bc.AddBlock([]*blockchain.Transaction{tx4}); err != nil {
		panic(err)
	}

	for _, block := range bc.Blocks() {
		fmt.Printf("Height: %d, Hash: %x, Prev: %x\n", block.Height, block.Hash, block.PrevBlockHash)
		for i, tx := range block.Transactions {
			fmt.Printf("Txn(%d) - Sender %s, Receiver %s, Amount %d - {%s}\n",
				i, tx.Sender, tx.Receiver, tx.Amount, tx.Data)
		}
		fmt.Println("-----")
	}
}
