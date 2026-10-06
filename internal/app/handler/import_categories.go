// Package handler — слой контроллеров (View в терминах MVT).
// Каждый обработчик получает данные из репозитория и передаёт их
// в шаблон через gin.H.
//
// Шесть методов: три GET и два POST через ORM (кнопки «Далее»
// и «Опубликовать» — два статуса: черновик и опубликована) и один POST
// логического удаления — SQL UPDATE через курсор.
package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"rip/lab1/internal/app/ds"
	"rip/lab1/internal/app/repository"
)

// Handler связывает маршруты приложения с репозиторием.
type Handler struct {
	Repository *repository.Repository
	Logger     *logrus.Logger
	// CurrentUserID — пользователь, от имени которого работает приложение.
	// Авторизация появится в ЛР4.
	CurrentUserID uint
}

// New создаёт обработчик.
func New(repo *repository.Repository, logger *logrus.Logger, currentUserID uint) *Handler {
	return &Handler{Repository: repo, Logger: logger, CurrentUserID: currentUserID}
}

// RegisterHandler описывает маршруты приложения.
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/", h.Index)
	router.GET("/import_categories/feed", h.Index)

	// GET №1 — плитка: опубликованные категории, фильтр по дате начала.
	router.GET("/import_categories", h.ImportCategoryList)
	// GET №2 — лента: одна категория по идентификатору.
	router.GET("/import_categories/feed/:id", h.ImportCategoryFeed)
	// GET №3 — добавление: шаг 1 (фото, видео, название) или черновик.
	router.GET("/import_categories/add", h.ImportCategoryAdd)

	// POST №1 — кнопка «Далее»: создание черновика (ORM).
	router.POST("/import_categories", h.ImportCategoryCreate)
	// POST №2 — кнопка «Опубликовать»: черновик → опубликована (ORM).
	router.POST("/import_categories/:id/publish", h.ImportCategoryPublish)
	// POST №3 — удаление из плитки: SQL UPDATE через курсор.
	router.POST("/import_categories/:id/delete", h.ImportCategoryDelete)
}

// Index перенаправляет на первую категорию ленты.
func (h *Handler) Index(c *gin.Context) {
	id := h.Repository.FirstImportCategoryID()
	c.Redirect(http.StatusFound, "/import_categories/feed/"+strconv.FormatUint(uint64(id), 10))
}

// ==========================================================================
// GET №1 — плитка с фильтром по дате начала
// ==========================================================================

func (h *Handler) ImportCategoryList(c *gin.Context) {
	raw := c.Query("date_start")

	var dateStart time.Time
	if raw != "" {
		parsed, err := time.Parse(ds.DateLayout, raw)
		if err != nil {
			h.Logger.Warnf("некорректная дата фильтра: %q", raw)
			raw = ""
		} else {
			dateStart = parsed
		}
	}

	items, err := h.Repository.ImportCategories(dateStart)
	if err != nil {
		h.renderError(c, http.StatusInternalServerError, "Не удалось получить список категорий")
		return
	}

	h.Logger.Infof("GET /import_categories?date_start=%s — найдено %d", raw, len(items))

	c.HTML(http.StatusOK, "import_categories.html", h.page(gin.H{
		"Tab":              "catalog",
		"ImportCategories": items,
		"DateStart":        raw,
		"HasFilter":        raw != "",
		"Found":            len(items),
	}))
}

// ==========================================================================
// GET №2 — лента: одна категория по идентификатору
// ==========================================================================

func (h *Handler) ImportCategoryFeed(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		h.renderError(c, http.StatusNotFound, "Некорректный идентификатор категории импорта")
		return
	}

	item, err := h.Repository.ImportCategoryByID(id)
	if err != nil {
		h.Logger.Warnf("категория импорта %d не найдена или удалена", id)
		h.renderError(c, http.StatusNotFound, "Категория импорта не найдена или удалена")
		return
	}

	position, total := h.Repository.Position(id)
	h.Logger.Infof("GET /import_categories/feed/%d — %s", id, item.Title)

	c.HTML(http.StatusOK, "import_category.html", h.page(gin.H{
		"Tab":            "feed",
		"ImportCategory": item,
		"Position":       position,
		"Total":          total,
		"NextID":         h.Repository.NextImportCategoryID(id),
	}))
}

// ==========================================================================
// GET №3 — добавление: шаг 1 или черновик (шаг 2)
// ==========================================================================

func (h *Handler) ImportCategoryAdd(c *gin.Context) {
	data := gin.H{"Tab": "add"}

	// Если у пользователя уже есть черновик — показываем шаг 2:
	// фото и видео сверху, остальные поля и кнопка «Опубликовать».
	if draft, err := h.Repository.DraftByUser(h.CurrentUserID); err == nil {
		data["Draft"] = draft
		h.Logger.Infof("GET /import_categories/add — черновик №%d", draft.ID)
	} else {
		h.Logger.Info("GET /import_categories/add — шаг 1")
	}

	c.HTML(http.StatusOK, "import_category_add.html", h.page(data))
}

