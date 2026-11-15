Примеры использования команд администратора (при запущенном go run cmd/main.go):
• Сброс бакетов:bash
./ab-admin reset --login=admin --ip=192.1.1.5

• Добавление подсети в черный список:bash
./ab-admin blacklist add --subnet=192.1.1.0/25

• Удаление подсети из белого списка:bash
./ab-admin whitelist remove --subnet=10.0.0.0/24