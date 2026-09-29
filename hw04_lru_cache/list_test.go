package hw04lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("complex", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})
}

func TestListOrder(t *testing.T) {
	t.Run("Проверка порядка элементов", func(t *testing.T) {
		l := NewList()
		l.PushBack("Второй")
		l.PushFront("Первый")
		l.PushBack("Третий")
		require.Equal(t, 3, l.Len())
		require.Nil(t, l.Front().Prev)
		require.Equal(t, "Первый", l.Front().Value)
		require.Equal(t, "Второй", l.Front().Next.Value)
		require.Equal(t, "Третий", l.Back().Value)
		require.Nil(t, l.Back().Next)
	})
}

func TestListCleaning(t *testing.T) {
	t.Run("Проверка удаления", func(t *testing.T) {
		l := NewList()
		firstItem := l.PushFront("Первый")
		secondItem := l.PushBack("Второй")
		l.PushBack("Третий")
		l.Remove(l.Back())
		if l.Back() != secondItem {
			t.Error("Последним должен быть второй элемент")
		}
		l.Remove(firstItem)
		if l.Front() != secondItem {
			t.Error("Первым должен быть второй элемент")
		}
		require.Equal(t, 1, l.Len())
	})
}
