package heap

//ImplicitHeapMax A dynamic list of numbers, stored as a Binary tree in a dynamic slice.
//Used to quickly get the biggest number from a list/queue/priority queue.
//
//Inherits all the methods of ImplicitHeapMin.
type ImplicitHeapMax struct {
	ImplicitHeapMin
}

func maxShouldGoUp(p, c implicitHeapNode) bool {
	return c.priority > p.priority
}

//NewImplicitHeapMax Builds an empty Implicit Heap Max struct.
func NewImplicitHeapMax(autoLockMutex bool) *ImplicitHeapMax {
	h := &ImplicitHeapMax{}
	h.compare = maxShouldGoUp
	h.autoLockMutex = autoLockMutex
	h.Reset()
	return h
}
