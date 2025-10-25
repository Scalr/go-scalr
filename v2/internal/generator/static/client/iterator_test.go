package client

import (
	"context"
	"errors"
	"testing"
)

// TestIteratorBasic tests basic iterator functionality
func TestIteratorBasic(t *testing.T) {
	ctx := context.Background()

	// Create test data - 3 pages with 2 items each
	pages := [][]*string{
		{ptr("item1"), ptr("item2")},
		{ptr("item3"), ptr("item4")},
		{ptr("item5"), ptr("item6")},
	}

	fetchPage := func(ctx context.Context, pageNum int) ([]*string, *Pagination, error) {
		if pageNum < 1 || pageNum > len(pages) {
			return nil, nil, nil
		}

		nextPage := pageNum + 1
		var next *int
		if pageNum < len(pages) {
			next = &nextPage
		}

		pagination := &Pagination{
			CurrentPage: pageNum,
			TotalPages:  len(pages),
			PageSize:    2,
			NextPage:    next,
		}

		return pages[pageNum-1], pagination, nil
	}

	iter := NewIterator(ctx, 2, fetchPage)

	// Iterate through all items
	var collected []string
	for iter.Next() {
		collected = append(collected, *iter.Value())
	}

	if err := iter.Err(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []string{"item1", "item2", "item3", "item4", "item5", "item6"}
	if len(collected) != len(expected) {
		t.Fatalf("Expected %d items, got %d", len(expected), len(collected))
	}

	for i, item := range collected {
		if item != expected[i] {
			t.Errorf("Item %d: expected %q, got %q", i, expected[i], item)
		}
	}
}

// TestIteratorEmpty tests iterator with no data
func TestIteratorEmpty(t *testing.T) {
	ctx := context.Background()

	fetchPage := func(ctx context.Context, pageNum int) ([]*string, *Pagination, error) {
		return []*string{}, &Pagination{CurrentPage: 1, TotalPages: 1}, nil
	}

	iter := NewIterator(ctx, 10, fetchPage)

	if iter.Next() {
		t.Error("Next() should return false for empty data")
	}

	if err := iter.Err(); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestIteratorError tests iterator with fetch error
func TestIteratorError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("fetch failed")

	fetchPage := func(ctx context.Context, pageNum int) ([]*string, *Pagination, error) {
		return nil, nil, expectedErr
	}

	iter := NewIterator(ctx, 10, fetchPage)

	if iter.Next() {
		t.Error("Next() should return false on error")
	}

	if iter.Err() != expectedErr {
		t.Errorf("Err() = %v, want %v", iter.Err(), expectedErr)
	}
}

// TestIteratorContextCancellation tests context cancellation
func TestIteratorContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	fetchPage := func(ctx context.Context, pageNum int) ([]*string, *Pagination, error) {
		return []*string{ptr("item1"), ptr("item2")}, &Pagination{CurrentPage: 1, TotalPages: 1}, nil
	}

	iter := NewIterator(ctx, 10, fetchPage)

	// Get first item
	if !iter.Next() {
		t.Fatal("First Next() should succeed")
	}

	// Cancel context
	cancel()

	// Next call should fail
	if iter.Next() {
		t.Error("Next() should fail after context cancellation")
	}

	if iter.Err() != context.Canceled {
		t.Errorf("Err() = %v, want %v", iter.Err(), context.Canceled)
	}
}

// TestIteratorPageInfo tests PageInfo functionality
func TestIteratorPageInfo(t *testing.T) {
	ctx := context.Background()

	pages := [][]*string{
		{ptr("item1"), ptr("item2")},
		{ptr("item3"), ptr("item4")},
	}

	fetchPage := func(ctx context.Context, pageNum int) ([]*string, *Pagination, error) {
		if pageNum < 1 || pageNum > len(pages) {
			return nil, nil, nil
		}

		nextPage := pageNum + 1
		var next *int
		if pageNum < len(pages) {
			next = &nextPage
		}

		return pages[pageNum-1], &Pagination{
			CurrentPage: pageNum,
			TotalPages:  len(pages),
			PageSize:    2,
			NextPage:    next,
		}, nil
	}

	iter := NewIterator(ctx, 2, fetchPage)

	// Before starting
	info := iter.PageInfo()
	if info.HasStarted {
		t.Error("HasStarted should be false before first Next()")
	}

	// First page
	iter.Next()
	info = iter.PageInfo()
	if !info.HasStarted {
		t.Error("HasStarted should be true after first Next()")
	}
	if info.CurrentPage != 1 {
		t.Errorf("CurrentPage = %d, want 1", info.CurrentPage)
	}
	if info.TotalPages != 2 {
		t.Errorf("TotalPages = %d, want 2", info.TotalPages)
	}

	// Consume first page
	iter.Next()

	// Second page
	iter.Next()
	info = iter.PageInfo()
	if info.CurrentPage != 2 {
		t.Errorf("CurrentPage = %d, want 2", info.CurrentPage)
	}

	// Consume all items
	for iter.Next() {
	}

	info = iter.PageInfo()
	if !info.IsDone {
		t.Error("IsDone should be true after consuming all items")
	}
}

