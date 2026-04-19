package blockchain

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNodeBroadcastAndAcceptBlock(t *testing.T) {
	seedChain := NewBlockchain()
	peerNode := NewNode(seedChain)
	peerServer := httptest.NewServer(peerNode.Handler())
	defer peerServer.Close()

	localChain := NewBlockchain()
	localNode := NewNode(localChain)
	if err := localNode.AddPeer(peerServer.URL); err != nil {
		t.Fatalf("expected no error adding peer: %v", err)
	}

	if err := localNode.Blockchain().AddBlock([]*Transaction{NewTransaction("Alice", "Bob", 10)}); err != nil {
		t.Fatalf("expected no error mining local block: %v", err)
	}

	if err := localNode.BroadcastLatestBlock(context.Background()); err != nil {
		t.Fatalf("expected no error broadcasting block: %v", err)
	}

	if got, want := len(peerNode.Blockchain().Blocks()), 2; got != want {
		t.Fatalf("expected peer chain length %d, got %d", want, got)
	}
}

func TestNodeSyncFromPeersUsesLongestValidChain(t *testing.T) {
	shortChain := NewBlockchain()
	shortNode := NewNode(shortChain)
	shortServer := httptest.NewServer(shortNode.Handler())
	defer shortServer.Close()

	longChain := NewBlockchain()
	if err := longChain.AddBlock([]*Transaction{NewTransaction("Alice", "Bob", 10)}); err != nil {
		t.Fatalf("expected no error mining first block: %v", err)
	}
	if err := longChain.AddBlock([]*Transaction{NewTransaction("Bob", "Charlie", 5)}); err != nil {
		t.Fatalf("expected no error mining second block: %v", err)
	}
	longNode := NewNode(longChain)
	longServer := httptest.NewServer(longNode.Handler())
	defer longServer.Close()

	if err := shortNode.AddPeer(longServer.URL); err != nil {
		t.Fatalf("expected no error adding long peer: %v", err)
	}
	if err := shortNode.AddPeer(shortServer.URL); err != nil {
		t.Fatalf("expected no error adding short peer: %v", err)
	}

	updated, err := shortNode.SyncFromPeers(context.Background())
	if err != nil {
		t.Fatalf("expected no sync error, got %v", err)
	}
	if !updated {
		t.Fatal("expected sync to update the local chain")
	}
	if got, want := len(shortNode.Blockchain().Blocks()), 3; got != want {
		t.Fatalf("expected synced chain length %d, got %d", want, got)
	}
}

func TestNodeDiscoverPeers(t *testing.T) {
	seedNode := NewNode(NewBlockchain())
	seedNode.AddPeer("http://example-one.local:8080")
	seedNode.AddPeer("http://example-two.local:8080")
	seedServer := httptest.NewServer(seedNode.Handler())
	defer seedServer.Close()

	observer := NewNode(NewBlockchain())
	if err := observer.AddPeer(seedServer.URL); err != nil {
		t.Fatalf("expected no error adding seed peer: %v", err)
	}

	if err := observer.DiscoverPeers(context.Background()); err != nil {
		t.Fatalf("expected no discovery error, got %v", err)
	}

	peers := observer.Peers()
	if len(peers) != 3 {
		t.Fatalf("expected 3 peers after discovery, got %d", len(peers))
	}
}

func TestHandleChainReturnsJSON(t *testing.T) {
	node := NewNode(NewBlockchain())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/chain", nil)

	node.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var payload chainPayload
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("expected JSON chain payload, got %v", err)
	}
	if len(payload.Blocks) != 1 {
		t.Fatalf("expected genesis chain in payload, got %d blocks", len(payload.Blocks))
	}
}

func TestAddPeerRejectsEmptyURL(t *testing.T) {
	node := NewNode(NewBlockchain())
	if err := node.AddPeer("   "); err == nil {
		t.Fatal("expected error for empty peer URL")
	}
}

func TestBroadcastLatestBlockWithoutPeers(t *testing.T) {
	node := NewNode(NewBlockchain())
	if err := node.BroadcastLatestBlock(context.Background()); err == nil {
		t.Fatal("expected broadcast error without peers")
	}
}
