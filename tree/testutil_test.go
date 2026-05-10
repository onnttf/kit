package tree

type TestItem struct {
	ID       int
	Name     string
	ParentID int
	Sort     int
}

func keyFn(item TestItem) int { return item.ID }

func parentFn(item TestItem) (int, bool) {
	if item.ParentID == 0 || item.ParentID == item.ID {
		return 0, false
	}
	return item.ParentID, true
}

func sortFn(item TestItem) int { return item.Sort }
