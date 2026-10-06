// Package handler — слой контроллеров (View в терминах MVT).
// Каждый обработчик получает данные из репозитория и передаёт их
// в шаблон через gin.H.
package handler

import (
	"net/http"
	"strconv"
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
}

// New создаёт обработчик.
func New(repo *repository.Repository, logger *logrus.Logger) *Handler {
	return &Handler{Repository: repo, Logger: logger}
}

// RegisterHandler описывает маршрутизацию приложения:
// три URL — три контроллера — три страницы.
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/", h.Index)

	// GET №1 — список категорий импорта с фильтрацией по дате начала.
	router.GET("/import_categories", h.ImportCategoryList)
	// GET №2 — форма добавления категории импорта.
	router.GET("/import_categories/add", h.ImportCategoryAdd)
	// GET №3 — лента: одна категория импорта по идентификатору.
	router.GET("/import_categories/feed", h.Index)
	router.GET("/import_categories/feed/:id", h.ImportCategoryFeed)
}

// Index перенаправляет на первую категорию ленты.
func (h *Handler) Index(c *gin.Context) {
	id := h.Repository.FirstImportCategoryID()
	c.Redirect(http.StatusFound, "/import_categories/feed/"+strconv.Itoa(id))
}

// ImportCategoryFeed — вкладка «Лента»: вертикальная развёртка
// с автопроигрыванием видео и информацией поверх него.
// Категория импорта запрашивается по ID.
func (h *Handler) ImportCategoryFeed(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.Logger.Warnf("некорректный ID категории импорта: %q", c.Param("id"))
		h.renderNotFound(c, "Некорректный идентификатор категории импорта")
		return
	}

	importCategory, err := h.Repository.ImportCategoryByID(id)
	if err != nil {
		h.Logger.Warnf("категория импорта %d не найдена или удалена", id)
		h.renderNotFound(c, "Категория импорта не найдена или удалена")
		return
	}

	position, total := h.Repository.Position(id)
	h.Logger.Infof("GET /import_categories/feed/%d — %s", id, importCategory.Title)

	c.HTML(http.StatusOK, "import_category.html", gin.H{
		"Tab":            "feed",
		"Site":           h.Repository.Site(),
		"ImportCategory": importCategory,
		"Share":          h.Repository.ShareOf(importCategory),
		"Position":       position,
		"Total":          total,
		"NextID":         h.Repository.NextImportCategoryID(id),
	})
}

// ImportCategoryAdd — вкладка «Добавление»: пустая форма с раздельными
// полями для изображения, видео, текста и параметров категории.
// Сохранение появится в ЛР3.
func (h *Handler) ImportCategoryAdd(c *gin.Context) {
	h.Logger.Info("GET /import_categories/add — форма добавления")

	c.HTML(http.StatusOK, "import_category_add.html", gin.H{
		"Tab":     "add",
		"Site":    h.Repository.Site(),
		"Regions": h.Repository.Regions(),
	})
}

// ImportCategoryList — вкладка «Плитка»: двухколончатый список карточек
// с фильтрацией на сервере по дате начала бытования типа.
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

	importCategories := h.Repository.ImportCategories(dateStart)
	h.Logger.Infof("GET /import_categories?date_start=%s — найдено %d категорий",
		raw, len(importCategories))

	c.HTML(http.StatusOK, "import_categories.html", gin.H{
		"Tab":              "catalog",
		"Site":             h.Repository.Site(),
		"ImportCategories": importCategories,
		"DateStart":        raw,
		"HasFilter":        raw != "",
		"Shares":           h.Repository.RegionShares(),
		"MaxShare":         h.Repository.MaxRegionPercent(),
		"TotalFinds":       h.Repository.TotalFinds(),
		"Found":            len(importCategories),
	})
}

func (h *Handler) renderNotFound(c *gin.Context, message string) {
	c.HTML(http.StatusNotFound, "error.html", gin.H{
		"Tab":     "feed",
		"Site":    h.Repository.Site(),
		"Message": message,
	})
}
