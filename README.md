# Амфора — лабораторная работа №3 (веб-сервис)

Курс «Разработка интернет-приложений», МГТУ им. Н. Э. Баумана, ИУ5.

**Тема.** История и культура: оценка торговых связей по статистике импортных находок.

* **Услуга** — категория импортных находок. В коде, адресах и БД — **`import_categories`**.
* **Поля по теме** — дата начала и дата конца бытования типа (`date_start`, `date_end`).
  Фильтрация списка — по дате начала.

В ЛР3 приложение стало веб-сервисом (REST API) для будущего SPA: HTML-шаблоны убраны,
все ответы — JSON, все адреса начинаются с `/api`. Данные — PostgreSQL через ORM GORM,
изображения и видео услуг — файлы в MinIO.

---

## 1. Запуск

Нужны Go 1.21+ и Docker.

```bash
go mod tidy                                             # зависимости: gin, gorm, minio-go
docker compose -f deployments/docker-compose.yml up -d  # PostgreSQL, Adminer, MinIO
go run ./cmd/migrate -drop                              # таблицы + начальные данные
go run ./cmd/main                                       # API: http://localhost:8080/api
```

| Адрес | Что это |
|---|---|
| <http://localhost:8080/api/import_categories> | API (список услуг) |
| <http://localhost:8081> | Adminer: сервер `postgres`, база `amphora`, `rip` / `rip` |
| <http://localhost:9001> | консоль MinIO, `minioadmin` / `minioadmin`, бакет `import-categories` |

Коллекция запросов для Postman / Insomnia —
[`docs/postman/amphora.postman_collection.json`](docs/postman/amphora.postman_collection.json)
(Postman: *Import*; Insomnia: *Import → From File*). В ней папка «Основные запросы» —
10 методов API в порядке показа, и папка «Проверка ошибок» — запросы с плохими ответами.

---

## 2. Текущий пользователь — функция-singleton

Авторизация появится в ЛР4. До этого пользователь-создатель зафиксирован константой
и выдаётся функцией-singleton `auth.CurrentUserID()`
([internal/app/auth/current_user.go](internal/app/auth/current_user.go)):

```go
const creatorID uint = 1 // пользователь umid

var (
	once          sync.Once
	currentUserID uint
)

func CurrentUserID() uint {
	once.Do(func() { currentUserID = creatorID })
	return currentUserID
}
```

Функция вызывается во всех методах домена «услуга»
([internal/app/handler/import_categories.go](internal/app/handler/import_categories.go)):
признак `is_creator` в списке, `is_liked` в ленте, черновик, создание, публикация,
удаление и лайк.

---

## 3. HTTP-методы

Ответ всегда JSON одного вида:

```json
{"status": "success", "data": { ... }, "message": "черновик создан"}
{"status": "fail", "message": "у пользователя уже есть черновик"}
```

Коды: 200 — успех, 201 — создано, 400 — неверные данные, 404 — записи нет или она
недоступна, 409 — конфликт состояний, 500 — ошибка сервера. Ответ «успех» приходит,
только если данные в БД действительно изменились. Услуги в статусе `deleted`
клиенту не передаются.

### Домен «услуга» — `/api/import_categories`

| № | Метод | Адрес | Что делает | Тело запроса | Ответ `data` |
|---|---|---|---|---|---|
| 1 | GET | `/api/import_categories?date_start=0500-01-01` | список **опубликованных** услуг; фильтр — дата начала не позднее указанной | — | массив; `is_creator` = 1, если создатель услуги — текущий пользователь |
| 2 | GET | `/api/import_categories/feed` | лента: первая опубликованная услуга | — | услуга; `is_liked` = 1, если текущий пользователь её лайкнул; `creator`, `position`, `total`, `next_id` |
| | GET | `/api/import_categories/feed?id=3&next=true` | лента по ид: с `next=true` — следующая после услуги 3 (после последней — первая) | — | то же |
| 3 | GET | `/api/import_categories/draft` | черновик текущего пользователя (не больше одного), ид не указывается | — | черновик / 404 |
| 4 | POST | `/api/import_categories` | добавление: создаёт черновик, файлы кладёт в MinIO | `multipart/form-data`: `title` (обяз.), `image` и `video` — файлы | 201, черновик с именами файлов и ссылками / 409, если черновик уже есть |
| 5 | PUT | `/api/import_categories/:id/publish` | публикация: черновик → опубликована | JSON: `date_start`, `date_end` (обяз.), `description` | опубликованная услуга |
| 6 | DELETE | `/api/import_categories/:id` | мягкое удаление (статус `deleted`), только услуги текущего пользователя | — | `{"id": 17, "status": "deleted"}` |
| 7 | POST | `/api/import_categories/:id/like` | лайк от текущего пользователя | JSON: `{"like": 1}` ставит, `{"like": 0}` отменяет | `is_liked`, `likes` / 409 при повторе |

### Домен «пользователь» — `/api/users`

| № | Метод | Адрес | Что делает | Тело запроса | Ответ `data` |
|---|---|---|---|---|---|
| 8 | POST | `/api/users/register` | регистрация | JSON: `login` (3–64), `password` (от 6) | 201, пользователь без пароля / 409, логин занят |
| 9 | POST | `/api/users/login` | аутентификация — заглушка до ЛР4 | — | — |
| 10 | POST | `/api/users/logout` | деавторизация — заглушка до ЛР4 | — | — |

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
| image | varchar(256) | NULL | сервер: имя загруженного файла |
| video | varchar(256) | NULL | сервер: имя загруженного файла |
| date_start | date | NULL, INDEX | клиент, PUT publish |
| date_end | date | NULL | клиент, PUT publish |
| description | text | NULL | клиент, PUT publish |
| status | varchar(16) | NOT NULL: `draft` / `published` / `deleted` | сервер |
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
internal/api/server.go                  сборка: PostgreSQL, MinIO, маршруты
internal/app/auth/current_user.go       функция-singleton текущего пользователя
internal/app/ds/                        модели (таблицы): ImportCategory, User, Like
internal/app/serializer/                сериализаторы ответов и запросов
internal/app/repository/                доступ к данным: GORM (услуги, пользователи) и MinIO (файлы)
internal/app/handler/                   контроллеры: домены «услуга» и «пользователь»
deployments/docker-compose.yml          PostgreSQL, Adminer, MinIO
docs/                                   ER-диаграмма, диаграмма классов, коллекция Postman
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

## 6. Порядок показа ЛР3

1. **Скриншоты 1–10** — Postman/Insomnia, папка «Основные запросы»: список с фильтром,
   добавление с фото и видео, черновик, публикация, лента без ид, лента `?id=…&next=true`,
   лайк, удаление, регистрация.
2. **Скриншоты 11–13** — изменённые данные в Adminer:
   ```sql
   SELECT id, title, image, video, date_start, date_end, status, creator_id
     FROM import_categories ORDER BY id DESC;           -- новая услуга: published → deleted
   SELECT * FROM likes ORDER BY id DESC;                -- лайк текущего пользователя
   SELECT id, login, password, is_moderator FROM users ORDER BY id DESC;  -- новый пользователь
   ```
   Файлы с латинскими именами видны в консоли MinIO (бакет `import-categories`, папки `img`, `video`).
3. **Скриншоты 14–15** — модели (`internal/app/ds`) и сериализаторы (`internal/app/serializer`).
4. **Скриншоты 16–17** — функция-singleton и её вызовы, этот README.

Ответы на контрольные вопросы — [docs/questions.md](docs/questions.md).
