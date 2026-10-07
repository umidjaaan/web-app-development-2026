// Приложение «Амфора» — каталог импортных находок и оценка торговых связей.
// Веб-сервис (REST API), ЛР3 курса «Разработка интернет-приложений», ИУ5.
package main

import (
	"log"

	"rip/lab1/internal/api"
)

func main() {
	log.Println("Приложение запускается")
	api.StartServer()
	log.Println("Приложение завершило работу")
}
