package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"
)

type Transaction struct {
	Sender   string
	Receiver string
	Amount   int
	Data     string
	PublicKey []byte
	Signature []byte
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
	if !tx.VerifySignature() {
		return fmt.Errorf("transaction signature is invalid")
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

// VerifySignature checks a signed transaction against its encoded payload.
// Unsigned transactions are treated as valid so existing flows remain compatible.
func (tx *Transaction) VerifySignature() bool {
	if tx == nil {
		return false
	}
	if len(tx.Signature) == 0 || len(tx.PublicKey) == 0 {
		return true
	}

	publicKeyX, publicKeyY := elliptic.Unmarshal(elliptic.P256(), tx.PublicKey)
	if publicKeyX == nil || publicKeyY == nil {
		return false
	}

	publicKey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     publicKeyX,
		Y:     publicKeyY,
	}
	hash := sha256.Sum256(tx.payload())
	return ecdsa.VerifyASN1(&publicKey, hash[:], tx.Signature)
}

// Sign attaches the wallet's public key and signature to the transaction.
func (tx *Transaction) Sign(wallet *Wallet) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	if wallet == nil || wallet.PrivateKey == nil {
		return fmt.Errorf("wallet is nil")
	}

	hash := sha256.Sum256(tx.payload())
	signature, err := ecdsa.SignASN1(rand.Reader, wallet.PrivateKey, hash[:])
	if err != nil {
		return fmt.Errorf("sign transaction: %w", err)
	}

	tx.PublicKey = elliptic.Marshal(wallet.PrivateKey.Curve, wallet.PrivateKey.PublicKey.X, wallet.PrivateKey.PublicKey.Y)
	tx.Signature = signature
	return nil
}

func (tx *Transaction) payload() []byte {
	type transactionPayload struct {
		Sender   string `json:"sender"`
		Receiver string `json:"receiver"`
		Amount   int    `json:"amount"`
		Data     string `json:"data"`
	}

	data, err := json.Marshal(transactionPayload{
		Sender:   tx.Sender,
		Receiver: tx.Receiver,
		Amount:   tx.Amount,
		Data:     tx.Data,
	})
	if err != nil {
		return []byte{}
	}
	return data
}

// Wallet stores an ECDSA private key and its derived public key.
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
}

// NewWallet creates a fresh wallet using the P-256 curve.
func NewWallet() (*Wallet, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate wallet key: %w", err)
	}
	return &Wallet{PrivateKey: privateKey}, nil
}

// Address returns the hex-like encoded public key bytes for the wallet.
func (w *Wallet) Address() string {
	if w == nil || w.PrivateKey == nil {
		return ""
	}
	return fmt.Sprintf("%x", elliptic.Marshal(w.PrivateKey.Curve, w.PrivateKey.PublicKey.X, w.PrivateKey.PublicKey.Y))
}

// SignTransaction signs a transaction using the wallet's private key.
func (w *Wallet) SignTransaction(tx *Transaction) error {
	return tx.Sign(w)
}

// PublicKeyBytes returns the wallet's public key encoded on the curve.
func (w *Wallet) PublicKeyBytes() []byte {
	if w == nil || w.PrivateKey == nil {
		return nil
	}
	return elliptic.Marshal(w.PrivateKey.Curve, w.PrivateKey.PublicKey.X, w.PrivateKey.PublicKey.Y)
}

// IsSameWallet checks whether the wallet's public key matches the provided key bytes.
func (w *Wallet) IsSameWallet(publicKey []byte) bool {
	if w == nil || w.PrivateKey == nil {
		return false
	}
	return big.NewInt(0).SetBytes(w.PublicKeyBytes()).Cmp(big.NewInt(0).SetBytes(publicKey)) == 0
}
