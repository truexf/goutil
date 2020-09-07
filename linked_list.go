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
	Len        int
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

func (m *LinkedList) InsertBefore(data interface{}, relative *LinkedNode) bool {
	if m == nil || data == nil || relative == nil {
		return false
	}
	m.internalLock()
	defer m.internalUnlock()
	if relative == m.head {
		m.PushHead(data, false)
		return true
	}
	node := &LinkedNode{Data: data, PriorNode: relative.PriorNode, NextNode: relative}
	if relative.PriorNode != nil {
		relative.PriorNode.NextNode = node
	}
	relative.PriorNode = node
	m.Len ++
	return true
}

func (m *LinkedList) InsertAfter(data interface{}, relative *LinkedNode) bool {
	if m == nil || data == nil || relative == nil {
		return false
	}
	m.internalLock()
	defer m.internalUnlock()
	if relative == m.tail {
		m.PushTail(data, false)
		return true
	}
	node := &LinkedNode{Data: data, PriorNode: relative, NextNode: relative.NextNode}
	if relative.NextNode != nil {
		relative.NextNode.PriorNode = node
	}
	relative.NextNode = node
	m.Len ++
	return true
}

func (m *LinkedList) PushTail(data interface{}, lock bool) {
	if lock {
		m.internalLock()
		defer m.internalUnlock()
	}
	node := &LinkedNode{PriorNode: m.tail, NextNode: nil, Data: data}
	if m.tail != nil {
		node.PriorNode = m.tail
		m.tail.NextNode = node
	}
	m.tail = node
	if m.head == nil {
		m.head = node
	}
	m.Len ++
}

func (m *LinkedList) PushHead(data interface{}, lock bool) {
	if lock {
		m.internalLock()
		defer m.internalUnlock()
	}
	node := &LinkedNode{PriorNode: nil, NextNode: m.head, Data: data}
	if m.head != nil {
		node.NextNode = m.head
		m.head.PriorNode = node
	}
	m.head = node
	if m.tail == nil {
		m.tail = node
	}
	m.Len ++
}

func (m *LinkedList) Delete(node *LinkedNode) {
	if m == nil || node == nil {
		return
	}
	m.internalLock()
	defer m.internalUnlock()
	if node == m.head {
		m.PopHead(false)
		return
	} else if node == m.tail {
		m.PopTail(false)
		return
	} else {
		if node.PriorNode != nil {
			node.PriorNode.NextNode = node.NextNode
		}
		if node.NextNode != nil {
			node.NextNode.PriorNode = node.PriorNode
		}
		m.Len --
	}
}

func (m *LinkedList) PopTail(lock bool) interface{} {
	if lock {
		m.internalLock()
		defer m.internalUnlock()
	}
	ret := m.tail
	if ret != nil {
		if ret.PriorNode != nil {
			m.tail = ret.PriorNode
			ret.PriorNode.NextNode = nil
		} else {
			if m.tail == m.head {
				m.tail = nil
				m.head = nil
			}
		}
		m.Len --
	}
	if ret == nil {
		return nil
	}
	return ret.Data
}

func (m *LinkedList) PopHead(lock bool) interface{} {
	if lock {
		m.internalLock()
		defer m.internalUnlock()
	}
	ret := m.head
	if ret != nil {
		if ret.NextNode != nil {
			m.head = ret.NextNode
			ret.NextNode.PriorNode = nil
		} else {
			if m.tail == m.head {
				m.tail = nil
				m.head = nil
			}
		}
		m.Len --
	}
	if ret == nil {
		return nil
	}
	return ret.Data
}

func (m *LinkedList) Iterate(iterator func(node *LinkedNode, canceled *bool)) {
	m.internalLock()
	defer m.internalUnlock()
	node := m.head
	for {
		if node == nil {
			break
		}
		canceled := false
		//if node.Data.(int) == 6 {
		//	fmt.Printf("*%d-%d\n", node.PriorNode.Data.(int), node.NextNode.Data.(int))
		//}
		iterator(node, &canceled)
		if canceled {
			break
		}
		node = node.NextNode
	}
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
