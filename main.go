package main

import (
	"fmt"
	"github.com/0xdeafc0de/blockchain-go/blockchain"
)

func main() {
	bc := blockchain.NewBlockchain()

	aliceWallet, err := blockchain.NewWallet()
	if err != nil {
		panic(err)
	}
	bobWallet, err := blockchain.NewWallet()
	if err != nil {
		panic(err)
	}

	tx1 := blockchain.NewTransaction(aliceWallet.Address(), bobWallet.Address(), 10, "Alice pays Bob")
	if err := aliceWallet.SignTransaction(tx1); err != nil {
		panic(err)
	}
	if err := bc.AddBlock([]*blockchain.Transaction{tx1}); err != nil {
		panic(err)
	}

	charlieWallet, err := blockchain.NewWallet()
	if err != nil {
		panic(err)
	}
	tx2 := blockchain.NewTransaction(bobWallet.Address(), charlieWallet.Address(), 5, "Bob pays Charlie")
	if err := bobWallet.SignTransaction(tx2); err != nil {
		panic(err)
	}
	if err := bc.AddBlock([]*blockchain.Transaction{tx2}); err != nil {
		panic(err)
	}

	daveWallet, err := blockchain.NewWallet()
	if err != nil {
		panic(err)
	}
	tx3 := blockchain.NewTransaction(charlieWallet.Address(), daveWallet.Address(), 100, "Charlie pays Dave")
	if err := charlieWallet.SignTransaction(tx3); err != nil {
		panic(err)
	}
	if err := bc.AddBlock([]*blockchain.Transaction{tx3}); err != nil {
		panic(err)
	}

	tx4 := blockchain.NewTransaction(aliceWallet.Address(), daveWallet.Address(), 200, "Alice pays Dave")
	if err := aliceWallet.SignTransaction(tx4); err != nil {
		panic(err)
	}
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
