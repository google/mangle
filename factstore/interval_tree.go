// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package factstore

import (
	"codeberg.org/TauCeti/mangle-go/ast"
)

// IntervalTree is an augmented interval tree for efficient interval queries.
// It uses a balanced BST (AVL tree) where each node stores:
// - An interval with start and end times
// - The maximum end time in the subtree rooted at this node
// This enables O(log n + k) queries where n is the number of intervals
// and k is the number of matching results.
type IntervalTree struct {
	root *treeNode
	size int
}

// treeNode represents a node in the interval tree.
type treeNode struct {
	interval ast.Interval
	maxEnd   int64 // Maximum end time in subtree
	height   int   // Height for AVL balancing
	left     *treeNode
	right    *treeNode
}

// NewIntervalTree creates a new interval tree.
func NewIntervalTree() *IntervalTree {
	return &IntervalTree{}
}

// Insert adds an interval to the tree.
// Returns true if the interval was added, false if it was a duplicate.
func (t *IntervalTree) Insert(interval ast.Interval) bool {
	// Check for duplicate
	if t.contains(interval) {
		return false
	}

	t.root = t.insert(t.root, interval)
	t.size++
	return true
}

// insert recursively inserts an interval and rebalances.
func (t *IntervalTree) insert(node *treeNode, interval ast.Interval) *treeNode {
	if node == nil {
		return &treeNode{
			interval: interval,
			maxEnd:   GetEndTime(interval),
			height:   1,
		}
	}

	start := GetStartTime(interval)
	nodeStart := GetStartTime(node.interval)

	if start < nodeStart {
		node.left = t.insert(node.left, interval)
	} else {
		node.right = t.insert(node.right, interval)
	}

	// Update max end
	node.maxEnd = maxInt(node.maxEnd, GetEndTime(interval))

	// Rebalance
	return t.rebalance(node)
}

// contains checks if an exact interval exists in the tree.
func (t *IntervalTree) contains(interval ast.Interval) bool {
	return t.findExact(t.root, interval)
}

// findExact searches for an exact interval match.
func (t *IntervalTree) findExact(node *treeNode, interval ast.Interval) bool {
	if node == nil {
		return false
	}

	if node.interval.Equals(interval) {
		return true
	}

	start := GetStartTime(interval)
	nodeStart := GetStartTime(node.interval)

	if start < nodeStart {
		return t.findExact(node.left, interval)
	}
	// Search both subtrees for same start time
	if t.findExact(node.right, interval) {
		return true
	}
	if start == nodeStart {
		return t.findExact(node.left, interval)
	}
	return false
}

// QueryPoint returns all intervals containing the given timestamp.
func (t *IntervalTree) QueryPoint(timestamp int64, fn func(ast.Interval) error) error {
	return t.queryPoint(t.root, timestamp, fn)
}

// queryPoint recursively queries for intervals containing a point.
func (t *IntervalTree) queryPoint(node *treeNode, timestamp int64, fn func(ast.Interval) error) error {
	if node == nil {
		return nil
	}

	// If max end in this subtree is less than timestamp, no intervals here contain it
	if node.maxEnd < timestamp {
		return nil
	}

	// Check left subtree
	if err := t.queryPoint(node.left, timestamp, fn); err != nil {
		return err
	}

	// Check current node
	if containsTimestamp(node.interval, timestamp) {
		if err := fn(node.interval); err != nil {
			return err
		}
	}

	// Check right subtree only if its intervals could contain the timestamp
	// (start time <= timestamp is guaranteed by the query, we check end time via maxEnd)
	nodeStart := GetStartTime(node.interval)
	if timestamp >= nodeStart {
		if err := t.queryPoint(node.right, timestamp, fn); err != nil {
			return err
		}
	}

	return nil
}

// QueryRange returns all intervals overlapping with [start, end].
func (t *IntervalTree) QueryRange(start, end int64, fn func(ast.Interval) error) error {
	return t.queryRange(t.root, start, end, fn)
}

// queryRange recursively queries for overlapping intervals.
// Two intervals [s1, e1] and [s2, e2] overlap if s1 <= e2 AND s2 <= e1.
func (t *IntervalTree) queryRange(node *treeNode, start, end int64, fn func(ast.Interval) error) error {
	if node == nil {
		return nil
	}

	// If max end in this subtree is less than query start, no overlap possible
	if node.maxEnd < start {
		return nil
	}

	// Check left subtree
	if err := t.queryRange(node.left, start, end, fn); err != nil {
		return err
	}

	// Check current node for overlap
	nodeStart := GetStartTime(node.interval)
	nodeEnd := GetEndTime(node.interval)

	// Overlap condition: nodeStart <= end AND start <= nodeEnd
	if nodeStart <= end && start <= nodeEnd {
		if err := fn(node.interval); err != nil {
			return err
		}
	}

	// Check right subtree only if possible overlap exists
	if nodeStart <= end {
		if err := t.queryRange(node.right, start, end, fn); err != nil {
			return err
		}
	}

	return nil
}