// ==========================================================================
// POST №1 — кнопка «Далее»: черновик (ORM)
// ==========================================================================

func (h *Handler) ImportCategoryCreate(c *gin.Context) {
	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		h.renderError(c, http.StatusBadRequest, "Название категории обязательно")
		return
	}

	// По кнопке «Далее» заполняются только название, фото и видео.
	item := ds.ImportCategory{
		Title:     title,
		CreatorID: h.CurrentUserID,
	}

	// Фото и видео, выбранные в проводнике. Если файл не выбран,
	// ссылка остаётся пустой и в HTML подставится заглушка.
	var err error
	if item.ImageURL, err = saveUpload(c, "image", imageExts); err != nil {
		h.renderError(c, http.StatusBadRequest, err.Error())
		return
	}
	if item.VideoURL, err = saveUpload(c, "video", videoExts); err != nil {
		h.renderError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Repository.CreateDraft(&item); err != nil {
		h.Logger.Errorf("создание черновика: %v", err)
		h.renderError(c, http.StatusInternalServerError,
			"Не удалось создать черновик: у пользователя уже есть черновик")
		return
	}

	h.Logger.Infof("POST /import_categories — создан черновик №%d", item.ID)
	c.Redirect(http.StatusSeeOther, "/import_categories/add")
}

// ==========================================================================
// POST №2 — кнопка «Опубликовать»: черновик → опубликована (ORM)
// ==========================================================================

func (h *Handler) ImportCategoryPublish(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		h.renderError(c, http.StatusNotFound, "Некорректный идентификатор категории импорта")
		return
	}

	dateStart, err1 := time.Parse(ds.DateLayout, c.PostForm("date_start"))
	dateEnd, err2 := time.Parse(ds.DateLayout, c.PostForm("date_end"))
	if err1 != nil || err2 != nil {
		h.renderError(c, http.StatusBadRequest, "Укажите дату начала и дату конца")
		return
	}
	// Перед публикацией заполняются два поля по теме и описание.
	fields := map[string]any{
		"date_start":  dateStart,
		"date_end":    dateEnd,
		"description": c.PostForm("description"),
	}

	if err := h.Repository.PublishDraft(id, h.CurrentUserID, fields); err != nil {
		h.Logger.Warnf("публикация черновика %d: %v", id, err)
		h.renderError(c, http.StatusNotFound, "Черновик не найден")
		return
	}

	h.Logger.Infof("POST /import_categories/%d/publish — опубликована", id)
	c.Redirect(http.StatusSeeOther, "/import_categories/feed/"+strconv.FormatUint(uint64(id), 10))
}

// ==========================================================================
// POST №3 — удаление из плитки: SQL UPDATE через курсор
// ==========================================================================

func (h *Handler) ImportCategoryDelete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		h.renderError(c, http.StatusNotFound, "Некорректный идентификатор категории импорта")
		return
	}

	if err := h.Repository.DeleteImportCategory(id); err != nil {
		h.Logger.Warnf("удаление категории %d: %v", id, err)
		h.renderError(c, http.StatusNotFound, "Категория импорта не найдена")
		return
	}

	h.Logger.Infof("POST /import_categories/%d/delete — логически удалена", id)
	c.Redirect(http.StatusSeeOther, "/import_categories")
}

// ==========================================================================
// Вспомогательное
// ==========================================================================

// page добавляет к данным шаблона то, что нужно каждой странице:
// памятник, текущего пользователя и число его отметок.
func (h *Handler) page(data gin.H) gin.H {
	data["Site"] = h.Repository.Site()

	if user, err := h.Repository.UserByID(h.CurrentUserID); err == nil {
		data["CurrentUser"] = user
	}
	if likes, err := h.Repository.LikesByUser(h.CurrentUserID); err == nil {
		data["MyLikes"] = likes
	}
	return data
}

func (h *Handler) renderError(c *gin.Context, code int, message string) {
	c.HTML(code, "error.html", h.page(gin.H{
		"Tab":     "feed",
		"Code":    code,
		"Message": message,
	}))
}

func parseID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(id), err
}

// Допустимые расширения загружаемых файлов.
var (
	imageExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	videoExts = map[string]bool{".mp4": true, ".webm": true}
)

// uploadDir — папка для загруженных файлов; отдаётся статикой по /media/uploads.
const uploadDir = "media/uploads"

// saveUpload сохраняет файл из поля формы field и возвращает ссылку на него.
// Пустая строка без ошибки — файл не выбран.
func saveUpload(c *gin.Context, field string, allowed map[string]bool) (string, error) {
	file, err := c.FormFile(field)
	if err != nil {
		return "", nil // поле пустое — файл не выбирали
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowed[ext] {
		return "", fmt.Errorf("недопустимый формат файла %q", file.Filename)
	}

	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return "", err
	}
	name := strconv.FormatInt(time.Now().UnixNano(), 10) + ext
	if err := c.SaveUploadedFile(file, filepath.Join(uploadDir, name)); err != nil {
		return "", err
	}
	return "/media/uploads/" + name, nil
}
