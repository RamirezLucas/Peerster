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
	forks       map[string]*FileBlock
	allBlocks   map[string]*FileBlock
	chainLength int
	head        *FileBlockBuilder // the block we will be mining over (not yet on the blockchain, hence *Builder)

	MineChan MineChan

	sync.RWMutex
}

func NewBCF() *BCF {
	return &BCF{
		forks:       map[string]*FileBlock{},
		allBlocks:   map[string]*FileBlock{},
		chainLength: 0,
		head:        NewFileBlockBuilder(nil),
		MineChan:    NewMineChan(true),
	}
}

func (bcf *BCF) AddTx(tx *Tx) {
	bcf.RLock()
	defer bcf.RUnlock()

	bcf.head.AddTxIfValid(tx)
}

func (bcf *BCF) GetHead() *FileBlockBuilder {
	bcf.RLock()
	defer bcf.RUnlock()

	return bcf.head
}

func (bcf *BCF) AddBlock(block *messages.Block) bool {
	bcf.Lock()
	defer bcf.Unlock()

	//genesisHash := [32]byte{}
	//genesisHashString := utils.HashToHex(genesisHash[:])

	previousId := utils.HashToHex(block.PrevHash[:])
	var previousBlock *FileBlock
	if bcf.chainLength == 0 {
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
		// same as above ?
		logger.Printlnf("block ignored (unknown reference from previous block)")
		return false
	}

	newFBB := NewFileBlockBuilder(previousBlock)

	if fb, err := newFBB.SetBlockAndBuild(block); err != nil {
		fail.HandleAbort("adding block failed when building", err)
		return false
	} else {
		ret := bcf.addFileBlock(fb)
		return ret
	}
}

func (bcf *BCF) MineOnce() bool {
	bcf.Lock()
	defer bcf.Unlock()

	nonce := utils.Random32Bytes()
	bcf.head.SetNonce(nonce)
	fb, err := bcf.head.Build()
	if err == nil {
		logger.Printlnf("FOUND-BLOCK %s", utils.HashToHex(fb.hash[:])) //hw03 print
		if fb.IsAfterGenesis() {
			time.Sleep(5 * time.Second)
			logger.Printlnf("5sec waited on %s", utils.HashToHex(fb.hash[:]))
		}
		if bcf.addFileBlock(fb) {
			bcf.MineChan.Push(fb)
			return true
		} else {
			fail.HandleError(fmt.Errorf("block mined not added to chain, this should not happen"))
		}
	}
	return false
}

// public functions without locks

func (bcf *BCF) MiningRoutine(group *sync.WaitGroup) {
	defer group.Done()
	for {
		if len(bcf.head.newTransactions) > 0 {
			// only mine if new transactions
			bcf.MineOnce()
		} else {
			// allows cpu not to be overused when no transactions
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// private functions without locks

func (bcf *BCF) addFileBlock(fb *FileBlock) bool {
	//logger.Printlnf("adding file block %s", fb.String())
	if fb.previous == nil {
		bcf.allBlocks[fb.id] = fb
		bcf.forks[fb.id] = fb
		if bcf.chainLength == 0 {
			logger.Printlnf(fb.ChainString()) // hw03 print
			bcf.chainLength = fb.length
			bcf.head = NewFileBlockBuilder(fb)
		} else {
			_, hashString, _ := findMergure(fb, bcf.head.previous)
			logger.Printlnf("FORK-SHORTER %s", hashString)
		}
		return true
	} else if _, ok := bcf.forks[fb.previous.id]; ok {
		bcf.allBlocks[fb.id] = fb
		delete(bcf.forks, fb.previous.id)
		bcf.forks[fb.id] = fb

		if fb.length > bcf.chainLength {
			// even the longest fork now! changing head!
			// we need to keep the transactions that are not invalidated nor included in the new block

			newHead := NewFileBlockBuilder(fb)
			for _, tx := range bcf.head.newTransactions {
				newHead.AddTxIfValid(tx)
			}

			if rewind, _, rewindTransactions := findMergure(fb, bcf.head.previous); rewind > 0 {
				for _, tx := range rewindTransactions {
					newHead.AddTxIfValid(tx)
				}
				logger.Printlnf("FORK-LONGER rewind %d blocks", rewind)
			}
			logger.Printlnf(fb.ChainString()) // hw03 print
			bcf.chainLength = fb.length       //not new head which is 1 greater
			bcf.head = newHead
		} else {
			_, hashString, _ := findMergure(fb, bcf.head.previous)
			logger.Printlnf("FORK-SHORTER %s", hashString)
		}
		return true
	} else if _, ok := bcf.allBlocks[fb.previous.id]; ok {
		bcf.allBlocks[fb.id] = fb
		bcf.forks[fb.id] = fb
		_, hashString, _ := findMergure(fb, bcf.head.previous)
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
		newChainBlock = newChainBlock.previous
	}

	oldChainBlock := oldBlock
	for oldChainBlock != nil {
		for _, tx := range oldChainBlock.transactions {
			rewindTransactions = append(rewindTransactions, tx)
		}
		if _, ok := newChainBlocks[oldChainBlock.id]; ok {
			break
		}
		rewind += 1
		oldChainBlock = oldChainBlock.previous
	}
	if oldChainBlock == nil {
		genesisHash := [32]byte{}
		genesisHashString := utils.HashToHex(genesisHash[:])
		return rewind, genesisHashString, rewindTransactions
	}
	return rewind, oldChainBlock.id, rewindTransactions
}
