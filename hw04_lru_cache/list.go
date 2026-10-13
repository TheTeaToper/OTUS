package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	// List // Remove me after realization.
	FirstItem *ListItem
	LastItem  *ListItem
	Length    int
	// Place your code here.
}

func NewList() List {
	return new(list)
}

func (l *list) Len() int {
	return l.Length
}

func (l *list) Front() *ListItem {
	return l.FirstItem
}

func (l *list) Back() *ListItem {
	return l.LastItem
}

func (l *list) PushFront(v interface{}) *ListItem {
	newItem := ListItem{Value: v}
	if l.Length == 0 {
		l.FirstItem = &newItem
		l.LastItem = &newItem
	} else {
		newItem.Next = l.FirstItem
		l.FirstItem.Prev = &newItem
		l.FirstItem = &newItem
	}
	l.Length++
	return &newItem
}

func (l *list) PushBack(v interface{}) *ListItem {
	newItem := ListItem{Value: v}
	if l.Length == 0 {
		l.FirstItem = &newItem
		l.LastItem = &newItem
	} else {
		newItem.Prev = l.LastItem
		l.LastItem.Next = &newItem
		l.LastItem = &newItem
	}
	l.Length++
	return &newItem
}

func (l *list) Remove(i *ListItem) {
	if i == nil || l.Length == 0 {
		return
	}
	l.detachItem(i)
	i.Prev = nil
	i.Next = nil
	l.Length--
}

func (l *list) detachItem(i *ListItem) {
	if i.Prev == nil {
		l.FirstItem = i.Next
		if i.Next != nil {
			i.Next.Prev = nil
		}
	} else {
		i.Prev.Next = i.Next
	}
	if i.Next == nil {
		l.LastItem = i.Prev
		if i.Prev != nil {
			i.Prev.Next = nil
		}
	} else {
		i.Next.Prev = i.Prev
	}
}

func (l *list) MoveToFront(i *ListItem) {
	if i == nil || l.Length < 2 || l.FirstItem == i {
		return
	}
	l.detachItem(i)
	i.Prev = nil
	i.Next = l.FirstItem
	l.FirstItem.Prev = i
	l.FirstItem = i
}
