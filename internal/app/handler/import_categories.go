package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rip/lab1/internal/app/auth"
	"rip/lab1/internal/app/ds"
	"rip/lab1/internal/app/repository"
	"rip/lab1/internal/app/serializer"
)

// systemFields — поля, которые вычисляются на бэкенде: с клиента их не принимаем.
var systemFields = []string{"id", "status", "creator_id", "formed_at", "is_moderator"}

// maxCreateBody — предел тела запроса на создание: видео 50 МБ + фото 10 МБ + поля.
const maxCreateBody = 61 << 20

// GetImportCategories — GET /api/import_categories?date_start=0500-01-01
// Список опубликованных услуг с фильтром по дате начала.
// is_creator = 1, если создатель услуги — текущий пользователь.
//
//	@Summary		Список услуг
//	@Description	Опубликованные категории импорта с фильтром по дате начала. Доступен гостю; с токеном is_creator = 1 у услуг текущего пользователя.
//	@Tags			import_categories
//	@Produce		json
//	@Param			date_start	query		string	false	"Дата начала, ГГГГ-ММ-ДД (год до н. э.)"	example(0500-01-01)
//	@Success		200			{object}	serializer.Response{data=[]serializer.ImportCategorySerializer}
//	@Failure		400			{object}	serializer.Response
//	@Router			/api/import_categories [get]
func (h *Handler) GetImportCategories(c *gin.Context) {
	var dateStart time.Time
	if raw := c.Query("date_start"); raw != "" {
		parsed, err := time.Parse(ds.DateLayout, raw)
		if err != nil {
			h.fail(c, http.StatusBadRequest, errors.New("date_start: дата в формате ГГГГ-ММ-ДД"))
			return
		}
		dateStart = parsed
	}

	items, err := h.Repository.ImportCategories(dateStart)
	if err != nil {
		h.failRepo(c, err)
		return
	}

	userID := auth.CurrentUserID(c)
	result := make([]serializer.ImportCategorySerializer, 0, len(items))
	for _, item := range items {
		result = append(result, serializer.ToImportCategory(item, userID, h.Repository))
	}
	h.success(c, http.StatusOK, result, "")
}

// GetFeed — лента опубликованных услуг, по одной:
//
//	GET /api/import_categories/feed                 — первая услуга ленты
//	GET /api/import_categories/feed?id=3            — услуга с ид 3
//	GET /api/import_categories/feed?id=3&next=true  — следующая после неё
//
// is_liked = 1, если текущий пользователь лайкнул услугу.
//
//	@Summary		Лента
//	@Description	Одна опубликованная услуга ленты. Без id — первая, с id — эта услуга, с id и next=true — следующая. Доступна гостю; с токеном заполняется is_liked.
//	@Tags			import_categories
//	@Produce		json
//	@Param			id		query		int		false	"Ид услуги"
//	@Param			next	query		bool	false	"true — следующая после id"
//	@Success		200		{object}	serializer.Response{data=serializer.FeedSerializer}
//	@Failure		400		{object}	serializer.Response
//	@Failure		404		{object}	serializer.Response
//	@Router			/api/import_categories/feed [get]
func (h *Handler) GetFeed(c *gin.Context) {
	ids, err := h.Repository.FeedIDs()
	if err != nil {
		h.failRepo(c, err)
		return
	}
	if len(ids) == 0 {
		h.fail(c, http.StatusNotFound, errors.New("опубликованных услуг нет"))
		return
	}

	pos := 0 // позиция в ленте
	if raw := c.Query("id"); raw != "" {
		id, err := parseID(raw)
		if err != nil {
			h.fail(c, http.StatusBadRequest, err)
			return
		}
		pos = indexOf(ids, id)
		if pos < 0 {
			h.fail(c, http.StatusNotFound, fmt.Errorf("услуги %d нет в ленте", id))
			return
		}
		if c.Query("next") == "true" {
			pos = (pos + 1) % len(ids) // после последней — снова первая
		}
	}

	item, err := h.Repository.PublishedByID(ids[pos])
	if err != nil {
		h.failRepo(c, err)
		return
	}
	isLiked, err := h.Repository.IsLiked(auth.CurrentUserID(c), item.ID)
	if err != nil {
		h.failRepo(c, err)
		return
	}

	next := ids[(pos+1)%len(ids)]
	h.success(c, http.StatusOK, serializer.ToFeed(item, isLiked, pos+1, len(ids), next, h.Repository), "")
}

