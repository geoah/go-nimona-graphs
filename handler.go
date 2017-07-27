package hashgraph

import (
	mb "github.com/nimona/go-nimona-messagebus"
	net "github.com/nimona/go-nimona-net"
)

const (
	protocolID  = "/hashgraph/v1"
	messageType = protocolID + "/event"
)

type HashgraphBus struct {
	peer      net.Peer
	eventbus  mb.MessageBus
	hashgraph *Hashgraph
	handlers  []func(bl *Block) error
}

func New(network net.Network, peer net.Peer) (*HashgraphBus, error) {
	eb, err := mb.New(protocolID, network, peer)
	if err != nil {
		return nil, err
	}

	hb := &HashgraphBus{
		peer:     peer,
		handlers: []func(bl *Block) error{},
		eventbus: eb,
		hashgraph: &Hashgraph{
			peer: peer,
			blocks: &BlockStore{
				blocks: map[string]*Block{},
			},
		},
	}

	if err := eb.HandleMessage(hb.messageHander); err != nil {
		return nil, err
	}

	return hb, nil
}

func (hb *HashgraphBus) messageHander(hash []byte, message *mb.Message) error {
	// attempt to decode message payload
	block, err := DecodeBlock(message.Payload)
	if err != nil {
		return err
	}

	// try to add block to our hashgraph,	 return if already exists
	if err := hb.hashgraph.Add(block); err != nil {
		return err
	}

	// if it doesn't exist, check if we have all parents

	// if not, ignore block and ask subscribers for tips

	// finally trigger handlers
	for _, handler := range hb.handlers {
		handler(block) // TODO handle or log error
	}

	return nil
}

func (hb *HashgraphBus) HandleBlock(handler func(block *Block) error) error {
	hb.handlers = append(hb.handlers, handler)
	return nil
}

func (hb *HashgraphBus) CreateGraph(acl BlockEventACL) (*Block, error) {
	bl, err := hb.hashgraph.CreateGraph(acl)
	if err != nil {
		return nil, err
	}

	hb.sendBlock(bl)

	return bl, nil
}

func (hb *HashgraphBus) Append(threadID, data string) (*Block, error) {
	bl, err := hb.hashgraph.Append(threadID, EventTypeGraphAppend, data)
	if err != nil {
		return nil, err
	}

	hb.sendBlock(bl)

	return bl, nil
}

func (hb *HashgraphBus) Subscribe(threadID string) (*Block, error) {
	bl, err := hb.hashgraph.Append(threadID, EventTypeGraphSubscribe, "")
	if err != nil {
		return nil, err
	}

	hb.sendBlock(bl)

	return bl, nil
}

func (hb *HashgraphBus) sendBlock(bl *Block) (string, error) {
	// find recipients
	recipients, err := hb.hashgraph.blocks.FindSubscribers(bl.Hash)
	if err != nil {
		return "", err
	}

	// add people with read access
	// TODO this should be reading the whole graph, not just the first event
	if bl.Event.Root != "" {
		rbl, err := hb.hashgraph.blocks.Get(bl.Event.Root)
		if err != nil {
			// TODO Handle error
		} else {
			recipients = append(recipients, rbl.Event.ACL.Read...)
			recipients = append(recipients, rbl.Event.ACL.Write...)
		}
	}

	// TODO Remove dublicates from recipients

	// encode block
	// TODO Use message codec
	bs, _ := EncodeBlock(bl)

	// create a new message
	ev := &mb.Payload{
		Type:    messageType,
		Creator: hb.peer.ID,
		Data:    bs,
	}

	// sent block to recipiends
	if err := hb.eventbus.Send(ev, recipients); err != nil {
		return "", err
	}

	return bl.Hash, nil
}
