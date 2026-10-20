# Амфора — лабораторная работа №4 (авторизация, Swagger, Redis)

Курс «Разработка интернет-приложений», МГТУ им. Н. Э. Баумана, ИУ5.

**Тема.** История и культура: оценка торговых связей по статистике импортных находок.

* **Услуга** — категория импортных находок. В коде, адресах и БД — **`import_categories`**.
* **Поля по теме** — дата начала и дата конца бытования типа (`date_start`, `date_end`).
  Фильтрация списка — по дате начала.

В ЛР3 приложение стало веб-сервисом (REST API) для будущего SPA: HTML-шаблоны убраны,
все ответы — JSON, все адреса начинаются с `/api`. Данные — PostgreSQL через ORM GORM,
изображения и видео услуг — файлы в MinIO.

В ЛР4 добавлены **аутентификация по JWT**, **чёрный список токенов в Redis** (выход),
**хеширование паролей bcrypt** и описание API в **Swagger**. Гостю доступно только
чтение (список и лента), методы создателя требуют токен. Создатель услуги берётся
из токена — константа-singleton из ЛР3 больше не нужна.

---

## 1. Запуск

Нужны Go 1.25+ и Docker.

```bash
go mod tidy                                             # зависимости: gin, gorm, minio-go, jwt, go-redis, swag
docker compose -f deployments/docker-compose.yml up -d  # PostgreSQL, Adminer, MinIO, Redis
go run ./cmd/migrate -drop                              # таблицы + начальные данные (пароли — bcrypt)
go run ./cmd/main                                       # API: http://localhost:8080/api
```

