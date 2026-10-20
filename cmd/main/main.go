// Приложение «Амфора» — каталог импортных находок и оценка торговых связей.
// Веб-сервис (REST API) курса «Разработка интернет-приложений», ИУ5.
//
// Описание для Swagger (swaggo/swag). После изменения комментариев
// документация пересобирается командой:
//
//	swag init -g cmd/main/main.go -o docs/swagger
package main

import (
	"log"

	_ "rip/lab1/docs/swagger" // сгенерированная документация Swagger

	"rip/lab1/internal/api"
)

//	@title			Амфора API
//	@version		1.0
//	@description	Оценка торговых связей по статистике импортных находок: категории импорта (услуги), черновик, публикация, лента и лайки.
//	@description	Гостю доступно только чтение (список и лента). Методы создателя требуют JWT: выполните POST /api/users/login, нажмите Authorize и вставьте access_token.

//	@contact.name	Солижонов Умид, ИУ5

//	@host		localhost:8080
//	@BasePath	/
//	@schemes	http

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				JWT из ответа /api/users/login. Можно вставить с префиксом «Bearer » или без него.

func main() {
	log.Println("Приложение запускается")
	api.StartServer()
	log.Println("Приложение завершило работу")
}
