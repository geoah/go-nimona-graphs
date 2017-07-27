package hashgraph

import (
	"testing"

	mock "github.com/stretchr/testify/mock"
	suite "github.com/stretchr/testify/suite"

	mb "github.com/nimona/go-nimona-messagebus"
	net "github.com/nimona/go-nimona-net"
)

type HashgraphTestSuite struct {
	suite.Suite
	hashgraph    *Hashgraph
	blockstore   *BlockStore
	messagebus   *mb.MockMessageBus
	hashgraphbus *HashgraphBus
	handler      func(block *Block) error
	peerID       string
}

func (s *HashgraphTestSuite) SetupTest() {
	peer := net.Peer{
		ID: "peer-owner",
	}
	s.peerID = peer.ID
	s.messagebus = &mb.MockMessageBus{}
	s.blockstore = &BlockStore{
		blocks: map[string]*Block{},
	}
	s.hashgraph = &Hashgraph{
		peer:   peer,
		blocks: s.blockstore,
	}
	s.hashgraphbus = &HashgraphBus{
		peer:      peer,
		eventbus:  s.messagebus,
		hashgraph: s.hashgraph,
		handlers:  []func(bl *Block) error{},
	}
}

func (s *HashgraphTestSuite) TestCreateGraphSucceeds() {
	acl := BlockEventACL{}

	// reset blockstore
	for k := range s.blockstore.blocks {
		delete(s.blockstore.blocks, k)
	}

	// check that the blockstore is empty
	s.Len(s.blockstore.blocks, 0)

	// keep some things from the messagebus for future tests
	var payload *mb.Payload
	var block *Block

	// mock event bus send
	s.messagebus.On("Send", mock.AnythingOfType("*messagebus.Payload"), []string{}).Return(nil).
		Run(func(args mock.Arguments) {
			s.IsType(&mb.Payload{}, args[0])
			payload = args[0].(*mb.Payload)
			s.Equal(s.peerID, payload.Creator)
			block, _ = DecodeBlock(payload.Data)
			s.Equal(s.peerID, block.Event.Author)
			s.Len(block.Event.Parents, 0)
			s.Equal(EventTypeGraphCreate, block.Event.Type)
		})

	// create the graph
	oblock, err := s.hashgraphbus.CreateGraph(acl)
	s.Nil(err)
	s.NotNil(block)
	s.NotNil(oblock)
	s.Equal(oblock.Hash, block.Hash)

	// check that the blockstore has at least one block
	s.Len(s.blockstore.blocks, 1)
	s.Equal(s.peerID, s.blockstore.blocks[block.Hash].Event.Author)
}

func (s *HashgraphTestSuite) TestSubscribeGraphSucceeds() {
	acl := BlockEventACL{}

	// reset blockstore
	for k := range s.blockstore.blocks {
		delete(s.blockstore.blocks, k)
	}

	// check that the blockstore is empty
	s.Len(s.blockstore.blocks, 0)

	// keep some things from the messagebus for future tests
	var payload *mb.Payload
	var block *Block

	// mock event bus send
	s.messagebus.On("Send", mock.AnythingOfType("*messagebus.Payload"), []string{}).Return(nil).
		Run(func(args mock.Arguments) {
			s.IsType(&mb.Payload{}, args[0])
			payload = args[0].(*mb.Payload)
			s.Equal(s.peerID, payload.Creator)
			block, _ = DecodeBlock(payload.Data)
			s.Equal(s.peerID, block.Event.Author)
			s.Len(block.Event.Parents, 0)
			s.Equal(EventTypeGraphCreate, block.Event.Type)
		})

	// create the graph
	cblock, err := s.hashgraphbus.CreateGraph(acl)
	s.Nil(err)
	s.Equal(cblock.Hash, block.Hash)

	// clear previous mocks
	s.messagebus.Mock = mock.Mock{}

	// mock event bus send
	s.messagebus.On("Send", mock.AnythingOfType("*messagebus.Payload"), []string{}).Return(nil).
		Run(func(args mock.Arguments) {
			s.IsType(&mb.Payload{}, args[0])
			payload = args[0].(*mb.Payload)
			s.Equal(s.peerID, payload.Creator)
			block, _ = DecodeBlock(payload.Data)
			s.Equal(s.peerID, block.Event.Author)
			s.Len(block.Event.Parents, 1)
			s.Equal(EventTypeGraphSubscribe, block.Event.Type)
			s.Equal(s.peerID, block.Event.Author)
			s.Equal([]string{cblock.Hash}, block.Event.Parents)
			s.Equal(EventTypeGraphSubscribe, block.Event.Type)
		})

	// subscribe to the graph
	sblock, err := s.hashgraphbus.Subscribe(cblock.Hash)
	s.Nil(err)
	s.Equal(sblock.Hash, block.Hash)

	// check that the blockstore has at least one block
	s.Len(s.blockstore.blocks, 2)
	// and check something in the block just in case
	s.Equal(s.peerID, s.blockstore.blocks[block.Hash].Event.Author)
}

func TestHashgraphTestSuite(t *testing.T) {
	suite.Run(t, new(HashgraphTestSuite))
}
