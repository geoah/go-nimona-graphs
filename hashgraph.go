package hashgraph

import (
	net "github.com/nimona/go-nimona-net"
)

type Hashgraph struct {
	peer   net.Peer
	blocks *BlockStore
}

func (hg *Hashgraph) Get(blockID string) (*Block, error) {
	return hg.blocks.Get(blockID)
}

func (hg *Hashgraph) Add(block *Block) error {
	return hg.blocks.Add(block)
}

func (hg *Hashgraph) CreateGraph(acl BlockEventACL) (*Block, error) {
	// create block
	bl := &Block{
		Event: BlockEvent{
			ACL:    acl,
			Author: string(hg.peer.ID),
			Nonce:  RandStringBytesMaskImprSrc(10),
			Type:   EventTypeGraphCreate,
		},
	}
	bl.Hash = HashBlock(bl)

	// store block
	if err := hg.blocks.Add(bl); err != nil {
		return nil, err
	}

	return bl, nil
}

func (hg *Hashgraph) Append(threadID, eventType, data string) (*Block, error) {
	// find tip of thread
	parents, err := hg.blocks.FindTip(threadID)
	if err != nil {
		return nil, err
	}

	// create block
	bl := &Block{
		Event: BlockEvent{
			Root:    threadID,
			Author:  string(hg.peer.ID),
			Nonce:   RandStringBytesMaskImprSrc(10),
			Parents: parents,
			Type:    eventType,
			Data:    data,
		},
	}
	bl.Hash = HashBlock(bl)

	// store block
	if err := hg.blocks.Add(bl); err != nil {
		return nil, err
	}

	return bl, nil
}
