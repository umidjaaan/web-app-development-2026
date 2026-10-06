package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"rip/lab1/internal/app/ds"
)

// seedCategory — запись для первоначального наполнения таблицы
// import_categories. Поле LikedBy превращается в строки связной таблицы likes.
// Slug — имя файлов фото и видео в бакете MinIO (в таблицу не пишется).
type seedCategory struct {
	ID          uint
	Slug        string
	Title       string
	DateStart   time.Time
	DateEnd     time.Time
	Description string
	LikedBy     []int
	Status      ds.Status
}

// seedUsers — пользователи приложения. Пароли до ЛР4 хранятся как есть.
var seedUsers = []ds.User{
	{ID: 1, Login: "umid", Password: "rip2026", IsModerator: true},
	{ID: 2, Login: "anna", Password: "rip2026"},
	{ID: 3, Login: "pavel", Password: "rip2026"},
	{ID: 4, Login: "irina", Password: "rip2026"},
	{ID: 5, Login: "sergey", Password: "rip2026"},
	{ID: 6, Login: "olga", Password: "rip2026"},
	{ID: 7, Login: "denis", Password: "rip2026"},
	{ID: 8, Login: "marina", Password: "rip2026"},
	{ID: 9, Login: "artem", Password: "rip2026"},
	{ID: 10, Login: "elena", Password: "rip2026"},
	{ID: 11, Login: "nikita", Password: "rip2026"},
	{ID: 12, Login: "vera", Password: "rip2026"},
}

