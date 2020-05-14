// 双向链表
package goutil

import (
	"sync"
)

type LinkedNode struct {
	PriorNode *LinkedNode
	NextNode  *LinkedNode
	Data      interface{}
}
type LinkedList struct {
	lock       sync.Mutex
	threadSafe bool
	head       *LinkedNode
	tail       *LinkedNode
}

func NewLinkedList(threadSafe bool) *LinkedList {
	ret := new(LinkedList)
	ret.head = nil
	ret.tail = nil
	ret.threadSafe = threadSafe
	return ret
}

func (m *LinkedList) internalLock() {
	if m.threadSafe {
		m.lock.Lock()
	}
}

func (m *LinkedList) internalUnlock() {
	if m.threadSafe {
		m.lock.Unlock()
	}
}

func (m *LinkedList) PushTail(data interface{}) {
	m.internalLock()
	node := &LinkedNode{PriorNode: m.tail, NextNode: nil, Data: data}
	if m.tail != nil {
		node.PriorNode = m.tail
		m.tail.NextNode = node
	}
	m.tail = node
	if m.head == nil {
		m.head = node
	}
	m.internalUnlock()
}

func (m *LinkedList) PushHead(data interface{}) {
	m.internalLock()
	node := &LinkedNode{PriorNode: nil, NextNode: m.head, Data: data}
	if m.head != nil {
		node.NextNode = m.head
		m.head.PriorNode = node
	}
	m.head = node
	if m.tail == nil {
		m.tail = node
	}
	m.internalUnlock()
}

func (m *LinkedList) PopTail() interface{} {
	m.internalLock()
	ret := m.tail
	if ret != nil {
		if ret.PriorNode != nil {
			ret.PriorNode.NextNode = nil
			m.tail = ret.PriorNode
		} else {
			if m.tail == m.head {
				m.tail = nil
				m.head = nil
			}
		}
	}
	m.internalUnlock()
	if ret == nil {
		return nil
	}
	return ret.Data
}

func (m *LinkedList) PopHead() interface{} {
	m.internalLock()
	ret := m.head
	if ret != nil {
		if ret.NextNode != nil {
			ret.NextNode.PriorNode = nil
			m.head = ret.NextNode
		} else {
			if m.tail == m.head {
				m.tail = nil
				m.head = nil
			}
		}
	}
	m.internalUnlock()
	if ret == nil {
		return nil
	}
	return ret.Data
}

func (m *LinkedList) Iterate(iterator func(nodeData interface{}, canceled *bool)) {
	m.internalLock()
	node := m.head
	for {
		if node == nil {
			break
		}
		canceled := false
		//if node.Data.(int) == 6 {
		//	fmt.Printf("*%d-%d\n", node.PriorNode.Data.(int), node.NextNode.Data.(int))
		//}
		iterator(node.Data, &canceled)
		if canceled {
			break
		}
		node = node.NextNode
	}
	m.internalUnlock()
}

func (m *LinkedList) ReverseIterate(iterator func(nodeData interface{}, canceled *bool)) {
	m.internalLock()
	node := m.tail
	for {
		if node == nil {
			break
		}
		canceled := false
		iterator(node.Data, &canceled)
		if canceled {
			break
		}
		node = node.PriorNode
	}
	m.internalUnlock()
}
