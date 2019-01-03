package blockchain

import "Peerster/updates"

type T = *FileBlock

type MineChan interface {
	Get() T
	Push(file T)
}

type matchChan struct {
	updates.Chan
}

func NewMineChan(activated bool) MineChan {
	return &matchChan{Chan: updates.NewChan(activated)}
}

func (ch *matchChan) Push(match T) {
	ch.Chan.Push(match)
}

func (ch *matchChan) Get() T {
	match, ok := ch.Chan.Get().(T)
	if !ok {
		return nil
	}
	return match
}