// seedCategories — исходная сводка по памятнику.
//
// Названия и датировки опираются на реальную археологическую
// номенклатуру античного импорта.
//
// Даты бытования задаются годом до нашей эры: ds.BCE(620) — 620 г. до н. э.
var seedCategories = []seedCategory{
	{
		ID: 1, Slug: "attic-black-figure",
		Title:       "Аттическая чернофигурная керамика",
		DateStart:   ds.BCE(620),
		DateEnd:     ds.BCE(480),
		Description: "Массовый показатель прямых связей с Афинами в архаический период. Столовая посуда, попадавшая на периферию вместе с вином и маслом; по формам (килики «типа В», лекифы) датируется с точностью до четверти века, поэтому служит опорной категорией при построении хронологии слоя.",
		LikedBy:     []int{1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 91, 102, 117, 128},
		Status:      ds.StatusPublished,
	},
	{
		ID: 2, Slug: "attic-red-figure",
		Title:       "Аттическая краснофигурная керамика",
		DateStart:   ds.BCE(530),
		DateEnd:     ds.BCE(320),
		Description: "Сменяет чернофигурную технику и удерживает рынок до конца IV в. до н. э. Поздние «керченские» пелики и кратеры составляют характерный набор северопричерноморского импорта и маркируют пик афинской торговли.",
		LikedBy:     []int{2, 4, 6, 9, 14, 27, 41, 63, 77, 84, 96, 118},
		Status:      ds.StatusPublished,
	},
	{
		ID: 3, Slug: "etruscan-bucchero",
		Title:       "Этрусское буккеро (bucchero nero)",
		DateStart:   ds.BCE(675),
		DateEnd:     ds.BCE(500),
		Description: "Опознаваемая по излому «чёрная» этрусская посуда — та самая «чернощёковая керамика» вариантов задания. Её появление на памятнике фиксирует тирренское направление связей, обычно опосредованное греческими эмпориями, поэтому единичные фрагменты весят в реконструкции больше, чем их доля в массе.",
		LikedBy:     []int{3, 7, 11, 19, 23, 38, 52, 61, 74},
		Status:      ds.StatusPublished,
	},
	{
		ID: 4, Slug: "chian-amphora",
		Title:       "Хиосские амфоры",
		DateStart:   ds.BCE(600),
		DateEnd:     ds.BCE(300),
		Description: "Тара знаменитого хиосского вина — одна из самых ранних массовых импортных категорий в Причерноморье. Эволюция формы горла (пухлое — прямое — колпачковая ножка) даёт независимую шкалу датировок внутри VI–IV вв. до н. э.",
		LikedBy:     []int{5, 12, 18, 29, 33, 47, 58, 66, 81, 93},
		Status:      ds.StatusPublished,
	},
	{
		ID: 5, Slug: "thasian-amphora",
		Title:       "Фасосские амфоры",
		DateStart:   ds.BCE(420),
		DateEnd:     ds.BCE(210),
		Description: "Клеймёная тара делает Фасос эталонной категорией: по именам магистратов сводка находок переводится в узкие хронологические интервалы, а доля фасосского вина показывает вес северноэгейского направления.",
		LikedBy:     []int{1, 8, 15, 22, 30, 44, 57, 69, 88},
		Status:      ds.StatusPublished,
	},
	{
		ID: 6, Slug: "heraclean-amphora",
		Title:       "Гераклейские амфоры",
		DateStart:   ds.BCE(410),
		DateEnd:     ds.BCE(200),
		Description: "Ближайший к Понту крупный поставщик вина. Обычно даёт наибольшую долю в массе амфорного материала IV в. до н. э., поэтому именно от неё отсчитывается «фон», на котором видны дальние направления импорта.",
		LikedBy:     []int{4, 9, 16, 24, 35, 49, 62, 71, 83, 95, 108},
		Status:      ds.StatusPublished,
	},
	{
		ID: 7, Slug: "sinopean-amphora",
		Title:       "Синопские амфоры",
		DateStart:   ds.BCE(400),
		DateEnd:     ds.BCE(100),
		Description: "Вино и оливковое масло Синопы вытесняют гераклейские поставки в III в. до н. э. Пироксеновые включения различимы даже в мелком фрагменте, что делает категорию удобной для количественного учёта.",
		LikedBy:     []int{6, 13, 20, 28, 40, 51, 64, 79, 92},
		Status:      ds.StatusPublished,
	},
	{
		ID: 8, Slug: "rhodian-amphora",
		Title:       "Родосские амфоры",
		DateStart:   ds.BCE(300),
		DateEnd:     ds.BCE(40),
		Description: "Эллинистический маркёр дальней средиземноморской торговли. Родосские клейма датируются по периодам эпонимов, что позволяет привязывать верхние слои памятника к абсолютной хронологии.",
		LikedBy:     []int{7, 17, 26, 37, 53, 68, 85},
		Status:      ds.StatusPublished,
	},
	{
		ID: 9, Slug: "koan-amphora",
		Title:       "Косские амфоры",
		DateStart:   ds.BCE(300),
		DateEnd:     ds.BCE(30),
		Description: "Тара косского вина с морской водой. Форма ручки настолько специфична, что категория надёжно выделяется даже в сильно фрагментированном материале, и потому включена в сводку отдельной позицией.",
		LikedBy:     []int{10, 25, 39, 54, 72, 87},
		Status:      ds.StatusPublished,
	},
	{
		ID: 10, Slug: "megarian-bowl",
		Title:       "Мегарские рельефные чаши",
		DateStart:   ds.BCE(225),
		DateEnd:     ds.BCE(75),
		Description: "Эллинистическая имитация металлической посуды. Оттиски одной матрицы связывают находки с конкретной мастерской, поэтому категория используется для различения пергамского и ионийского направлений внутри одного региона.",
		LikedBy:     []int{11, 31, 45, 59, 76},
		Status:      ds.StatusPublished,
	},
	{
		ID: 11, Slug: "arretine-sigillata",
		Title:       "Арретинская терра сигиллата",
		DateStart:   ds.BCE(40),
		DateEnd:     ds.BCE(10),
		Description: "Италийская столовая посуда раннеримского времени. Даже единичные фрагменты фиксируют включение памятника в средиземноморскую систему обмена эпохи Августа и датируются по форме клейма с точностью до десятилетий.",
		LikedBy:     []int{14, 32, 48, 67},
		Status:      ds.StatusPublished,
	},
	{
		ID: 12, Slug: "eastern-sigillata-a",
		Title:       "Восточная сигиллата А (ESA)",
		DateStart:   ds.BCE(150),
		DateEnd:     ds.BCE(10),
		Description: "Основная краснолаковая столовая серия Восточного Средиземноморья. Отделяется от италийской сигиллаты по цвету теста, что позволяет разделить в сводке западное и восточное направления импорта.",
		LikedBy:     []int{18, 36, 50, 73, 90},
		Status:      ds.StatusPublished,
	},
	{
		ID: 13, Slug: "naukratis-faience",
		Title:       "Египетский фаянс: скарабеи и бусы",
		DateStart:   ds.BCE(650),
		DateEnd:     ds.BCE(525),
		Description: "Мелкий престижный импорт, расходившийся по Средиземноморью через греческий эмпорий в Навкратисе. В массе находок доля ничтожна, но категория важна как индикатор дальних, «цепочечных» связей.",
		LikedBy:     []int{21, 42, 56, 80, 99},
		Status:      ds.StatusPublished,
	},
	{
		ID: 14, Slug: "phoenician-glass",
		Title:       "Финикийское сердечниковое стекло",
		DateStart:   ds.BCE(550),
		DateEnd:     ds.BCE(250),
		Description: "Сосудики для благовоний, изготовленные до изобретения стеклодувной трубки на песчано-глиняном сердечнике. Дорогой штучный товар: единичные фрагменты указывают на присутствие торговцев, а не на регулярный поток.",
		LikedBy:     []int{29, 60, 82},
		Status:      ds.StatusPublished,
	},
	{
		ID: 15, Slug: "knidian-amphora",
		Title:       "Книдские амфоры",
		DateStart:   ds.BCE(250),
		DateEnd:     ds.BCE(50),
		Description: "Поздняя эллинистическая тара, замыкающая список эгейских поставок. Книдские клейма фиксируют, что связи с Восточной Эгеидой сохраняются и после спада афинского импорта, поэтому категория важна для верхней границы сводки.",
		LikedBy:     []int{5, 19, 44, 71, 96},
		Status:      ds.StatusDraft,
	},
	{
		ID: 16, Slug: "mendean-amphora",
		Title:       "Мендейские амфоры",
		DateStart:   ds.BCE(480),
		DateEnd:     ds.BCE(350),
		Description: "Категория выведена из сводки: определения переатрибутированы к фасосскому кругу мастерских.",
		LikedBy:     []int{4, 17},
		Status:      ds.StatusDeleted,
	},
}

