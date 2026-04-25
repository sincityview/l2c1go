### Lineage II Dark Ages

------

#### Текущий прогресс:
- LoginServer: 
  - Полная авторизация
  - Шифрование Blowfish
  - Выбор сервера

- GameServer: 
  - Протоколы 411/414/419
  - Handshake и XOR-шифрование
  - Вход в лобби, создание персонажа, выбор персонажа
  - Загрузка мира
  - Движение с сохранением координат в БД
  - Чат и системные сообщения
  - Действия, выбор цели, радар, карта, инвентарь
  - Logout и Restart

- Database: 
  - MariaDB

- Client:
  - Lineage 2 Chronicle 1: Harbingers of War

#### Запуск:
1. Поднять MariaDB (docker compose up)

import darkages.sql

accounts:
```sql
INSERT INTO accounts (login, password) 
VALUES ('admin', SHA2('admin', 256));
```

bbs:
receiver_id = 0 (all), charId (character)
sender_id = 0 (admin), charId (character)
```sql
INSERT INTO bbs (receiver_id, sender_id, subject, message)
VALUES ('0', '0', 'Subject', 'Message');
```

2. Запустить `go run cmd/loginserver/main.go`.
3. Запустить `go run cmd/gameserver/main.go`.

#### TODO:
- Inventory items, eqip items
- NPC
- Monsters
