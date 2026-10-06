package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"rip/lab1/internal/app/ds"
)

// seedCategory — запись для первоначального наполнения таблицы
// import_categories. Поле LikedBy превращается в строки связной таблицы likes.
type seedCategory struct {
	ID               uint
	Slug             string
	Title            string
	Shape            string
	ProductionCenter string
	Region           string
	DateStart        time.Time
	DateEnd          time.Time
	Diagnostics      string
	Description      string
	FindsCount       int
	LikedBy          []int
	Status           ds.Status
}

// seedUsers — пользователи приложения. Пароли до ЛР4 хранятся как есть.
var seedUsers = []ds.User{
	{ID: 1, Login: "umid", Password: "rip2026", FullName: "Умид Солижонов", IsModerator: true},
	{ID: 2, Login: "anna", Password: "rip2026", FullName: "Анна Кузнецова"},
	{ID: 3, Login: "pavel", Password: "rip2026", FullName: "Павел Ермаков"},
	{ID: 4, Login: "irina", Password: "rip2026", FullName: "Ирина Ветрова"},
	{ID: 5, Login: "sergey", Password: "rip2026", FullName: "Сергей Лапин"},
	{ID: 6, Login: "olga", Password: "rip2026", FullName: "Ольга Мещерякова"},
	{ID: 7, Login: "denis", Password: "rip2026", FullName: "Денис Ковалёв"},
	{ID: 8, Login: "marina", Password: "rip2026", FullName: "Марина Стрельцова"},
	{ID: 9, Login: "artem", Password: "rip2026", FullName: "Артём Гордеев"},
	{ID: 10, Login: "elena", Password: "rip2026", FullName: "Елена Багрова"},
	{ID: 11, Login: "nikita", Password: "rip2026", FullName: "Никита Ершов"},
	{ID: 12, Login: "vera", Password: "rip2026", FullName: "Вера Полякова"},
}

