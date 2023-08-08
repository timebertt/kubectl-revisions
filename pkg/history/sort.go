package history

import (
	"sort"
)

// Sort sorts the given Revisions in-place by revision number (ascending).
func Sort(r Revisions) {
	sort.Sort(ByNumber(r))
}

// ByNumber returns an implementation of the sort.Interface for sorting Revisions by number.
func ByNumber(list Revisions) sort.Interface {
	return sortableRevisions{
		list: list,
		less: func(a, b Revision) bool {
			return a.Number() < b.Number()
		},
	}
}

type sortableRevisions struct {
	list Revisions
	less func(a, b Revision) bool
}

func (s sortableRevisions) Len() int {
	return len(s.list)
}

func (s sortableRevisions) Swap(i, j int) {
	s.list[i], s.list[j] = s.list[j], s.list[i]
}

func (s sortableRevisions) Less(i, j int) bool {
	return s.less(s.list[i], s.list[j])
}
