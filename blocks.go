package hashgraph

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
)

const (
	EventTypeGraphSubscribe = "graph.subscribe"
	EventTypeGraphCreate    = "graph.create"
	EventTypeGraphAppend    = "graph.append"
)

// Block -
type Block struct {
	Event     BlockEvent `json:"event"`
	Signature string     `json:"signature"`
	Hash      string     `json:"hash,omitempty"`
}

type BlockEventACL struct {
	Read  []string `json:"read,omitempty"`
	Write []string `json:"write,omitempty"`
}

type BlockEvent struct {
	Root    string        `json:"root,omitempty"`
	ACL     BlockEventACL `json:"acl,omitempty"`
	Author  string        `json:"author"`
	Data    string        `json:"data"`
	Nonce   string        `json:"nonce"`
	Parents []string      `json:"parents,omitempty"`
	Type    string        `json:"type"`
}

func HashBlock(block *Block) string {
	bs, _ := json.Marshal(block)
	return fmt.Sprintf("%x", sha1.Sum(bs))
}

func EncodeBlock(block *Block) ([]byte, error) {
	return json.Marshal(block)
}

func DecodeBlock(bs []byte) (*Block, error) {
	bl := &Block{}
	if err := json.Unmarshal(bs, bl); err != nil {
		return nil, err
	}
	return bl, nil
}