// seedCategories — исходная сводка по памятнику.
//
// Содержательная часть (центры производства, датировки, признаки-маркёры)
// опирается на реальную археологическую номенклатуру античного импорта.
// Количества фрагментов — учебная выборка.
//
// Даты бытования задаются годом до нашей эры: ds.BCE(620) — 620 г. до н. э.
var seedCategories = []seedCategory{
	{
		ID: 1, Slug: "attic-black-figure",
		Title:            "Аттическая чернофигурная керамика",
		Shape:            "Килик, амфора, лекиф",
		ProductionCenter: "Афины, квартал Керамик",
		Region:           "Аттика",
		DateStart:        ds.BCE(620),
		DateEnd:          ds.BCE(480),
		Diagnostics:      "Чёрный блестящий лак силуэтом по оранжевой глине; детали процарапаны иглой до глины, добавлены пурпур и белила.",
		Description:      "Массовый показатель прямых связей с Афинами в архаический период. Столовая посуда, попадавшая на периферию вместе с вином и маслом; по формам (килики «типа В», лекифы) датируется с точностью до четверти века, поэтому служит опорной категорией при построении хронологии слоя.",
		FindsCount:       214,
		LikedBy:          []int{1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 91, 102, 117, 128},
		Status:           ds.StatusPublished,
	},
	{
		ID: 2, Slug: "attic-red-figure",
		Title:            "Аттическая краснофигурная керамика",
		Shape:            "Килик, кратер, пелика",
		ProductionCenter: "Афины, квартал Керамик",
		Region:           "Аттика",
		DateStart:        ds.BCE(530),
		DateEnd:          ds.BCE(320),
		Diagnostics:      "Фон залит лаком, фигуры оставлены в цвете глины; детали нанесены накладной рельефной линией.",
		Description:      "Сменяет чернофигурную технику и удерживает рынок до конца IV в. до н. э. Поздние «керченские» пелики и кратеры составляют характерный набор северопричерноморского импорта и маркируют пик афинской торговли.",
		FindsCount:       168,
		LikedBy:          []int{2, 4, 6, 9, 14, 27, 41, 63, 77, 84, 96, 118},
		Status:           ds.StatusPublished,
	},
	{
		ID: 3, Slug: "etruscan-bucchero",
		Title:            "Этрусское буккеро (bucchero nero)",
		Shape:            "Кантар, ойнохоя, киаф",
		ProductionCenter: "Цере (Черветери) и Вульчи, Южная Этрурия",
		Region:           "Этрурия",
		DateStart:        ds.BCE(675),
		DateEnd:          ds.BCE(500),
		Diagnostics:      "Черепок чёрный насквозь (восстановительный обжиг), а не только по поверхности; лощение до металлического блеска, оттиснутые розетки и веерный орнамент.",
		Description:      "Опознаваемая по излому «чёрная» этрусская посуда — та самая «чернощёковая керамика» вариантов задания. Её появление на памятнике фиксирует тирренское направление связей, обычно опосредованное греческими эмпориями, поэтому единичные фрагменты весят в реконструкции больше, чем их доля в массе.",
		FindsCount:       47,
		LikedBy:          []int{3, 7, 11, 19, 23, 38, 52, 61, 74},
		Status:           ds.StatusPublished,
	},
	{
		ID: 4, Slug: "chian-amphora",
		Title:            "Хиосские амфоры",
		Shape:            "Тарная амфора",
		ProductionCenter: "Остров Хиос",
		Region:           "Восточная Эгеида",
		DateStart:        ds.BCE(600),
		DateEnd:          ds.BCE(300),
		Diagnostics:      "Светлая беложгущаяся глина, «пухлое» горло у ранних серий, роспись красной краской по тулову.",
		Description:      "Тара знаменитого хиосского вина — одна из самых ранних массовых импортных категорий в Причерноморье. Эволюция формы горла (пухлое — прямое — колпачковая ножка) даёт независимую шкалу датировок внутри VI–IV вв. до н. э.",
		FindsCount:       96,
		LikedBy:          []int{5, 12, 18, 29, 33, 47, 58, 66, 81, 93},
		Status:           ds.StatusPublished,
	},
	{
		ID: 5, Slug: "thasian-amphora",
		Title:            "Фасосские амфоры",
		Shape:            "Тарная амфора",
		ProductionCenter: "Остров Фасос",
		Region:           "Северная Эгеида",
		DateStart:        ds.BCE(420),
		DateEnd:          ds.BCE(210),
		Diagnostics:      "Плотная слюдистая глина, биконическое тулово; прямоугольные клейма с именем магистрата, эмблемой и этниконом.",
		Description:      "Клеймёная тара делает Фасос эталонной категорией: по именам магистратов сводка находок переводится в узкие хронологические интервалы, а доля фасосского вина показывает вес северноэгейского направления.",
		FindsCount:       132,
		LikedBy:          []int{1, 8, 15, 22, 30, 44, 57, 69, 88},
		Status:           ds.StatusPublished,
	},
	{
		ID: 6, Slug: "heraclean-amphora",
		Title:            "Гераклейские амфоры",
		Shape:            "Тарная амфора",
		ProductionCenter: "Гераклея Понтийская",
		Region:           "Южное Причерноморье",
		DateStart:        ds.BCE(410),
		DateEnd:          ds.BCE(200),
		Diagnostics:      "Коричнево-красная глина без слюды; энглифические (вдавленные) клейма с именем фабриканта и магистрата.",
		Description:      "Ближайший к Понту крупный поставщик вина. Обычно даёт наибольшую долю в массе амфорного материала IV в. до н. э., поэтому именно от неё отсчитывается «фон», на котором видны дальние направления импорта.",
		FindsCount:       187,
		LikedBy:          []int{4, 9, 16, 24, 35, 49, 62, 71, 83, 95, 108},
		Status:           ds.StatusPublished,
	},
	{
		ID: 7, Slug: "sinopean-amphora",
		Title:            "Синопские амфоры",
		Shape:            "Тарная амфора",
		ProductionCenter: "Синопа",
		Region:           "Южное Причерноморье",
		DateStart:        ds.BCE(400),
		DateEnd:          ds.BCE(100),
		Diagnostics:      "Красноглиняное тесто с чёрными включениями пироксена; клейма астиномов с эмблемой и именем фабриканта.",
		Description:      "Вино и оливковое масло Синопы вытесняют гераклейские поставки в III в. до н. э. Пироксеновые включения различимы даже в мелком фрагменте, что делает категорию удобной для количественного учёта.",
		FindsCount:       145,
		LikedBy:          []int{6, 13, 20, 28, 40, 51, 64, 79, 92},
		Status:           ds.StatusPublished,
	},
	{
		ID: 8, Slug: "rhodian-amphora",
		Title:            "Родосские амфоры",
		Shape:            "Тарная амфора",
		ProductionCenter: "Остров Родос",
		Region:           "Восточная Эгеида",
		DateStart:        ds.BCE(300),
		DateEnd:          ds.BCE(40),
		Diagnostics:      "Высоко поднятые «рогатые» ручки; круглые клейма с розой или головой Гелиоса и именем эпонима.",
		Description:      "Эллинистический маркёр дальней средиземноморской торговли. Родосские клейма датируются по периодам эпонимов, что позволяет привязывать верхние слои памятника к абсолютной хронологии.",
		FindsCount:       78,
		LikedBy:          []int{7, 17, 26, 37, 53, 68, 85},
		Status:           ds.StatusPublished,
	},
	{
		ID: 9, Slug: "koan-amphora",
		Title:            "Косские амфоры",
		Shape:            "Тарная амфора",
		ProductionCenter: "Остров Кос",
		Region:           "Восточная Эгеида",
		DateStart:        ds.BCE(300),
		DateEnd:          ds.BCE(30),
		Diagnostics:      "Двуствольные (сдвоенные) ручки — признак, опознаваемый по одному фрагменту; светлая глина.",
		Description:      "Тара косского вина с морской водой. Форма ручки настолько специфична, что категория надёжно выделяется даже в сильно фрагментированном материале, и потому включена в сводку отдельной позицией.",
		FindsCount:       41,
		LikedBy:          []int{10, 25, 39, 54, 72, 87},
		Status:           ds.StatusPublished,
	},
	{
		ID: 10, Slug: "megarian-bowl",
		Title:            "Мегарские рельефные чаши",
		Shape:            "Полусферическая чаша",
		ProductionCenter: "Пергам, ионийские мастерские, Делос",
		Region:           "Малая Азия",
		DateStart:        ds.BCE(225),
		DateEnd:          ds.BCE(75),
		Diagnostics:      "Формованная в матрице полусферическая чаша с рельефными «дольками» и растительным фризом под чёрным или бурым лаком.",
		Description:      "Эллинистическая имитация металлической посуды. Оттиски одной матрицы связывают находки с конкретной мастерской, поэтому категория используется для различения пергамского и ионийского направлений внутри одного региона.",
		FindsCount:       33,
		LikedBy:          []int{11, 31, 45, 59, 76},
		Status:           ds.StatusPublished,
	},
	{
		ID: 11, Slug: "arretine-sigillata",
		Title:            "Арретинская терра сигиллата",
		Shape:            "Тарелка, чаша",
		ProductionCenter: "Арреций (Ареццо), Италия",
		Region:           "Италия",
		DateStart:        ds.BCE(40),
		DateEnd:          ds.BCE(10),
		Diagnostics:      "Ярко-красный блестящий ангоб, тонкий черепок; клейма мастеров in planta pedis (в форме ступни).",
		Description:      "Италийская столовая посуда раннеримского времени. Даже единичные фрагменты фиксируют включение памятника в средиземноморскую систему обмена эпохи Августа и датируются по форме клейма с точностью до десятилетий.",
		FindsCount:       26,
		LikedBy:          []int{14, 32, 48, 67},
		Status:           ds.StatusPublished,
	},
	{
		ID: 12, Slug: "eastern-sigillata-a",
		Title:            "Восточная сигиллата А (ESA)",
		Shape:            "Чаша с загнутым краем",
		ProductionCenter: "Северная Сирия, равнина Амук",
		Region:           "Левант",
		DateStart:        ds.BCE(150),
		DateEnd:          ds.BCE(10),
		Diagnostics:      "Светлая известковая глина с красным ангобом неполного покрытия; характерный загнутый внутрь край.",
		Description:      "Основная краснолаковая столовая серия Восточного Средиземноморья. Отделяется от италийской сигиллаты по цвету теста, что позволяет разделить в сводке западное и восточное направления импорта.",
		FindsCount:       58,
		LikedBy:          []int{18, 36, 50, 73, 90},
		Status:           ds.StatusPublished,
	},
	{
		ID: 13, Slug: "naukratis-faience",
		Title:            "Египетский фаянс: скарабеи и бусы",
		Shape:            "Амулет, бусина, арибалл",
		ProductionCenter: "Навкратис, дельта Нила",
		Region:           "Египет",
		DateStart:        ds.BCE(650),
		DateEnd:          ds.BCE(525),
		Diagnostics:      "Кварцевая масса с бирюзовой щелочной глазурью; оттиснутые иероглифические картуши на основании скарабеев.",
		Description:      "Мелкий престижный импорт, расходившийся по Средиземноморью через греческий эмпорий в Навкратисе. В массе находок доля ничтожна, но категория важна как индикатор дальних, «цепочечных» связей.",
		FindsCount:       62,
		LikedBy:          []int{21, 42, 56, 80, 99},
		Status:           ds.StatusPublished,
	},
	{
		ID: 14, Slug: "phoenician-glass",
		Title:            "Финикийское сердечниковое стекло",
		Shape:            "Алабастр, амфориск, ойнохойя",
		ProductionCenter: "Сидон и Тир, финикийское побережье",
		Region:           "Левант",
		DateStart:        ds.BCE(550),
		DateEnd:          ds.BCE(250),
		Diagnostics:      "Синий или тёмно-фиолетовый корпус, навитые жёлтые и белые нити, растянутые гребнем в зигзаг или «перья».",
		Description:      "Сосудики для благовоний, изготовленные до изобретения стеклодувной трубки на песчано-глиняном сердечнике. Дорогой штучный товар: единичные фрагменты указывают на присутствие торговцев, а не на регулярный поток.",
		FindsCount:       19,
		LikedBy:          []int{29, 60, 82},
		Status:           ds.StatusPublished,
	},
	{
		ID: 15, Slug: "knidian-amphora",
		Title:            "Книдские амфоры",
		Shape:            "Тарная амфора",
		ProductionCenter: "Книд, Кария",
		Region:           "Восточная Эгеида",
		DateStart:        ds.BCE(250),
		DateEnd:          ds.BCE(50),
		Diagnostics:      "Ручки с «книдским» коленчатым изломом; клейма с именем фрурарха и эмблемой — бычьей головой.",
		Description:      "Поздняя эллинистическая тара, замыкающая список эгейских поставок. Книдские клейма фиксируют, что связи с Восточной Эгеидой сохраняются и после спада афинского импорта, поэтому категория важна для верхней границы сводки.",
		FindsCount:       37,
		LikedBy:          []int{5, 19, 44, 71, 96},
		Status:           ds.StatusDraft,
	},
	{
		ID: 16, Slug: "mendean-amphora",
		Title:            "Мендейские амфоры",
		Shape:            "Тарная амфора",
		ProductionCenter: "Менда, полуостров Паллена",
		Region:           "Северная Эгеида",
		DateStart:        ds.BCE(480),
		DateEnd:          ds.BCE(350),
		Diagnostics:      "Светлоглиняное тулово, ручки с прогибом; клейма редки.",
		Description:      "Категория выведена из сводки: определения переатрибутированы к фасосскому кругу мастерских.",
		FindsCount:       54,
		LikedBy:          []int{4, 17},
		Status:           ds.StatusDeleted,
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

			item := ds.ImportCategory{
				ID:               seed.ID,
				Slug:             seed.Slug,
				Title:            seed.Title,
				Shape:            seed.Shape,
				ProductionCenter: seed.ProductionCenter,
				Region:           seed.Region,
				DateStart:        seed.DateStart,
				DateEnd:          seed.DateEnd,
				Diagnostics:      seed.Diagnostics,
				Description:      seed.Description,
				FindsCount:       seed.FindsCount,
				Status:           seed.Status,
				CreatorID:        &creator,
				ImageURL:         r.mediaURL("img", seed.Slug, "jpg"),
				VideoURL:         r.mediaURL("video", seed.Slug, "mp4"),
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
