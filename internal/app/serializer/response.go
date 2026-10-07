// Package serializer — сериализаторы API.
//
// Модели ds описывают таблицы БД. Сериализаторы описывают, что уходит
// клиенту (ответы) и что приходит от него (запросы):
//   - в ответах нет пароля и есть готовые ссылки на файлы и признаки 0/1;
//   - в запросах нет системных полей (id, status, creator_id, is_moderator) —
//     их задаёт бэкенд, передать их с клиента нельзя.
package serializer

// Response — общий вид ответа API (как в методичке):
//
//	успех:  {"status": "success", "data": ..., "message": "..."}
//	ошибка: {"status": "fail", "message": "..."}
type Response struct {
	Status  string `json:"status"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

// Success — ответ с данными.
func Success(data any, message string) Response {
	return Response{Status: "success", Data: data, Message: message}
}

// Fail — ответ с ошибкой.
func Fail(message string) Response {
	return Response{Status: "fail", Message: message}
}

// flag — признак 0/1 для клиента.
func flag(v bool) int {
	if v {
		return 1
	}
	return 0
}