Документация Swagger уже сгенерирована в `docs/swagger`. После изменения комментариев
`@Summary`, `@Param` и т. д. её пересобирают:

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.4
swag init -g cmd/main/main.go -o docs/swagger
```

Пользователи из начальных данных: `umid`, `anna`, `pavel`, … — пароль у всех `rip2026`.

| Адрес | Что это |
|---|---|
| <http://localhost:8080/api/import_categories> | API (список услуг) |
| <http://localhost:8080/swagger/index.html> | Swagger UI |
| <http://localhost:8081> | Adminer: сервер `postgres`, база `amphora`, `rip` / `rip` |
| <http://localhost:9001> | консоль MinIO, `minioadmin` / `minioadmin`, бакет `import-categories` |
| `localhost:6379` | Redis, пароль `rip`: `docker exec -it rip-redis redis-cli -a rip` |

Коллекция запросов для Postman / Insomnia —
[`docs/postman/amphora.postman_collection.json`](docs/postman/amphora.postman_collection.json)
(Postman: *Import*; Insomnia: *Import → From File*). В ней папка «Основные запросы» —
10 методов API в порядке показа, и папка «Проверка ошибок» — запросы с плохими ответами.

---

## 2. Авторизация: JWT и Redis

1. **Вход.** `POST /api/users/login` с `{"login": "umid", "password": "rip2026"}`.
   Сервер находит пользователя, сравнивает пароль с хешем bcrypt и выдаёт JWT
   (подпись HS256, срок действия `JWT_TTL`, по умолчанию 1 час):
   ```json
   {"status": "success", "data": {"access_token": "eyJhbGciOi…", "token_type": "Bearer",
    "expires_in": 3600, "user": {"id": 1, "login": "umid", "is_moderator": true}}}
   ```
   В токене (payload): `user_id`, `login`, `is_moderator`, `iat`, `exp`, `iss`.
2. **Запросы.** Токен передаётся заголовком `Authorization: Bearer <access_token>`.
   В Swagger — кнопка **Authorize**, в Postman — вкладка *Authorization → Bearer Token*
   (в коллекции токен сохраняется автоматически после запроса «Аутентификация»).
3. **Проверка.** Middleware `WithAuthCheck`
   ([internal/app/handler/middleware.go](internal/app/handler/middleware.go))
   проверяет подпись и срок токена, затем ищет его в чёрном списке Redis и кладёт
   пользователя в контекст запроса. Методы читают его функцией `auth.CurrentUserID(c)`
   ([internal/app/auth/current_user.go](internal/app/auth/current_user.go)).
   * `WithAuthCheck(false)` — **методы чтения** (список, лента): работают и без токена
     (гость, признаки `is_creator` / `is_liked` = 0), а с токеном — для пользователя.
   * `WithAuthCheck(true)` — **методы создателя**: без действующего токена — **401**.
4. **Выход.** `POST /api/users/logout` кладёт токен в Redis под ключом
   `amphora.jwt.<токен>` со значением — логином пользователя и временем жизни до конца
   срока токена. После этого токен не принимается (401). Посмотреть:
   ```bash
   docker exec -it rip-redis redis-cli -a rip
   KEYS *            # amphora.jwt.eyJhbGciOi…
   GET  <ключ>       # umid — чей токен
   TTL  <ключ>       # сколько секунд ещё хранится
   ```
5. **Регистрация.** `POST /api/users/register` — в таблицу `users` пишется хеш bcrypt,
   а не пароль. Новый пользователь всегда не модератор.
6. **Автозаполнение создателя.** При `POST /api/import_categories` поле `creator_id`
   берётся из токена.

---

## 3. HTTP-методы

Ответ всегда JSON одного вида:

```json
{"status": "success", "data": { ... }, "message": "черновик создан"}
{"status": "fail", "message": "у пользователя уже есть черновик"}
```

Коды: 200 — успех, 201 — создано, 400 — неверные данные, 401 — нет токена, он
недействителен или отозван (а также неверный логин или пароль), 404 — записи нет или она
недоступна, 409 — конфликт состояний, 500 — ошибка сервера. Ответ «успех» приходит,
только если данные в БД действительно изменились. Услуги в статусе `deleted`
клиенту не передаются.

### Домен «услуга» — `/api/import_categories`

Доступ: **гость** — без токена, **создатель** — нужен `Authorization: Bearer <JWT>`.

| № | Метод | Адрес | Доступ | Что делает | Тело запроса | Ответ `data` |
|---|---|---|---|---|---|---|
| 1 | GET | `/api/import_categories?date_start=0500-01-01` | гость | список **опубликованных** услуг; фильтр — дата начала не позднее указанной | — | массив; `is_creator` = 1, если создатель услуги — текущий пользователь |
| 2 | GET | `/api/import_categories/feed` | гость | лента: первая опубликованная услуга | — | услуга; `is_liked` = 1, если текущий пользователь её лайкнул; `creator`, `position`, `total`, `next_id` |
| | GET | `/api/import_categories/feed?id=3&next=true` | гость | лента по ид: с `next=true` — следующая после услуги 3 (после последней — первая) | — | то же |
| 3 | GET | `/api/import_categories/draft` | создатель | черновик текущего пользователя (не больше одного), ид не указывается | — | черновик / 404 |
| 4 | POST | `/api/import_categories` | создатель | добавление: создаёт черновик, файлы кладёт в MinIO | `multipart/form-data`: `title` (обяз.), `image` и `video` — файлы | 201, черновик с именами файлов и ссылками / 409, если черновик уже есть |
| 5 | PUT | `/api/import_categories/:id/publish` | создатель | публикация: черновик → опубликована | JSON: `date_start`, `date_end` (обяз.), `description` | опубликованная услуга |
| 6 | DELETE | `/api/import_categories/:id` | создатель | мягкое удаление (статус `deleted`), только услуги текущего пользователя | — | `{"id": 17, "status": "deleted"}` |
| 7 | POST | `/api/import_categories/:id/like` | создатель | лайк от текущего пользователя | JSON: `{"like": 1}` ставит, `{"like": 0}` отменяет | `is_liked`, `likes` / 409 при повторе |

### Домен «пользователь» — `/api/users`

| № | Метод | Адрес | Доступ | Что делает | Тело запроса | Ответ `data` |
|---|---|---|---|---|---|---|
| 8 | POST | `/api/users/register` | гость | регистрация, пароль хранится хешем bcrypt | JSON: `login` (3–64), `password` (от 6) | 201, пользователь без пароля / 409, логин занят |
| 9 | POST | `/api/users/login` | гость | аутентификация: выдаёт JWT | JSON: `login`, `password` | `access_token`, `token_type`, `expires_in`, `user` / 401 |
| 10 | POST | `/api/users/logout` | создатель | деавторизация: токен в чёрный список Redis | — | 200 / 401 |

### Бизнес-правила

* **Статусы** меняются только так: создание → `draft`; `PUT …/publish` → `published`;
  `DELETE` → `deleted`. У создателя два разных метода — опубликовать и удалить.
  Вернуть в черновик нельзя: такого метода нет, а публикация меняет только запись
  в статусе `draft`. Повторная публикация, публикация чужой или удалённой услуги — 404.
* **Системные поля** — `id`, `status`, `creator_id`, `is_moderator` — с клиента не принимаются:
  их нет в сериализаторах запросов, а JSON с неизвестным полем отклоняется (400).
  В форме создания такие поля тоже дают 400. Имена файлов генерирует сервер.
* **Один черновик** на пользователя: проверка в коде и частичный уникальный индекс в БД.
* **Файлы.** Фото — JPG / PNG / WEBP до 10 МБ, видео — MP4 / WEBM до 50 МБ. Тип проверяется
  по содержимому (`http.DetectContentType`). Имя генерируется на латинице из названия:
  «Книдские амфоры» → `knidskie-amfory-3f9a0c.jpg`. В поля `image` / `video` пишется имя,
  файл кладётся в бакет `import-categories` в папку `img/` или `video/`, ссылку собирает
  сериализатор. Если запись в БД не создалась, загруженные файлы удаляются.
* **Даты** — годы до н. э. хранятся положительным числом (`0620-01-01` = 620 г. до н. э.),
  поэтому дата начала должна быть не меньше даты конца, а список упорядочен по убыванию даты.

---

## 4. Таблицы БД

ER-диаграмма — [docs/er.png](docs/er.png), исходник — [docs/er.drawio](docs/er.drawio).

**users**

| Поле | Тип | Ограничения |
|---|---|---|
| id | bigint | PK |
| login | varchar(64) | NOT NULL, UNIQUE |
| password | varchar(128) | NOT NULL |
| is_moderator | boolean | NOT NULL, DEFAULT false |

**import_categories**

| Поле | Тип | Ограничения | Кто заполняет |
|---|---|---|---|
| id | bigint | PK | сервер |
| title | varchar(128) | NOT NULL | клиент, POST |
| image | varchar(256) | NOT NULL, DEFAULT '' | сервер: имя загруженного файла (нет файла — '') |
| video | varchar(256) | NOT NULL, DEFAULT '' | сервер: имя загруженного файла (нет файла — '') |
| date_start | date | NULL, INDEX | клиент, PUT publish |
| date_end | date | NULL | клиент, PUT publish |
| description | text | NULL | клиент, PUT publish |
| status | varchar(16) | NOT NULL: `draft` / `published` / `deleted` | сервер |
| formed_at | timestamp | NULL | сервер, PUT publish: момент публикации |
| creator_id | bigint | NOT NULL, FK → users.id | сервер (singleton) |

Частичный уникальный индекс `uniq_draft_per_user (creator_id) WHERE status = 'draft'`.

**likes** — связь «многие ко многим»

| Поле | Тип | Ограничения |
|---|---|---|
| id | bigint | PK |
| user_id | bigint | NOT NULL, FK → users.id |
| import_category_id | bigint | NOT NULL, FK → import_categories.id |

Пара `(user_id, import_category_id)` уникальна.

---

## 5. Структура кода

```
cmd/main/main.go                        точка входа
cmd/migrate/main.go                     миграции и начальные данные
internal/api/server.go                  сборка: PostgreSQL, MinIO, Redis, маршруты
internal/app/auth/                      JWT (jwt.go), bcrypt (password.go), пользователь из контекста (current_user.go)
internal/app/redis/redis.go             чёрный список JWT в Redis
internal/app/handler/middleware.go      middleware WithAuthCheck: гость / создатель
internal/app/ds/                        модели (таблицы): ImportCategory, User, Like
internal/app/serializer/                сериализаторы ответов и запросов
internal/app/repository/                доступ к данным: GORM (услуги, пользователи) и MinIO (файлы)
internal/app/handler/                   контроллеры: домены «услуга» и «пользователь»
deployments/docker-compose.yml          PostgreSQL, Adminer, MinIO, Redis
docs/swagger/                           документация Swagger (swag init)
docs/                                   ER-диаграмма, диаграммы классов и последовательности, коллекция Postman
```

* **Модели** (`ds`) описывают таблицы: теги `gorm` — колонки и связи, теги `json` — имена
  полей; пароль не сериализуется (`json:"-"`).
* **Сериализаторы** (`serializer`) описывают, что уходит клиенту и что приходит от него:
  ответы без пароля, с готовыми ссылками на файлы, признаками 0/1 и вложенным создателем
  (`creator` — вложенная сериализация); запросы — без системных полей, с проверкой
  `binding:"required"`.
* **Домены** — интерфейсы `ImportCategoriesDomain` и `UsersDomain` в
  [internal/app/handler/handler.go](internal/app/handler/handler.go).

Диаграмма классов (страницы → домены → модели → таблицы): [docs/classes.png](docs/classes.png),
исходник — [docs/classes.drawio](docs/classes.drawio).

---

## 6. Порядок показа ЛР4

1. **Скриншоты 1–3** — Swagger (<http://localhost:8080/swagger/index.html>) в режиме
   инкогнито: `POST /api/users/login` → скопировать `access_token`; `GET /api/import_categories`
   (работает без авторизации). Токен из ответа — в Postman: *Authorization → Bearer Token*
   (коллекция сохраняет его сама после запроса «9. Аутентификация»).
2. **Скриншоты 4–5** — Postman: `GET` списка услуг с токеном и без — работает в обоих случаях.
3. **Скриншоты 6–8** — `PUT …/publish`: папка «ЛР4: авторизация» → «Публикация гостем → 401»;
   затем с токеном — 200, поля обновлены (`SELECT` в Adminer). Содержимое Redis после выхода:
   `docker exec -it rip-redis redis-cli -a rip` → `KEYS *`, `GET <ключ>` (логин пользователя).
4. **Скриншот 9** — диаграмма последовательности [docs/sequence.drawio](docs/sequence.drawio)
   ([png](docs/sequence.png)): сообщения — методы доменов из диаграммы классов с аргументами.
   Все три диаграммы (ER, классы, последовательность) — в одном файле
   [docs/amphora_diagrams.drawio](docs/amphora_diagrams.drawio), каждая на своём листе.

Ответы на контрольные вопросы — [docs/questions.md](docs/questions.md).
