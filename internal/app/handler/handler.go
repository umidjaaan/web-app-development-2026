// Package handler — слой контроллеров (View в терминах MVT).
// Каждый обработчик получает данные из репозитория и передаёт их
// в шаблон через gin.H.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

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
	router.GET("/", h.Index)              // редирект на первую карточку ленты
	router.GET("/feed", h.Index)          // редирект на первую карточку ленты
	router.GET("/feed/:id", h.FeedPage)   // GET №1 — лента по ID услуги
	router.GET("/add", h.AddPage)         // GET №2 — получение черновика
	router.GET("/catalog", h.CatalogPage) // GET №3 — список всех карточек + фильтр
}

// Index перенаправляет на первую услугу ленты.
func (h *Handler) Index(c *gin.Context) {
	id := h.Repository.FirstPublishedID()
	c.Redirect(http.StatusFound, "/feed/"+strconv.Itoa(id))
}

// FeedPage — вкладка «Лента»: вертикальная развёртка с автопроигрыванием
// видео и информацией поверх него. Услуга запрашивается по ID.
func (h *Handler) FeedPage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.Logger.Warnf("некорректный ID услуги: %q", c.Param("id"))
		h.renderNotFound(c, "Некорректный идентификатор услуги")
		return
	}

	category, err := h.Repository.GetByID(id)
	if err != nil {
		h.Logger.Warnf("услуга %d не найдена или недоступна", id)
		h.renderNotFound(c, "Услуга не найдена или снята с публикации")
		return
	}

	position, total := h.Repository.Position(id)
	h.Logger.Infof("GET /feed/%d — %s", id, category.Title)

	c.HTML(http.StatusOK, "feed.html", gin.H{
		"Tab":      "feed",
		"Site":     h.Repository.Site(),
		"Category": category,
		"Share":    h.Repository.ShareOf(category),
		"Position": position,
		"Total":    total,
		"NextID":   h.Repository.NextID(id),
	})
}

// AddPage — вкладка «Добавление»: форма с раздельными полями для
// изображения, видео, текста и параметров услуги.
// Форма заполняется данными черновика; сохранение появится в ЛР3.
func (h *Handler) AddPage(c *gin.Context) {
	draft, err := h.Repository.GetDraft()
	if err != nil {
		h.Logger.Warn("черновик не найден")
	}
	h.Logger.Infof("GET /add — черновик %d", draft.ID)

	c.HTML(http.StatusOK, "add.html", gin.H{
		"Tab":      "add",
		"Site":     h.Repository.Site(),
		"Draft":    draft,
		"HasDraft": err == nil,
		"Regions":  h.Repository.Regions(),
	})
}

// CatalogPage — вкладка «Плитка»: двухколончатый список карточек
// с фильтрацией на сервере по числу фрагментов.
func (h *Handler) CatalogPage(c *gin.Context) {
	raw := c.Query("min_finds")
	minFinds, err := strconv.Atoi(raw)
	if err != nil || minFinds < 0 {
		minFinds = 0
	}

	categories := h.Repository.GetPublished(minFinds)
	h.Logger.Infof("GET /catalog?min_finds=%d — найдено %d услуг", minFinds, len(categories))

	c.HTML(http.StatusOK, "catalog.html", gin.H{
		"Tab":         "catalog",
		"Site":        h.Repository.Site(),
		"Categories":  categories,
		"MinFinds":    minFinds,
		"MinFindsRaw": raw,
		"Shares":      h.Repository.RegionShares(),
		"MaxShare":    h.Repository.MaxRegionPercent(),
		"TotalFinds":  h.Repository.TotalFinds(),
		"Found":       len(categories),
	})
}

func (h *Handler) renderNotFound(c *gin.Context, message string) {
	c.HTML(http.StatusNotFound, "error.html", gin.H{
		"Tab":     "feed",
		"Site":    h.Repository.Site(),
		"Message": message,
	})
}
