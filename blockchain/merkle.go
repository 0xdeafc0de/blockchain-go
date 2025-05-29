package blockchain

import "crypto/sha256"

// CalculateMerkleRoot calculates a Merkle root hash for the given transactions.
func CalculateMerkleRoot(txs []*Transaction) []byte {
	if len(txs) == 0 {
		return []byte{}
	}
	var hashes [][]byte
	for _, tx := range txs {
		hash := sha256.Sum256([]byte(tx.Data))
		hashes = append(hashes, hash[:])
	}
	for len(hashes) > 1 {
		var newLevel [][]byte
		for i := 0; i < len(hashes); i += 2 {
			if i+1 == len(hashes) {
				newLevel = append(newLevel, hashes[i])
			} else {
				combined := append(hashes[i], hashes[i+1]...)
				hash := sha256.Sum256(combined)
				newLevel = append(newLevel, hash[:])
			}
		}
		hashes = newLevel
	}
	return hashes[0]
}
