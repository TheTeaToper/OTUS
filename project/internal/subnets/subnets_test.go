package subnets

import "testing"

func TestIPList_AddAndContains(t *testing.T) {
	list := NewSubnets()

	// Из ТЗ: 192.1.1.0/25 — это диапазон от 192.1.1.0 до 192.1.1.127
	err := list.Add("192.1.1.0/25")
	if err != nil {
		t.Fatalf("Не удалось добавить подсеть: %v", err)
	}

	// Тестируем IP внутри подсети
	if !list.Contains("192.1.1.50") {
		t.Error("IP 192.1.1.50 должен входить в подсеть 192.1.1.0/25")
	}
	if !list.Contains("192.1.1.125") {
		t.Error("IP 192.1.1.125 должен входить в подсеть 192.1.1.0/25")
	}

	// Тестируем IP вне подсети (начиная со 128 маска /25 уже не покрывает)
	if list.Contains("192.1.1.130") {
		t.Error("IP 192.1.1.130 НЕ должен входить в подсеть 192.1.1.0/25")
	}
	if list.Contains("10.0.0.1") {
		t.Error("IP 10.0.0.1 НЕ должен входить в подсеть 192.1.1.0/25")
	}
}

func TestIPList_Delete(t *testing.T) {
	list := NewSubnets()
	cidr := "10.0.0.0/24"

	_ = list.Add(cidr)
	if !list.Contains("10.0.0.5") {
		t.Fatal("IP должен быть в списке")
	}

	// Удаляем подсеть
	err := list.Delete(cidr)
	if err != nil {
		t.Fatalf("Ошибка удаления: %v", err)
	}

	if list.Contains("10.0.0.5") {
		t.Error("IP не должен определяться после удаления подсети")
	}
}

func TestIPList_InvalidCIDR(t *testing.T) {
	list := NewSubnets()

	// Передаем некорректный формат строки
	err := list.Add("not-an-ip-or-cidr")
	if err == nil {
		t.Error("Ожидалась ошибка для некорректного формата CIDR")
	}
}
