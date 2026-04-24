### Lineage II Chronicle Dark Ages

------

#### Текущий прогресс:
- **LoginServer:** Полная авторизация, шифрование Blowfish, выбор сервера.
- **GameServer:** 
  - Протоколы 411/414/419.
  - Реализован Handshake и XOR-шифрование.
  - Вход в мир, выбор персонажа, создание персонажа.
  - Движение (MoveToLocation) с сохранением координат в БД.
  - Чат (Say2) и системные сообщения.
  - Корректный Logout и Restart.
- **Database:** PostgreSQL

<br>

#### Технологии:
- Go 1.21+
- PostgreSQL (pgx v5)
- Клиент Lineage 2 Chronicle 1: Harbingers of War

<br>

#### Запуск:
1. Поднять PostgreSQL (docker compose up)
2. Запустить `go run cmd/loginserver/main.go`.
3. Запустить `go run cmd/gameserver/main.go`.

<br>

#### TODO:
- ValidateLocation
- NPC
- Items
- Base actions (SocialAction etc)