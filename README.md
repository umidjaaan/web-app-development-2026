# Разработка интернет-приложений (РИП), 2026

Приложение «Амфора» — тема «История и культура: оценка торговых связей по статистике импортных находок».

Каждая лабораторная работа лежит в своей ветке:

| Ветка | Лабораторная | Содержание |
|---|---|---|
| [`lab1`](../../tree/lab1) | ЛР1 | Go + Gin, шаблоны (MVT), MinIO, три страницы: лента, добавление, каталог |
| [`lab2`](../../tree/lab2) | ЛР2 | PostgreSQL + GORM + Adminer, таблицы users / import_categories / likes, 3 GET + 3 POST, логическое удаление через SQL |
| [`lab3`](../../tree/lab3) | ЛР3 | REST-веб-сервис на Go + Gin + GORM: JSON API, MinIO для изображений и видео, 10 основных методов, коллекция Postman |
| [`lab4`](../../tree/lab4) | ЛР4 | Авторизация по JWT, чёрный список токенов в Redis, пароли bcrypt, Swagger (swaggo), гость — только чтение |
| [`lab6`](../../tree/lab6) | ЛР6 | `operationId` в Swagger для кодогенерации клиента фронтенда, диаграмма прецедентов, обновлённые диаграммы |

Переключиться на ветку: `git checkout lab6`.

ЛР5 выполняется только во фронтенде, поэтому ветки `lab5` здесь нет.

## Фронтенд

SPA на React + TypeScript — репозиторий
[amphora-frontend](https://github.com/umidjaaan/amphora-frontend):

| Ветка | Лабораторная | Содержание |
|---|---|---|
| [`lab5`](https://github.com/umidjaaan/amphora-frontend/tree/lab5) | ЛР5 | лента и плитка для гостя, React-Bootstrap, mock-данные, изображение и видео по умолчанию |
| [`lab6`](https://github.com/umidjaaan/amphora-frontend/tree/lab6) | ЛР6 | Redux Toolkit, redux-thunk, кодогенерация API из Swagger (axios), вход/выход, добавление, удаление, лайки |
