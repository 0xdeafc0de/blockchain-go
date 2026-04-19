package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

type Node struct {
	mu      sync.RWMutex
	chain   *Blockchain
	peers   map[string]struct{}
	client  *http.Client
}

type chainPayload struct {
	Blocks []*Block `json:"blocks"`
}

type peerPayload struct {
	Peers []string `json:"peers"`
}

// NewNode creates a network node backed by the provided blockchain.
func NewNode(chain *Blockchain) *Node {
	if chain == nil {
		chain = NewBlockchain()
	}

	return &Node{
		chain:  chain,
		peers:  map[string]struct{}{},
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// Blockchain returns the node's local chain.
func (n *Node) Blockchain() *Blockchain {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.chain
}

// AddPeer registers a peer URL for later sync and propagation.
func (n *Node) AddPeer(peer string) error {
	normalized, err := normalizePeerURL(peer)
	if err != nil {
		return err
	}

	n.mu.Lock()
	defer n.mu.Unlock()
	n.peers[normalized] = struct{}{}
	return nil
}

// Peers returns the known peers in a stable order.
func (n *Node) Peers() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	peers := make([]string, 0, len(n.peers))
	for peer := range n.peers {
		peers = append(peers, peer)
	}
	sort.Strings(peers)
	return peers
}

// Handler exposes the HTTP endpoints for chain sync and propagation.
func (n *Node) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/chain", n.handleChain)
	mux.HandleFunc("/blocks", n.handleBlocks)
	mux.HandleFunc("/peers", n.handlePeers)
	return mux
}

// SyncFromPeers pulls the longest valid chain from known peers.
func (n *Node) SyncFromPeers(ctx context.Context) (bool, error) {
	bestChain := n.snapshotChain()
	updated := false

	for _, peer := range n.Peers() {
		candidate, err := n.fetchChain(ctx, peer)
		if err != nil {
			continue
		}
		if len(candidate.blocks) <= len(bestChain.blocks) {
			continue
		}
		if err := candidate.ValidateChain(); err != nil {
			continue
		}
		bestChain = candidate
		updated = true
	}

	if updated {
		n.mu.Lock()
		n.chain = bestChain
		n.mu.Unlock()
	}

	return updated, nil
}

// DiscoverPeers asks known peers for their peer lists and merges the results.
func (n *Node) DiscoverPeers(ctx context.Context) error {
	for _, peer := range n.Peers() {
		peers, err := n.fetchPeers(ctx, peer)
		if err != nil {
			continue
		}
		for _, discovered := range peers {
			if err := n.AddPeer(discovered); err != nil {
				continue
			}
		}
	}
	return nil
}

// BroadcastLatestBlock sends the current tip to every known peer.
func (n *Node) BroadcastLatestBlock(ctx context.Context) error {
	chain := n.snapshotChain()
	if len(chain.blocks) == 0 {
		return fmt.Errorf("blockchain is empty")
	}
	if len(n.Peers()) == 0 {
		return fmt.Errorf("no peers configured")
	}

	latest := chain.blocks[len(chain.blocks)-1]
	var lastErr error
	for _, peer := range n.Peers() {
		if err := n.postBlock(ctx, peer, latest); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (n *Node) handleChain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	chain := n.snapshotChain()
	writeJSON(w, http.StatusOK, chainPayload{Blocks: chain.blocks})
}

func (n *Node) handlePeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, peerPayload{Peers: n.Peers()})
}

func (n *Node) handleBlocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var block Block
	if err := json.NewDecoder(r.Body).Decode(&block); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode block: %w", err))
		return
	}

	if err := n.AcceptBlock(&block); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// AcceptBlock attempts to append a propagated block to the local chain.
func (n *Node) AcceptBlock(block *Block) error {
	if block == nil {
		return fmt.Errorf("block is nil")
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if len(n.chain.blocks) == 0 {
		return fmt.Errorf("local chain is empty")
	}

	tip := n.chain.blocks[len(n.chain.blocks)-1]
	if block.Height != tip.Height+1 {
		return fmt.Errorf("incoming block height %d does not extend tip %d", block.Height, tip.Height)
	}
	if string(block.PrevBlockHash) != string(tip.Hash) {
		return fmt.Errorf("incoming block does not link to local tip")
	}

	candidateBlocks := append(append([]*Block{}, n.chain.blocks...), block)
	candidate := &Blockchain{blocks: candidateBlocks}
	if err := candidate.ValidateChain(); err != nil {
		return err
	}

	n.chain = candidate
	return nil
}

func (n *Node) snapshotChain() *Blockchain {
	n.mu.RLock()
	defer n.mu.RUnlock()

	blocks := append([]*Block{}, n.chain.blocks...)
	return &Blockchain{blocks: blocks}
}

func (n *Node) fetchChain(ctx context.Context, peer string) (*Blockchain, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, peer+"/chain", nil)
	if err != nil {
		return nil, err
	}

	response, err := n.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("peer chain request returned %s", response.Status)
	}

	var payload chainPayload
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode peer chain: %w", err)
	}

	return &Blockchain{blocks: payload.Blocks}, nil
}

func (n *Node) fetchPeers(ctx context.Context, peer string) ([]string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, peer+"/peers", nil)
	if err != nil {
		return nil, err
	}

	response, err := n.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("peer list request returned %s", response.Status)
	}

	var payload peerPayload
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode peer list: %w", err)
	}

	return payload.Peers, nil
}

func (n *Node) postBlock(ctx context.Context, peer string, block *Block) error {
	payload, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("marshal block: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, peer+"/blocks", strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := n.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("peer block post returned %s", response.Status)
	}

	return nil
}

func normalizePeerURL(peer string) (string, error) {
	trimmed := strings.TrimSpace(peer)
	if trimmed == "" {
		return "", fmt.Errorf("peer URL is empty")
	}
	if !strings.Contains(trimmed, "://") {
		trimmed = "http://" + trimmed
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", err
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/"), nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
