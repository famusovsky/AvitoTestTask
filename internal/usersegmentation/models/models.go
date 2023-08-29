// models - пакет, содержащий структуры, описывающие сущности, используемые в проекте.
package models

// UserSegmentationDbProcessor - интерфейс, предоставляющий методы для работы с БД, хранящей данные о сегментации пользователей.
type UserSegmentationDbProcessor interface {
	// AddSegment - добавляет сегмент в БД.
	//
	// Принимает: название сегмента.
	//
	// Возвращает: id добавленного сегмента и ошибку.
	AddSegment(slug string) (int, error)
	// DeleteSegment - удаляет сегмент из БД.
	//
	// Принимает: название сегмента.
	//
	// Возвращает: ошибку.
	DeleteSegment(slug string) error
	// ModifyUser - изменяет сегменты пользователя.
	//
	// Принимает: id пользователя, имена сегментов, в которые необходимо добавить пользователя, и имена сегментов, из которых необходимо убрать пользователя.
	//
	// Возвращает: ошибку.
	ModifyUser(id int, append []string, remove []string) error
	// GetUserRelations - возвращает сегменты, в которых состоит пользователь.
	//
	// Принимает: id пользователя.
	//
	// Возвращает: список сегментов, в которых состоит пользователь, и ошибку.
	GetUserRelations(id int) ([]string, error)
}

// Segment - структура, описывающая сегмент.
type Segment struct {
	Slug string `json:"slug"` // Slug - название сегмента.
}

// ID - структура, описывающая id.
type ID struct {
	Value int `json:"id"` // Value - id.
}

// UserModification - структура, описывающая изменение сегментов пользователя.
type UserModification struct {
	ID              // ID - id пользователя.
	Append []string `json:"append"` // Append - список сегментов, в которые необходимо добавить пользователя.
	Remove []string `json:"remove"` // Remove - список сегментов, из которых необходимо убрать пользователя.
}

// Err - структура, описывающая ошибку.
type Err struct {
	Text string `json:"error"` // Text - текст ошибки.
}
