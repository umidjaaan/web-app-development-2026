package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func init() {
	// Неизвестные поля в JSON — ошибка: так с клиента нельзя передать
	// системные поля (id, status, creator_id, is_moderator).
	binding.EnableDecoderDisallowUnknownFields = true

	// В сообщениях валидатора — имена полей из тегов json.
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(f reflect.StructField) string {
			return strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		})
	}
}

// bindJSON десериализует тело запроса и проверяет теги binding.
// Ошибки переводятся в понятные сообщения на русском.
func bindJSON(c *gin.Context, dst any) error {
	err := c.ShouldBindJSON(dst)
	if err == nil {
		return nil
	}

	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		msgs := make([]string, 0, len(verrs))
		for _, fe := range verrs {
			switch fe.Tag() {
			case "required":
				msgs = append(msgs, fmt.Sprintf("поле %s обязательно", fe.Field()))
			case "oneof":
				msgs = append(msgs, fmt.Sprintf("поле %s: допустимо %s", fe.Field(), strings.ReplaceAll(fe.Param(), " ", " или ")))
			case "min":
				msgs = append(msgs, fmt.Sprintf("поле %s: не короче %s символов", fe.Field(), fe.Param()))
			case "max":
				msgs = append(msgs, fmt.Sprintf("поле %s: не длиннее %s символов", fe.Field(), fe.Param()))
			default:
				msgs = append(msgs, fmt.Sprintf("поле %s заполнено неверно", fe.Field()))
			}
		}
		return errors.New(strings.Join(msgs, "; "))
	}

	if field, ok := strings.CutPrefix(err.Error(), "json: unknown field "); ok {
		return fmt.Errorf("поле %s задаётся на сервере, передавать его нельзя", strings.Trim(field, `"`))
	}
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return fmt.Errorf("поле %s: неверный тип значения (%s)", typeErr.Field, typeErr.Value)
	}
	if errors.Is(err, io.EOF) {
		return errors.New("тело запроса пустое, ожидается JSON")
	}
	return fmt.Errorf("некорректный JSON: %v", err)
}
