### Lineage II Dark Ages

------

#### Текущий прогресс:
- LoginServer: 
  - Полная авторизация, шифрование Blowfish
  - Выбор сервера

- GameServer: 
  - Протоколы 411/414/419
  - Реализован Handshake и XOR-шифрование
  - Вход в мир, создание персонажа, выбор персонажа
  - Движение с сохранением координат в БД
  - Чат и системные сообщения
  - Корректный Logout и Restart
  - Действия, выбор цели, радар, карта, инвентарь

- Database: 
  - MariaDB
  - Структура Mobius HoW 

#### Технологии:
- Go 1.21+
- MariaDB
- Клиент Lineage 2 Chronicle 1: Harbingers of War

#### Запуск:
1. Поднять MariaDB (docker compose up)
2. Запустить `go run cmd/loginserver/main.go`.
3. Запустить `go run cmd/gameserver/main.go`.

#### TODO:
- NPC
