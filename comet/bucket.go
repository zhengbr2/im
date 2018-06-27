package main

import (
	"sync"
)

type BucketOptions struct {
	ChannelSize int32
}

type Bucket struct {
	cLock    sync.RWMutex
	boptions BucketOptions

	chs map[string]*Channel
}

func NewBucket(boptions BucketOptions) (b *Bucket) {
	b = new(Bucket)
	b.boptions = boptions
	b.chs = make(map[string]*Channel, boptions.ChannelSize)
	return
}

func (b *Bucket) Put(key string, ch *Channel) (err error) {
	b.cLock.Lock()
	defer b.cLock.Unlock()

	b.chs[key] = ch
	return
}

func (b *Bucket) Del(key string) {
	b.cLock.Lock()
	defer b.cLock.Unlock()
	var (
		ch *Channel
	)
	if ch = b.chs[key]; ch != nil {
		delete(b.chs, key)
	}
}

func (b *Bucket) Channel(key string) (ch *Channel) {
	b.cLock.RLock()
	defer b.cLock.RUnlock()

	ch = b.chs[key]
	return
}
