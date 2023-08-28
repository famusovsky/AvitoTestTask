package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/famusovsky/AvitoTestTask/internal/usersegmentation/models"
)

// UserSegmentation - модель базы данных сегментирования пользователей.
type UserSegmentation struct {
	db *sql.DB
}

// GetModel - создание модели базы данных сегментирования пользователей.
//
// Принимает базу данных.
//
// Возвращает модель базы данных сегментирования пользователей и ошибку.
func GetModel(db *sql.DB) (models.UserSegmentationDbProcessor, error) {
	err := checkDB(db)
	if err != nil {
		return nil, err
	}

	return &UserSegmentation{db}, nil
}

// AddSegment - добавление нового сегмента в базу данных.
//
// Принимает: имя сегмента.
//
// Возвращает: id добавленного сегмента и ошибку.
func (model *UserSegmentation) AddSegment(slug string) (int, error) {
	return addSegmentToDB(model.db, slug)
}

// DeleteSegment - удаление сегмента из базы данных.
//
// Принимает: имя сегмента.
//
// Возвращает: ошибку.
func (model *UserSegmentation) DeleteSegment(slug string) error {
	return deleteSegmentFromDB(model.db, slug)
}

// ModifyUser - изменение пользователя по id.
//
// Принимает: id пользователя, имена сегментов, в которые необходимо добавить пользователя, и имена сегментов, из которых необходимо убрать пользователя.
//
// Возвращает: ошибку.
func (model *UserSegmentation) ModifyUser(id int, append []string, remove []string) error {
	return modifyUserInDB(model.db, id, append, remove)
}

// CheckupUser - получение данных о пользователе по id.
//
// Принимает: id пользователя.
//
// Возвращает: список сегментов, в которых состоит пользователь и ошибку.
func (model *UserSegmentation) CheckupUser(id int) ([]string, error) {
	return checkupUserInDB(model.db, id)
}

// addSegmentToDB - добавление нового сегмента в базу данных.
//
// Принимает: указатель на базу данных и имя сегмента.
//
// Возвращает: id добавленного сегмента и ошибку.
func addSegmentToDB(db *sql.DB, slug string) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, errors.New("error while starting transaction: " + err.Error())
	}
	defer tx.Rollback()

	q := `INSERT INTO segments (slug) VALUES ($1) RETURNING id;`
	var id int
	err = tx.QueryRow(q, slug).Scan(&id)
	if err != nil {
		return 0, errors.New("error while adding segment to the database: " + err.Error())
	}
	err = tx.Commit()
	if err != nil {
		return 0, errors.New("error while committing transaction: " + err.Error())
	}

	return id, nil
}

// deleteSegmentFromDB - удаление сегмента из базы данных.
//
// Принимает: указатель на базу данных и имя сегмента.
//
// Возвращает: ошибку.
func deleteSegmentFromDB(db *sql.DB, slug string) error {
	tx, err := db.Begin()
	if err != nil {
		return errors.New("error while starting transaction: " + err.Error())
	}
	defer tx.Rollback()

	q := `DELETE FROM segments WHERE slug = $1;`
	_, err = tx.Exec(q, slug)
	if err != nil {
		return fmt.Errorf("error while deleting segment with slug = %s from the database: %s", slug, err.Error())
	}
	err = tx.Commit()
	if err != nil {
		return errors.New("error while committing transaction: " + err.Error())
	}

	return nil
}

// modifyUserInDB - изменение пользователя в базе данных по id.
//
// Принимает: указатель на базу данных, id пользователя, имена сегментов, в которые необходимо добавить пользователя, и имена сегментов, из которых необходимо убрать пользователя.
//
// Возвращает: ошибку.
func modifyUserInDB(db *sql.DB, id int, append []string, remove []string) error {
	tx, err := db.Begin()
	if err != nil {
		return errors.New("error while starting transaction: " + err.Error())
	}
	defer tx.Rollback()

	errText := ""

	for _, slug := range append {
		_, err = db.Exec(`INSERT INTO user_segment_relations (user_id, segment_id) SELECT $1, id FROM segments WHERE slug = $2;`, id, slug)
		if err != nil {
			errText += fmt.Sprintf(`error while adding user %d to the segment "%s": %s`, id, slug, err.Error())
			errText += fmt.Sprintln()
		}
	}

	for _, slug := range remove {
		_, err = db.Exec(`DELETE FROM user_segment_relations WHERE user_id = $1 AND segment_id = (SELECT id FROM segments WHERE slug = $2);`, id, slug)
		if err != nil {
			errText += fmt.Sprintf(`error while removing user %d from the segment "%s": %s`, id, slug, err.Error())
			errText += fmt.Sprintln()
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.New("error while committing transaction: " + err.Error())
	}

	if errText != "" {
		return errors.New(errText)
	}
	return nil
}

// checkupUserInDB - получение данных о пользователе из базы данных по id.
//
// Принимает: указатель на базу данных и id пользователя.
//
// Возвращает: список сегментов, в которых состоит пользователь и ошибку.
func checkupUserInDB(db *sql.DB, id int) ([]string, error) {
	q := `SELECT slug FROM segments WHERE id IN (SELECT segment_id FROM user_segment_relations WHERE user_id = $1);`
	rows, err := db.Query(q, id)
	if err != nil {
		return []string{}, fmt.Errorf("error while getting user %d's segments from the database: %s", id, err.Error())
	}
	defer rows.Close()

	segments := make([]string, 0)
	for rows.Next() {
		var slug string
		err = rows.Scan(&slug)
		if err != nil {
			return []string{}, fmt.Errorf("error while getting user %d's segments from the database: %s", id, err.Error())
		}
		segments = append(segments, slug)
	}

	return segments, nil
}

// checkDB - проверка базы данных на соответствие требуемой б.д. сегментирования пользователей.
// Возвращает: ошибку.
func checkDB(db *sql.DB) error {
	var (
		qSegments = `SELECT COUNT(*) = 2 AS properSegments
		FROM information_schema.columns
		WHERE table_schema = 'public'
		AND table_name = 'segments'
		AND (
			(column_name = 'id' AND data_type = 'integer')
			OR (column_name = 'slug' AND data_type = 'text')
		);`
		qRelations = `SELECT COUNT(*) = 2 AS properRelations
		FROM information_schema.columns
		WHERE table_schema = 'public'
		AND table_name = 'user_segment_relations'
		AND (
			(column_name = 'user_id' AND data_type = 'integer')
			OR (column_name = 'segment_id' AND data_type = 'integer')
		);`
		properSegments  bool
		properRelations bool
	)

	var err error = nil

	err = db.QueryRow(qSegments).Scan(&properSegments)
	if err != nil {
		return errors.Join(errors.New("error while checking 'segments' table"), err)
	}
	err = db.QueryRow(qRelations).Scan(&properRelations)
	if err != nil {
		return errors.Join(errors.New("error while checking 'user_segment_relations' table"), err)
	}

	if !properSegments {
		err = errors.Join(err, errors.New(
			"'segments' table is not ok: proper 'segments' table is { id INTEGER; slug TEXT }"))
	}
	if !properRelations {
		err = errors.Join(err, errors.New(
			"'user_segment_relations' table is not ok: proper 'user_segment_relations' table is { user_id INTEGER; segment_id INTEGER }"))
	}

	return err
}
