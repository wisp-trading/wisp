package browse

import (
	"testing"

	"github.com/wisp-trading/sdk/pkg/types/config"
)

func TestStrategyListPagination(t *testing.T) {
	m := &strategyListView{
		pageSize: 3,
		pageNum:  1,
		cursor:   0,
		strategies: []config.Strategy{
			{Name: "a"}, {Name: "b"}, {Name: "c"},
			{Name: "d"}, {Name: "e"}, {Name: "f"},
			{Name: "g"},
		},
	}
	if m.totalPages() != 3 {
		t.Fatalf("pages=%d", m.totalPages())
	}
	if m.pageStart() != 0 || m.pageEnd() != 3 {
		t.Fatalf("page1 range %d-%d", m.pageStart(), m.pageEnd())
	}

	m.cursor = 4
	m.syncPageFromCursor()
	if m.pageNum != 2 {
		t.Fatalf("page=%d want 2", m.pageNum)
	}
	if m.pageStart() != 3 || m.pageEnd() != 6 {
		t.Fatalf("page2 range %d-%d", m.pageStart(), m.pageEnd())
	}

	m.pageNum = 3
	m.clampCursorToPage()
	if m.cursor != 6 {
		t.Fatalf("cursor=%d want 6", m.cursor)
	}
}