// GetDraft — GET /api/import_categories/draft
// Черновик текущего пользователя; ид не указывается, черновик один.
//
//	@Summary		Черновик
//	@Description	Черновик текущего пользователя (не больше одного). Нужна авторизация.
//	@Tags			import_categories
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	serializer.Response{data=serializer.DraftSerializer}
//	@Failure		401	{object}	serializer.Response
//	@Failure		404	{object}	serializer.Response
//	@Router			/api/import_categories/draft [get]
func (h *Handler) GetDraft(c *gin.Context) {
	draft, err := h.Repository.DraftByUser(auth.CurrentUserID(c))
	if err != nil {
		h.failRepo(c, err)
		return
	}
	h.success(c, http.StatusOK, serializer.ToDraft(draft, h.Repository), "")
}

// CreateImportCategory — POST /api/import_categories (multipart/form-data)
//
//	title — название (обязательно);
//	image — файл изображения, video — файл короткого видео (файл можно не
//	выбирать: тогда в поле пишется пустая строка, клиент покажет заглушку).
//
// Создаёт черновик текущего пользователя. Файлы уходят в MinIO под именами
// на латинице, в поля image и video таблицы пишутся эти имена.
//
//	@Summary		Добавление услуги (шаг «Далее»)
//	@Description	Создаёт черновик: название, фото и видео (файлами). Создатель — пользователь из токена. Нужна авторизация.
//	@Tags			import_categories
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			title	formData	string	true	"Название"
//	@Param			image	formData	file	false	"Фото: JPG, PNG, WEBP до 10 МБ"
//	@Param			video	formData	file	false	"Видео: MP4, WEBM до 50 МБ"
//	@Success		201		{object}	serializer.Response{data=serializer.DraftSerializer}
//	@Failure		400		{object}	serializer.Response
//	@Failure		401		{object}	serializer.Response
//	@Failure		409		{object}	serializer.Response	"Черновик уже есть"
//	@Router			/api/import_categories [post]
func (h *Handler) CreateImportCategory(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxCreateBody)
	if err := c.Request.ParseMultipartForm(8 << 20); err != nil {
		h.fail(c, http.StatusBadRequest, errors.New("ожидается multipart/form-data: title, image, video (не больше 60 МБ)"))
		return
	}
	for _, field := range systemFields {
		if _, ok := c.Request.MultipartForm.Value[field]; ok {
			h.fail(c, http.StatusBadRequest, fmt.Errorf("поле %s задаётся на сервере, передавать его нельзя", field))
			return
		}
	}
	for _, field := range []string{"image", "video"} {
		if _, ok := c.Request.MultipartForm.Value[field]; ok {
			h.fail(c, http.StatusBadRequest, fmt.Errorf("поле %s передаётся файлом, а не текстом: имя файла генерирует сервер", field))
			return
		}
	}

	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		h.fail(c, http.StatusBadRequest, errors.New("поле title обязательно"))
		return
	}

	// Черновик уже есть — новый не создаём и файлы не загружаем.
	userID := auth.CurrentUserID(c)
	exists, err := h.Repository.HasDraft(userID)
	if err != nil {
		h.failRepo(c, err)
		return
	}
	if exists {
		h.failRepo(c, repository.ErrDraftExists)
		return
	}

	item := ds.ImportCategory{Title: title, CreatorID: userID}
	if item.Image, err = h.upload(c, "image", repository.Image, title); err != nil {
		h.failRepo(c, err)
		return
	}
	if item.Video, err = h.upload(c, "video", repository.Video, title); err != nil {
		h.Repository.RemoveFile(repository.Image, item.Image)
		h.failRepo(c, err)
		return
	}

	if err := h.Repository.CreateDraft(&item); err != nil {
		// запись не создана — убираем уже загруженные файлы
		h.Repository.RemoveFile(repository.Image, item.Image)
		h.Repository.RemoveFile(repository.Video, item.Video)
		h.failRepo(c, err)
		return
	}

	draft, err := h.Repository.DraftByUser(userID)
	if err != nil {
		h.failRepo(c, err)
		return
	}
	h.Logger.Infof("создан черновик №%d «%s»", draft.ID, draft.Title)
	h.success(c, http.StatusCreated, serializer.ToDraft(draft, h.Repository), "черновик создан")
}

// upload загружает файл из поля формы в MinIO и возвращает его имя.
// Файл не выбран — не ошибка: в поле запишется пустая строка.
func (h *Handler) upload(c *gin.Context, field string, kind repository.FileKind, title string) (string, error) {
	header, err := c.FormFile(field)
	if errors.Is(err, http.ErrMissingFile) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("%w: поле %s: %v", repository.ErrBadFile, field, err)
	}
	return h.Repository.UploadFile(kind, title, header)
}