// TestIteratorRemaining tests Remaining() method
func TestIteratorRemaining(t *testing.T) {
	ctx := context.Background()

	pages := [][]*string{
		{ptr("item1"), ptr("item2"), ptr("item3")},
		{ptr("item4"), ptr("item5"), ptr("item6")},
	}

	fetchPage := func(ctx context.Context, pageNum int) ([]*string, *Pagination, error) {
		if pageNum < 1 || pageNum > len(pages) {
			return nil, nil, nil
		}

		nextPage := pageNum + 1
		var next *int
		if pageNum < len(pages) {
			next = &nextPage
		}

		return pages[pageNum-1], &Pagination{
			CurrentPage: pageNum,
			TotalPages:  len(pages),
			PageSize:    3,
			TotalCount:  6, // 2 pages * 3 items per page
			NextPage:    next,
		}, nil
	}

	iter := NewIterator(ctx, 3, fetchPage)

	// Before starting
	if remaining := iter.Remaining(); remaining != -1 {
		t.Errorf("Remaining() = %d before start, want -1", remaining)
	}

	// After first item
	iter.Next()
	remaining := iter.Remaining()
	// Should be: 2 remaining in current page + 1 page * 3 items = 5
	if remaining != 5 {
		t.Errorf("Remaining() = %d after first item, want 5", remaining)
	}

	// After three more items
	iter.Next()
	iter.Next()
	iter.Next()
	remaining = iter.Remaining()
	// Should be: 2 remaining in last page
	if remaining != 2 {
		t.Errorf("Remaining() = %d after first item, want 2", remaining)
	}

	// Consume all
	for iter.Next() {
	}

	remaining = iter.Remaining()
	if remaining != 0 {
		t.Errorf("Remaining() = %d after consuming all, want 0", remaining)
	}
}

// TestIteratorCollect tests Collect() method
func TestIteratorCollect(t *testing.T) {
	ctx := context.Background()

	pages := [][]*string{
		{ptr("item1"), ptr("item2")},
		{ptr("item3"), ptr("item4")},
	}

	fetchPage := func(ctx context.Context, pageNum int) ([]*string, *Pagination, error) {
		if pageNum < 1 || pageNum > len(pages) {
			return nil, nil, nil
		}

		nextPage := pageNum + 1
		var next *int
		if pageNum < len(pages) {
			next = &nextPage
		}

		return pages[pageNum-1], &Pagination{
			CurrentPage: pageNum,
			TotalPages:  len(pages),
			NextPage:    next,
		}, nil
	}

	iter := NewIterator(ctx, 2, fetchPage)

	collected, err := iter.Collect()
	if err != nil {
		t.Fatalf("Collect() error: %v", err)
	}

	if len(collected) != 4 {
		t.Errorf("Collected %d items, want 4", len(collected))
	}

	expected := []string{"item1", "item2", "item3", "item4"}
	for i, item := range collected {
		if *item != expected[i] {
			t.Errorf("Item %d: got %q, want %q", i, *item, expected[i])
		}
	}
}

// TestIteratorCollectWithError tests Collect() with error
func TestIteratorCollectWithError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("fetch error")

	fetchPage := func(ctx context.Context, pageNum int) ([]*string, *Pagination, error) {
		if pageNum == 2 {
			return nil, nil, expectedErr
		}
		return []*string{ptr("item1")}, &Pagination{CurrentPage: 1, TotalPages: 2, NextPage: ptr(2)}, nil
	}

	iter := NewIterator(ctx, 1, fetchPage)

	_, err := iter.Collect()
	if err != expectedErr {
		t.Errorf("Collect() error = %v, want %v", err, expectedErr)
	}
}

// TestIteratorValuePanic tests that Value() panics when called incorrectly
func TestIteratorValuePanic(t *testing.T) {
	ctx := context.Background()

	fetchPage := func(ctx context.Context, pageNum int) ([]*string, *Pagination, error) {
		return []*string{ptr("item1")}, &Pagination{CurrentPage: 1, TotalPages: 1}, nil
	}

	iter := NewIterator(ctx, 1, fetchPage)

	// Value() before Next() should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Value() should panic when called before Next()")
		}
	}()

	iter.Value()
}

// TestIteratorValuePanicAfterEnd tests that Value() panics after iteration ends
func TestIteratorValuePanicAfterEnd(t *testing.T) {
	ctx := context.Background()

	fetchPage := func(ctx context.Context, pageNum int) ([]*string, *Pagination, error) {
		return []*string{ptr("item1")}, &Pagination{CurrentPage: 1, TotalPages: 1}, nil
	}

	iter := NewIterator(ctx, 1, fetchPage)

	// Consume iterator
	for iter.Next() {
	}

	// Value() after Next() returns false should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Value() should panic when called after Next() returns false")
		}
	}()

	iter.Value()
}

// TestPageInfoString tests PageInfo.String() method
func TestPageInfoString(t *testing.T) {
	tests := []struct {
		name     string
		pageInfo PageInfo
		expected string
	}{
		{
			name:     "not started",
			pageInfo: PageInfo{HasStarted: false},
			expected: "not started",
		},
		{
			name: "in progress",
			pageInfo: PageInfo{
				CurrentPage: 2,
				TotalPages:  5,
				HasStarted:  true,
				IsDone:      false,
			},
			expected: "page 2/5",
		},
		{
			name: "done",
			pageInfo: PageInfo{
				CurrentPage: 5,
				TotalPages:  5,
				HasStarted:  true,
				IsDone:      true,
			},
			expected: "page 5/5 (done)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pageInfo.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestIteratorSinglePage tests iterator with a single page
func TestIteratorSinglePage(t *testing.T) {
	ctx := context.Background()

	fetchPage := func(ctx context.Context, pageNum int) ([]*string, *Pagination, error) {
		if pageNum != 1 {
			return nil, nil, nil
		}
		return []*string{ptr("item1"), ptr("item2")}, &Pagination{
			CurrentPage: 1,
			TotalPages:  1,
			NextPage:    nil, // No next page
		}, nil
	}

	iter := NewIterator(ctx, 2, fetchPage)

	count := 0
	for iter.Next() {
		count++
	}

	if err := iter.Err(); err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 items, got %d", count)
	}
}

// Helper function to create string pointer
func ptr[T any](v T) *T {
	return &v
}
