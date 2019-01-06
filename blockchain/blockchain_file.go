package blockchain

import (
	"Peerster/fail"
	"Peerster/logger"
	"Peerster/messages"
	"Peerster/utils"
	"fmt"
	"sync"
	"time"
)

type BCF struct {
	forks         map[string]*FileBlock      // all forks of the blockchain (the head is on top of the longest fork)
	allBlocks     map[string]*FileBlock      // all blocks of the blockchain
	pendingBlocks map[string]*messages.Block // blocks with no parents (hence *Builder)
	ChainLength   int
	Head          *FileBlockBuilder // the block we will be mining over (not yet on the blockchain, hence *Builder)

	MineChan MineChan

	sync.RWMutex
}

func NewBCF() *BCF {
	return &BCF{
		forks:         map[string]*FileBlock{},
		allBlocks:     map[string]*FileBlock{},
		pendingBlocks: map[string]*messages.Block{},
		ChainLength:   0,
		Head:          NewFileBlockBuilder(nil),
		MineChan:      NewMineChan(true),
	}
}

func (bcf *BCF) AddTx(tx *Tx) bool {
	bcf.RLock()
	defer bcf.RUnlock()

	return bcf.Head.AddTxIfValid(tx)
}

func (bcf *BCF) GetHead() *FileBlockBuilder {
	bcf.RLock()
	defer bcf.RUnlock()

	return bcf.Head
}

func (bcf *BCF) AddBlock(block *messages.Block) [][32]byte {
	bcf.Lock()
	defer bcf.Unlock()

	bcf.addBlock(block)
	return bcf.addPendingBlocks() // could be done only when addBlock is successful and returning the whole list otherwise
}

func (bcf *BCF) MineOnce() bool {
	bcf.Lock()
	defer bcf.Unlock()

	nonce := utils.Random32Bytes()
	bcf.Head.SetNonce(nonce)
	fb, err := bcf.Head.Build()
	if err == nil {
		logger.Printlnf("FOUND-BLOCK %s", utils.HashToHex(fb.Hash[:])) //hw03 print
		if bcf.addFileBlock(fb) {
			bcf.MineChan.Push(fb)
			return true
		} else {
			fail.HandleError(fmt.Errorf("block mined not added to chain, this should not happen"))
		}
	}
	return false
}

func (bcf *BCF) GetBlock(hash [32]byte) *messages.BlockPublish {
	hashString := utils.HashToHex(hash[:])
	if fb, ok := bcf.allBlocks[hashString]; ok {
		return fb.ToBlockPublish(32)
	}
	return nil
}

// public functions without locks

