// Package handler — контроллеры REST API. Два домена:
//   - «услуга»       — /api/import_categories (ImportCategoriesDomain);
//   - «пользователь» — /api/users            (UsersDomain).
//
// Все ответы — JSON вида {"status": "success" | "fail", "data": ..., "message": ...}.
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"rip/lab1/internal/app/repository"
	"rip/lab1/internal/app/serializer"
)

// ImportCategoriesDomain — домен «услуга», адреса /api/import_categories.
type ImportCategoriesDomain interface {
	GetImportCategories(c *gin.Context)   // GET    /api/import_categories?date_start=
	GetFeed(c *gin.Context)               // GET    /api/import_categories/feed?id=&next=true
	GetDraft(c *gin.Context)              // GET    /api/import_categories/draft
	CreateImportCategory(c *gin.Context)  // POST   /api/import_categories
	PublishImportCategory(c *gin.Context) // PUT    /api/import_categories/:id/publish
	DeleteImportCategory(c *gin.Context)  // DELETE /api/import_categories/:id
	LikeImportCategory(c *gin.Context)    // POST   /api/import_categories/:id/like
}

// UsersDomain — домен «пользователь», адреса /api/users.
type UsersDomain interface {
	Register(c *gin.Context) // POST /api/users/register
	Login(c *gin.Context)    // POST /api/users/login  (заглушка до ЛР4)
	Logout(c *gin.Context)   // POST /api/users/logout (заглушка до ЛР4)
}

// Handler реализует оба домена.
type Handler struct {
	Repository *repository.Repository
	Logger     *logrus.Logger
}

// Проверка на этапе компиляции: Handler реализует интерфейсы доменов.
var (
	_ ImportCategoriesDomain = (*Handler)(nil)
	_ UsersDomain            = (*Handler)(nil)
)

// New создаёт обработчик.
func New(repo *repository.Repository, logger *logrus.Logger) *Handler {
	return &Handler{Repository: repo, Logger: logger}
}

// RegisterHandler описывает маршруты API. Все адреса начинаются с /api.
func (h *Handler) RegisterHandler(router *gin.Engine) {
	api := router.Group("/api")

	categories := api.Group("/import_categories")
	categories.GET("", h.GetImportCategories)
	categories.GET("/feed", h.GetFeed)
	categories.GET("/draft", h.GetDraft)
	categories.POST("", h.CreateImportCategory)
	categories.PUT("/:id/publish", h.PublishImportCategory)
	categories.DELETE("/:id", h.DeleteImportCategory)
	categories.POST("/:id/like", h.LikeImportCategory)

	users := api.Group("/users")
	users.POST("/register", h.Register)
	users.POST("/login", h.Login)
	users.POST("/logout", h.Logout)
}

// ---------------------------------------------------------------- ответы

func (h *Handler) success(c *gin.Context, code int, data any, message string) {
	c.JSON(code, serializer.Success(data, message))
}

func (h *Handler) fail(c *gin.Context, code int, err error) {
	h.Logger.Warnf("%s %s → %d: %v", c.Request.Method, c.Request.URL.Path, code, err)
	c.JSON(code, serializer.Fail(err.Error()))
}

// failRepo переводит ошибки бизнес-логики в HTTP-коды:
// плохой файл — 400, нет записи — 404, конфликт состояний — 409, остальное — 500.
func (h *Handler) failRepo(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrBadFile):
		h.fail(c, http.StatusBadRequest, err)
	case errors.Is(err, repository.ErrNotFound), errors.Is(err, repository.ErrNotYours),
		errors.Is(err, repository.ErrNoDraft):
		h.fail(c, http.StatusNotFound, err)
	case errors.Is(err, repository.ErrDraftExists), errors.Is(err, repository.ErrLiked),
		errors.Is(err, repository.ErrNotLiked), errors.Is(err, repository.ErrLoginTaken):
		h.fail(c, http.StatusConflict, err)
	default:
		h.fail(c, http.StatusInternalServerError, err)
	}
}

// parseID разбирает ид из адреса: целое число больше нуля.
func parseID(raw string) (uint, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("ид должен быть целым числом больше 0")
	}
	return uint(id), nil
}
