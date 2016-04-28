package processor

// ========= Node in bucket counter
type bcNode struct {
  count int
  idx int
  prev *bcNode
  next *bcNode

  isBottom bool
  isTop bool
}

// Drop from top-list
func (n *bcNode) remove() {
  if n.next != nil { n.next.prev = n.prev }
  if n.prev != nil { n.prev.next = n.next }

  n.next = nil
  n.prev = nil
}

// Add to top-list behind given successor
func (n *bcNode) insertBehind(successor *bcNode) {
  if successor.prev == nil { return }
  n.prev = successor.prev
  successor.prev.next = n
  n.next = successor
  n.next.prev = n
}


// ========= Bucket counter
type bucketCounter struct {
  topSize int
  nodes []*bcNode
  bottom *bcNode
  top *bcNode
}

// Maintains a count for each bucked and a running list of the top N counts
func newBucketCounter(numIndices int, topSize int) *bucketCounter {
  if numIndices < topSize {
    panic("Cannot create bucket counter with topSize > number of indices")
  }

  nodes := make([]*bcNode, numIndices)
  bottom := &bcNode{idx: -1, isBottom: true}
  top := &bcNode{idx: -1, isTop: true}

  for i := range nodes {
    nodes[i] = &bcNode{idx: i}
  }

  prevNode := bottom
  for i := 0; i < topSize; i++ {
    nodes[i].prev = prevNode
    prevNode.next = nodes[i]
    prevNode = nodes[i]
  }
  prevNode.next = top
  top.prev = prevNode

  return &bucketCounter{
    topSize: topSize,
    nodes: nodes,
    bottom: bottom,
    top: top,
  }
}


// Increment the bucket at index idx
func (bc *bucketCounter) increment(idx int) {
  node := bc.nodes[idx]
  node.count += 1

  // Not currently in the top
  if node.next == nil {
    bottomNode := bc.bottom.next
    if node.count <= bottomNode.count { return }

    // Add to the bottom to the top-list
    successor := bottomNode.next
    bottomNode.remove()
    node.insertBehind(successor)
  }

  // Already the hightest count
  if node.next.isTop { return }

  // Move up top-list as necessary
  curSsr := node.next
  for {
    if node.count <= curSsr.count || curSsr.isTop {
      node.remove()
      node.insertBehind(curSsr)
      break
    }

    curSsr = curSsr.next
  }
}


type countResult struct {
  idx int
  count int
}

// Return indexes of top N buckets and their counts
func (bc *bucketCounter) getTop() []countResult {
  res := make([]countResult, 0, bc.topSize)
  for node := bc.top.prev; !node.isBottom; node = node.prev {
    res = append(res, countResult{idx: node.idx, count: node.count})
  }
  return res
}
