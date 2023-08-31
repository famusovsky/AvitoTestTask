// models - пакет, содержащий структуры, описывающие сущности, используемые в проекте.
package models

import "time"

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
	// GetLogs - возвращает логи.
	//
	// Принимает: начальное время и конечное время.
	//
	// Возвращает: список логов и ошибку.
	GetLogs(from time.Time, to time.Time) ([]Log, error)
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

// LogTimestamps - структура, описывающая временные рамки для логов.
type LogTimestamps struct {
	From time.Time `json:"from"` // From - начальное время.
	To   time.Time `json:"to"`   // To - конечное время.
}

// Log - структура, описывающая лог.
type Log struct {
	ID                  // ID - id пользователя.
	Segment             // Segment - сегмент.
	Type      string    `json:"type"`       // Type - тип события.
	Timestamp time.Time `json:"created_at"` // Timestamp - время события.
}

// Err - структура, описывающая ошибку.
type Err struct {
	Text string `json:"error"` // Text - текст ошибки.
}