// Seed наполняет пустые таблицы исходными данными. Если категории уже есть,
// ничего не делает — повторный запуск миграции безопасен.
func (r *Repository) Seed() error {
	var count int64
	if err := r.db.Model(&ds.ImportCategory{}).Count(&count).Error; err != nil {
		return fmt.Errorf("подсчёт категорий: %w", err)
	}
	if count > 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		users := make([]ds.User, len(seedUsers))
		copy(users, seedUsers)
		if err := tx.Create(&users).Error; err != nil {
			return fmt.Errorf("пользователи: %w", err)
		}

		for i, seed := range seedCategories {
			// Автор записи. Черновик отдаём второму пользователю, чтобы
			// у текущего (первого) страница добавления начиналась с шага 1.
			creator := uint(i%len(seedUsers)) + 1
			if seed.Status == ds.StatusDraft {
				creator = 2
			}

			start, end := seed.DateStart, seed.DateEnd
			item := ds.ImportCategory{
				ID:          seed.ID,
				Title:       seed.Title,
				ImageURL:    r.mediaURL("img", seed.Slug, "jpg"),
				VideoURL:    r.mediaURL("video", seed.Slug, "mp4"),
				DateStart:   &start,
				DateEnd:     &end,
				Description: seed.Description,
				Status:      seed.Status,
				CreatorID:   creator,
			}
			if err := tx.Create(&item).Error; err != nil {
				return fmt.Errorf("категория %s: %w", seed.Slug, err)
			}

			// Лайки: идентификаторы из учебной выборки приводим к числу
			// реальных пользователей и убираем повторы.
			seen := make(map[uint]bool)
			for _, raw := range seed.LikedBy {
				userID := uint((raw-1)%len(seedUsers)) + 1
				if seen[userID] {
					continue
				}
				seen[userID] = true

				like := ds.Like{UserID: userID, ImportCategoryID: item.ID}
				if err := tx.Create(&like).Error; err != nil {
					return fmt.Errorf("лайк %s: %w", seed.Slug, err)
				}
			}
		}

		// Записи вставлены с явными идентификаторами, поэтому счётчики
		// последовательностей нужно подтянуть вручную — иначе следующая
		// вставка через ORM упадёт на конфликте первичного ключа.
		for _, table := range []string{"users", "import_categories", "likes"} {
			query := fmt.Sprintf(
				"SELECT setval(pg_get_serial_sequence('%s', 'id'), COALESCE(MAX(id), 1)) FROM %s",
				table, table)
			if err := tx.Exec(query).Error; err != nil {
				return fmt.Errorf("последовательность %s: %w", table, err)
			}
		}
		return nil
	})
}