// PublishImportCategory — PUT /api/import_categories/:id/publish
// Тело: {"date_start": "0250-01-01", "date_end": "0050-01-01", "description": "..."}
// Черновик → опубликована. Только свой черновик; вернуть в черновик нельзя.
//
//	@Summary		Публикация услуги
//	@Description	Черновик → опубликована: два поля по теме и описание. Только свой черновик. Нужна авторизация (гостю 401).
//	@Tags			import_categories
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Ид черновика"
//	@Param			body	body		serializer.PublishRequest	true	"Поля публикации"
//	@Success		200		{object}	serializer.Response{data=serializer.ImportCategorySerializer}
//	@Failure		400		{object}	serializer.Response
//	@Failure		401		{object}	serializer.Response
//	@Failure		404		{object}	serializer.Response
//	@Router			/api/import_categories/{id}/publish [put]
func (h *Handler) PublishImportCategory(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		h.fail(c, http.StatusBadRequest, err)
		return
	}

	var req serializer.PublishRequest
	if err := bindJSON(c, &req); err != nil {
		h.fail(c, http.StatusBadRequest, err)
		return
	}
	start, err1 := time.Parse(ds.DateLayout, req.DateStart)
	end, err2 := time.Parse(ds.DateLayout, req.DateEnd)
	if err1 != nil || err2 != nil {
		h.fail(c, http.StatusBadRequest, errors.New("date_start и date_end: даты в формате ГГГГ-ММ-ДД"))
		return
	}
	// Годы до н. э. хранятся положительным числом: начало бытования
	// не может быть позже конца, то есть хранимая дата начала >= даты конца.
	if start.Before(end) {
		h.fail(c, http.StatusBadRequest, errors.New("дата начала должна быть раньше даты конца"))
		return
	}

	userID := auth.CurrentUserID(c)
	description := ds.Optional(strings.TrimSpace(req.Description))
	if err := h.Repository.PublishDraft(id, userID, start, end, description); err != nil {
		h.failRepo(c, err)
		return
	}

	item, err := h.Repository.PublishedByID(id)
	if err != nil {
		h.failRepo(c, err)
		return
	}
	h.Logger.Infof("услуга №%d опубликована", id)
	h.success(c, http.StatusOK, serializer.ToImportCategory(item, userID, h.Repository), "услуга опубликована")
}

// DeleteImportCategory — DELETE /api/import_categories/:id
// Мягкое удаление (status = deleted). Только услуги текущего пользователя.
//
//	@Summary		Удаление услуги
//	@Description	Мягкое удаление (status = deleted). Только свои услуги. Нужна авторизация.
//	@Tags			import_categories
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Ид услуги"
//	@Success		200	{object}	serializer.Response
//	@Failure		400	{object}	serializer.Response
//	@Failure		401	{object}	serializer.Response
//	@Failure		404	{object}	serializer.Response
//	@Router			/api/import_categories/{id} [delete]
func (h *Handler) DeleteImportCategory(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		h.fail(c, http.StatusBadRequest, err)
		return
	}
	if err := h.Repository.DeleteImportCategory(id, auth.CurrentUserID(c)); err != nil {
		h.failRepo(c, err)
		return
	}
	h.Logger.Infof("услуга №%d удалена (soft delete)", id)
	h.success(c, http.StatusOK, gin.H{"id": id, "status": ds.StatusDeleted}, "услуга удалена")
}

// LikeImportCategory — POST /api/import_categories/:id/like
// Тело: {"like": 1} — поставить лайк, {"like": 0} — отменить.
//
//	@Summary		Лайк
//	@Description	like = 1 ставит лайк от текущего пользователя, like = 0 отменяет. Нужна авторизация.
//	@Tags			import_categories
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int						true	"Ид услуги"
//	@Param			body	body		serializer.LikeRequest	true	"0 или 1"
//	@Success		200		{object}	serializer.Response{data=serializer.LikeSerializer}
//	@Failure		400		{object}	serializer.Response
//	@Failure		401		{object}	serializer.Response
//	@Failure		404		{object}	serializer.Response
//	@Failure		409		{object}	serializer.Response
//	@Router			/api/import_categories/{id}/like [post]
func (h *Handler) LikeImportCategory(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		h.fail(c, http.StatusBadRequest, err)
		return
	}
	var req serializer.LikeRequest
	if err := bindJSON(c, &req); err != nil {
		h.fail(c, http.StatusBadRequest, err)
		return
	}

	likes, err := h.Repository.SetLike(auth.CurrentUserID(c), id, *req.Like == 1)
	if err != nil {
		h.failRepo(c, err)
		return
	}
	message := "лайк поставлен"
	if *req.Like == 0 {
		message = "лайк отменён"
	}
	h.success(c, http.StatusOK, serializer.LikeSerializer{ImportCategoryID: id, IsLiked: *req.Like, Likes: likes}, message)
}

func indexOf(ids []uint, id uint) int {
	for i, v := range ids {
		if v == id {
			return i
		}
	}
	return -1
}
