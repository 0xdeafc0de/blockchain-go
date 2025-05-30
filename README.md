# Simple Blockchain in Go
A minimalist blockchain implementation in Go demonstrating:

- Block headers with metadata
- Transaction list in the body
- Merkle root calculation
- Block linking via hashes

## Features

- Genesis block creation
- Block chaining
- SHA-256 hashing
- Merkle root hash of transactions
- Modular block and transaction structures

## Proof of Work
Each block includes a nonce computed via Proof-of-Work. The PoW difficulty is determined by `targetBits`.

## Run

```bash
go run blockchain.go