// All calls fn for every interval in the tree (in-order traversal).
func (t *IntervalTree) All(fn func(ast.Interval) error) error {
	return t.inOrder(t.root, fn)
}

// inOrder performs an in-order traversal.
func (t *IntervalTree) inOrder(node *treeNode, fn func(ast.Interval) error) error {
	if node == nil {
		return nil
	}

	if err := t.inOrder(node.left, fn); err != nil {
		return err
	}
	if err := fn(node.interval); err != nil {
		return err
	}
	return t.inOrder(node.right, fn)
}

// Size returns the number of intervals in the tree.
func (t *IntervalTree) Size() int {
	return t.size
}

// Clear removes all intervals from the tree.
func (t *IntervalTree) Clear() {
	t.root = nil
	t.size = 0
}

// Rebuild reconstructs the tree from a list of intervals (for coalescing).
func (t *IntervalTree) Rebuild(intervals []ast.Interval) {
	t.Clear()
	for _, interval := range intervals {
		t.Insert(interval)
	}
}

// AVL tree helper functions

// height returns the height of a node.
func height(node *treeNode) int {
	if node == nil {
		return 0
	}
	return node.height
}

// updateHeight updates the height of a node based on children.
func updateHeight(node *treeNode) {
	left := height(node.left)
	right := height(node.right)
	if left > right {
		node.height = 1 + left
	} else {
		node.height = 1 + right
	}
}

// balanceFactor returns the balance factor of a node.
func balanceFactor(node *treeNode) int {
	if node == nil {
		return 0
	}
	return height(node.left) - height(node.right)
}

// updateMaxEnd updates the maxEnd value of a node based on children.
func updateMaxEnd(node *treeNode) {
	node.maxEnd = GetEndTime(node.interval)
	if node.left != nil && node.left.maxEnd > node.maxEnd {
		node.maxEnd = node.left.maxEnd
	}
	if node.right != nil && node.right.maxEnd > node.maxEnd {
		node.maxEnd = node.right.maxEnd
	}
}

// rotateRight performs a right rotation.
func (t *IntervalTree) rotateRight(y *treeNode) *treeNode {
	x := y.left
	z := x.right

	x.right = y
	y.left = z

	updateHeight(y)
	updateMaxEnd(y)
	updateHeight(x)
	updateMaxEnd(x)

	return x
}

// rotateLeft performs a left rotation.
func (t *IntervalTree) rotateLeft(x *treeNode) *treeNode {
	y := x.right
	z := y.left

	y.left = x
	x.right = z

	updateHeight(x)
	updateMaxEnd(x)
	updateHeight(y)
	updateMaxEnd(y)

	return y
}

// rebalance rebalances a node after insertion.
func (t *IntervalTree) rebalance(node *treeNode) *treeNode {
	updateHeight(node)
	updateMaxEnd(node)

	balance := balanceFactor(node)

	// Left-heavy
	if balance > 1 {
		if balanceFactor(node.left) < 0 {
			// Left-Right case
			node.left = t.rotateLeft(node.left)
		}
		// Left-Left case
		return t.rotateRight(node)
	}

	// Right-heavy
	if balance < -1 {
		if balanceFactor(node.right) > 0 {
			// Right-Left case
			node.right = t.rotateRight(node.right)
		}
		// Right-Right case
		return t.rotateLeft(node)
	}

	return node
}

// max returns the maximum of two int64 values.
func maxInt(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// Helper functions for interval time extraction

// GetStartTime extracts the start time from an interval.
// Returns MinInt64 for unbounded start (negative infinity).
func GetStartTime(interval ast.Interval) int64 {
	switch interval.Start.Type {
	case ast.TimestampBound:
		return interval.Start.Timestamp
	case ast.NegativeInfinityBound:
		return minInt64
	case ast.PositiveInfinityBound:
		return maxInt64
	default:
		return 0
	}
}

// GetEndTime extracts the end time from an interval.
// Returns MaxInt64 for unbounded end (positive infinity).
func GetEndTime(interval ast.Interval) int64 {
	switch interval.End.Type {
	case ast.TimestampBound:
		return interval.End.Timestamp
	case ast.PositiveInfinityBound:
		return maxInt64
	case ast.NegativeInfinityBound:
		return minInt64
	default:
		return 0
	}
}

// containsTimestamp checks if an interval contains a timestamp.
func containsTimestamp(interval ast.Interval, timestamp int64) bool {
	start := GetStartTime(interval)
	end := GetEndTime(interval)
	return start <= timestamp && timestamp <= end
}

const (
	minInt64 = -1 << 63
	maxInt64 = 1<<63 - 1
)