func (bcf *BCF) MiningRoutine() {
	for {
		if len(bcf.Head.Transactions) > 0 {
			// only mine if new transactions
			bcf.MineOnce()
		} else {
			// allows cpu not to be overused when no transactions
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// private functions without locks

func (bcf *BCF) addBlock(block *messages.Block) bool {
	previousId := utils.HashToHex(block.PrevHash[:])
	var previousBlock *FileBlock
	if bcf.ChainLength == 0 {
		// first block, welcome and be our master! (previous set to nil)
		previousBlock = nil
	} else if forkBlock, ok := bcf.forks[previousId]; ok {
		// no new fork but one longer head (previous is a fork)
		previousBlock = forkBlock
	} else if singleBlock, ok := bcf.allBlocks[previousId]; ok {
		// new fork, cannot be longest head (previous is part of the chain)
		previousBlock = singleBlock
	} else if utils.AllZero(block.PrevHash[:]) {
		logger.Printlnf("forking from the genesis block")
		// new fork from the genesis block
		previousBlock = nil
	} else {
		bcf.pendingBlocks[block.HashString()] = block
		return false
	}
	delete(bcf.pendingBlocks, block.HashString())

	newFBB := NewFileBlockBuilder(previousBlock)

	if fb, err := newFBB.SetBlockAndBuild(block); err != nil {
		fail.HandleAbort("adding block failed when building", err)
		return false
	} else {
		return bcf.addFileBlock(fb)
	}
}

func (bcf *BCF) addPendingBlocks() [][32]byte {
	missingBlocks := [][32]byte{}
	for _, pBlock := range bcf.pendingBlocks {
		if bcf.addBlock(pBlock) {
			delete(bcf.pendingBlocks, pBlock.HashString())
			bcf.addPendingBlocks()
		} else {
			missingBlocks = append(missingBlocks, pBlock.PrevHash)
		}
	}
	return missingBlocks
}

func (bcf *BCF) addFileBlock(fb *FileBlock) bool {
	//logger.Printlnf("adding file block %s", fb.String())
	if fb.Previous == nil {
		bcf.allBlocks[fb.id] = fb
		bcf.forks[fb.id] = fb
		if bcf.ChainLength == 0 {
			logger.Printlnf(fb.ChainString()) // hw03 print
			bcf.ChainLength = fb.Length
			bcf.Head = NewFileBlockBuilder(fb)
		} else {
			_, hashString, _ := findMergure(fb, bcf.Head.Previous)
			logger.Printlnf("FORK-SHORTER %s", hashString)
		}
		return true
	} else if _, ok := bcf.forks[fb.Previous.id]; ok {
		bcf.allBlocks[fb.id] = fb
		delete(bcf.forks, fb.Previous.id)
		bcf.forks[fb.id] = fb

		if fb.Length > bcf.ChainLength {
			// even the longest fork now! changing head!
			// we need to keep the transactions that are not invalidated nor included in the new block

			newHead := NewFileBlockBuilder(fb)
			for _, tx := range bcf.Head.Transactions {
				newHead.AddTxIfValid(tx)
			}

			if rewind, _, rewindTransactions := findMergure(fb, bcf.Head.Previous); rewind > 0 {
				for _, tx := range rewindTransactions {
					newHead.AddTxIfValid(tx)
				}
				logger.Printlnf("FORK-LONGER rewind %d blocks", rewind)
			}
			logger.Printlnf(fb.ChainString()) // hw03 print
			bcf.ChainLength = fb.Length       //not new head which is 1 greater
			bcf.Head = newHead
		} else {
			_, hashString, _ := findMergure(fb, bcf.Head.Previous)
			logger.Printlnf("FORK-SHORTER %s", hashString)
		}
		return true
	} else if _, ok := bcf.allBlocks[fb.Previous.id]; ok {
		bcf.allBlocks[fb.id] = fb
		bcf.forks[fb.id] = fb
		_, hashString, _ := findMergure(fb, bcf.Head.Previous)
		logger.Printlnf("FORK-SHORTER %s", hashString)
		return true
	}
	fail.HandleError(fmt.Errorf("file-block comes out of nowhere"))
	return false
}

func findMergure(newBlock, oldBlock *FileBlock) (int, string, []*Tx) {
	rewind := 0
	rewindTransactions := []*Tx{}
	newChainBlocks := map[string]bool{}
	newChainBlock := newBlock
	for newChainBlock != nil {
		newChainBlocks[newChainBlock.id] = true
		newChainBlock = newChainBlock.Previous
	}

	oldChainBlock := oldBlock
	for oldChainBlock != nil {
		for _, tx := range oldChainBlock.Transactions {
			rewindTransactions = append(rewindTransactions, tx)
		}
		if _, ok := newChainBlocks[oldChainBlock.id]; ok {
			break
		}
		rewind += 1
		oldChainBlock = oldChainBlock.Previous
	}
	if oldChainBlock == nil {
		genesisHash := [32]byte{}
		genesisHashString := utils.HashToHex(genesisHash[:])
		return rewind, genesisHashString, rewindTransactions
	}
	return rewind, oldChainBlock.id, rewindTransactions
}
