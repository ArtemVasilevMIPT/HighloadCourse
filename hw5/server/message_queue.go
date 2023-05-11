package main

import (
	pb "hw5/proto/messenger"
)

type Item struct {
	Content  *pb.Mail
	Priority int64
}

type PriorityQueue []*Item

func (p PriorityQueue) Len() int {
	return len(p)
}

func (p PriorityQueue) Less(i, j int) bool {
	return p[i].Priority < p[j].Priority
}

func (p PriorityQueue) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p *PriorityQueue) Push(x any) {
	item := x.(*Item)
	*p = append(*p, item)
}

func (p *PriorityQueue) Pop() any {
	old := *p
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*p = old[0 : n-1]
	return item
}
