package repository

// ListOptions – параметры пагинации/фильтрации, которые переиспользуются во всех репозиториях
type ListOptions struct {
	Search  string // LIKE %search%
	OrderBy string // колонка сортировки, напр. "created_at"
	Desc    bool   // сортировка DESC?
	Limit   int    // размер страницы
	Offset  int    // (page‑1)*size
}
