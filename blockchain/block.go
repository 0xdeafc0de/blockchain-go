package blockchain

// Block represents a single block in the blockchain.
// It contains a header and a list of transactions.
type Block struct {
	Timestamp      int64 // Time the block was created
	Height         int
	PrevBlockHash  []byte // Hash of the previous block
	Hash           []byte // Hash of the current block
	MerkleRootHash []byte
	Nonce          int // The Nonce found via proof-of-work
	Bits           int
	Version        int
	Reward         int
	Transactions   []*Transaction // List of transaction in the block
}

// Blockchain is a series of validated Blocks linked by hashes.
type Blockchain struct {
	blocks []*Block
}

// NewBlockchain initializes a new blockchain with a genesis block.
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewGenesisBlock()}}
}

// AddBlock adds a block to the blockchain using the provided transaction data.
func (bc *Blockchain) AddBlock(transactions []*Transaction) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(transactions, prevBlock.Hash, prevBlock.Height+1)
	bc.blocks = append(bc.blocks, newBlock)
}

// Blocks returns a copy of all blocks in the blockchain.
func (bc *Blockchain) Blocks() []*Block {
	return bc.blocks
}

// NewTransaction creates a new transaction with the given sender, receiver, amount, and optional data.
func NewTransaction(sender, receiver string, amount int, data ...string) *Transaction {
	txData := ""
	if len(data) > 0 {
		txData = data[0]
	} else {
		txData = sender + " pays " + receiver
	}

	return &Transaction{
		Sender:   sender,
		Receiver: receiver,
		Amount:   amount,
		Data:     txData,
	}
}
