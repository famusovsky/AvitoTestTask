package postgres

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func Test_checkDB(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal("error creating mock database")
	}
	defer db.Close()

	for i := 0; i < 10; i++ {
		queries := []string{
			`SELECT COUNT(*) = 2 AS properSegments
		FROM information_schema.columns
		WHERE table_schema = 'public'
		AND table_name = 'segments'
		AND (
			(column_name = 'id' AND data_type = 'integer')
			OR (column_name = 'slug' AND data_type = 'text')
		);`,
			`SELECT COUNT(*) = 2 AS properRelations
		FROM information_schema.columns
		WHERE table_schema = 'public'
		AND table_name = 'user_segment_relations'
		AND (
			(column_name = 'user_id' AND data_type = 'integer')
			OR (column_name = 'segment_id' AND data_type = 'integer')
		);`,
		}

		t.Run("normal db", func(t *testing.T) {
			mock.ExpectQuery(queries[0]).WillReturnRows(sqlmock.NewRows([]string{"properSegments"}).AddRow("true"))
			mock.ExpectQuery(queries[1]).WillReturnRows(sqlmock.NewRows([]string{"properRelations"}).AddRow("true"))

			err = checkResponce(checkDB(db), nil, mock, t)
			if err != nil {
				t.Error(err)
			}
		})

		segmentErr := errors.New("'segments' table is not ok: proper 'segments' table is { id INTEGER; slug TEXT }")
		relationsErr := errors.New("'user_segment_relations' table is not ok: proper 'user_segment_relations' table is { user_id INTEGER; segment_id INTEGER }")

		t.Run("db with wrong 'segments' table", func(t *testing.T) {
			mock.ExpectQuery(queries[0]).WillReturnRows(sqlmock.NewRows([]string{"properSegments"}).AddRow("false"))
			mock.ExpectQuery(queries[1]).WillReturnRows(sqlmock.NewRows([]string{"properRelations"}).AddRow("true"))

			err = checkResponce(checkDB(db), segmentErr, mock, t)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("db with wrong 'user_segment_relations' table", func(t *testing.T) {
			mock.ExpectQuery(queries[0]).WillReturnRows(sqlmock.NewRows([]string{"properSegments"}).AddRow("true"))
			mock.ExpectQuery(queries[1]).WillReturnRows(sqlmock.NewRows([]string{"properRelations"}).AddRow("false"))

			err = checkResponce(checkDB(db), relationsErr, mock, t)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("db with both wrong 'segments' and 'user_segment_relations' table", func(t *testing.T) {
			mock.ExpectQuery(queries[0]).WillReturnRows(sqlmock.NewRows([]string{"properSegments"}).AddRow("false"))
			mock.ExpectQuery(queries[1]).WillReturnRows(sqlmock.NewRows([]string{"properRelations"}).AddRow("false"))

			err = checkResponce(checkDB(db), errors.Join(segmentErr, relationsErr), mock, t)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func Test_createDB(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal("error creating mock database")
	}
	defer db.Close()

	for i := 0; i < 10; i++ {
		query :=
			`CREATE TABLE IF NOT EXISTS user_segment_relations (
			user_id INTEGER,
			segment_id INTEGER,
			CONSTRAINT unique_user_segment UNIQUE (user_id, segment_id)
		);
		
		CREATE TABLE IF NOT EXISTS segments (
			id SERIAL UNIQUE,
			slug TEXT PRIMARY KEY
		);
			
		CREATE TABLE IF NOT EXISTS logs (
			user_id INTEGER,
			slug TEXT,
			type TEXT,
			created_at TIMESTAMP
		);`

		t.Run("normal case", func(t *testing.T) {
			mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(0, 1))

			err = checkResponce(createDB(db), nil, mock, t)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("error while creating tables", func(t *testing.T) {
			mock.ExpectExec(query).WillReturnError(errors.New("test error"))

			err = checkResponce(createDB(db), errors.New("error while creating tables: test error"), mock, t)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

const (
	startTransactionErrText  = "error while starting transaction: "
	commitTransactionErrText = "error while committing transaction: "
)

func Test_addSegmentToDB(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal("error creating mock database")
	}
	defer db.Close()

	for i := 0; i < 10; i++ {
		var (
			testId      = rand.Int()
			testErrText = "test error " + strconv.Itoa(testId)
			testSlug    = "TEST " + strconv.Itoa(testId)
			query       = "INSERT INTO segments (slug) VALUES ($1) RETURNING id;"
		)

		t.Run("normal case", func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectQuery(query).WithArgs(testSlug).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testId))
			mock.ExpectCommit()

			id, err := addSegmentToDB(db, testSlug)
			err = checkResponce(err, nil, mock, t)
			if err != nil {
				t.Error(err)
			}

			if id != testId {
				t.Fatalf("got id = %d, expected %d", id, testId)
			}
		})

		t.Run("wrong case", func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectQuery(query).WithArgs(testSlug).WillReturnError(errors.New(testErrText))
			mock.ExpectRollback()

			_, err := addSegmentToDB(db, testSlug)
			err = checkResponce(err, fmt.Errorf("error while adding segment to the database: %s", testErrText), mock, t)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("error while starting transaction", func(t *testing.T) {
			mock.ExpectBegin().WillReturnError(errors.New(testErrText))

			_, err := addSegmentToDB(db, testSlug)
			err = checkResponce(err, fmt.Errorf("%s%s", startTransactionErrText, testErrText), mock, t)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("error while commiting transaction", func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectQuery(query).WithArgs(testSlug).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testId))
			mock.ExpectCommit().WillReturnError(errors.New(testErrText))

			_, err := addSegmentToDB(db, testSlug)
			err = checkResponce(err, fmt.Errorf("error while committing transaction: %s", testErrText), mock, t)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func Test_deleteSegmentFromDB(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal("error creating mock database")
	}
	defer db.Close()

	for i := 0; i < 10; i++ {
		var (
			testId      = rand.Int()
			testErrText = "test error " + strconv.Itoa(testId)
			testSlug    = "TEST " + strconv.Itoa(testId)
			queries     = []string{`DELETE FROM user_segment_relations WHERE segment_id = (SELECT id FROM segments WHERE slug = $1);`,
				`DELETE FROM segments WHERE slug = $1;`}
		)

		t.Run("normal case", func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectExec(queries[0]).WithArgs(testSlug).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectExec(queries[1]).WithArgs(testSlug).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()

			err = checkResponce(deleteSegmentFromDB(db, testSlug), nil, mock, t)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("wrong case", func(t *testing.T) {
			expectedErr := fmt.Errorf("error while deleting segment with slug = %s from the database: %s", testSlug, testErrText)

			mock.ExpectBegin()
			mock.ExpectExec(queries[0]).WillReturnError(errors.New(testErrText))
			mock.ExpectRollback()

			err = checkResponce(deleteSegmentFromDB(db, testSlug), expectedErr, mock, t)
			if err != nil {
				t.Error(err)
			}

			mock.ExpectBegin()
			mock.ExpectExec(queries[0]).WillReturnError(errors.New(testErrText))
			mock.ExpectRollback()

			err = checkResponce(deleteSegmentFromDB(db, testSlug), expectedErr, mock, t)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("error while starting transaction", func(t *testing.T) {
			mock.ExpectBegin().WillReturnError(errors.New(testErrText))

			err = checkResponce(deleteSegmentFromDB(db, testSlug),
				fmt.Errorf("%s%s", startTransactionErrText, testErrText), mock, t)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("error while commiting transaction", func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectExec(queries[0]).WithArgs(testSlug).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectExec(queries[1]).WithArgs(testSlug).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit().WillReturnError(errors.New(testErrText))

			err = checkResponce(deleteSegmentFromDB(db, testSlug),
				fmt.Errorf("%s%s", commitTransactionErrText, testErrText), mock, t)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func Test_modifyUserInDB(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal("error creating mock database")
	}
	defer db.Close()

	for i := 0; i < 1; i++ {
		var (
			testId      = rand.Int()
			testErrText = "test error " + strconv.Itoa(testId)
			testAppend  = make([]string, rand.Intn(15))
			testRemove  = make([]string, rand.Intn(15))
			queries     = []string{
				`INSERT INTO user_segment_relations (user_id, segment_id) SELECT $1, id FROM segments WHERE slug = $2;`,
				`DELETE FROM user_segment_relations WHERE user_id = $1 AND segment_id = (SELECT id FROM segments WHERE slug = $2);`,
			}
		)

		for j := 0; j < len(testAppend); j++ {
			testAppend[j] = "TEST " + strconv.Itoa(rand.Int())
		}
		for j := 0; j < len(testRemove); j++ {
			testRemove[j] = "TEST " + strconv.Itoa(rand.Int())
		}

		t.Run("normal case - segment and user could already exist or not", func(t *testing.T) {
			mock.ExpectBegin()
			for _, segment := range testAppend {
				mock.ExpectExec(
					queries[0]).WithArgs(testId, segment).WillReturnResult(sqlmock.NewResult(0, rand.Int63n(2)))
			}
			for _, segment := range testRemove {
				mock.ExpectExec(
					queries[1]).WithArgs(testId, segment).WillReturnResult(sqlmock.NewResult(0, rand.Int63n(2)))
			}
			mock.ExpectCommit()

			err = checkResponce(modifyUserInDB(db, testId, testAppend, testRemove), nil, mock, t)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("some relations already exist", func(t *testing.T) {
			expectedErrStr := ""

			mock.ExpectBegin()
			for _, segment := range testAppend {
				if rand.Intn(2) == 0 {
					mock.ExpectExec(
						queries[0]).WithArgs(testId, segment).WillReturnResult(sqlmock.NewResult(0, rand.Int63n(2)))
				} else {
					err := fmt.Errorf("duplicate key value violates unique constraint %d", rand.Int())
					mock.ExpectExec(
						queries[0]).WithArgs(testId, segment).WillReturnError(err)

					expectedErrStr += fmt.Sprintf(`error while adding user %d to the segment "%s": %s`, testId, segment, err.Error())
					expectedErrStr += fmt.Sprintln()
				}
			}
			for _, segment := range testRemove {
				mock.ExpectExec(queries[1]).WithArgs(testId, segment).WillReturnResult(sqlmock.NewResult(0, rand.Int63n(2)))
			}
			mock.ExpectCommit()

			err = checkResponce(modifyUserInDB(db, testId, testAppend, testRemove), errors.New(expectedErrStr), mock, t)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("error while starting transaction", func(t *testing.T) {
			mock.ExpectBegin().WillReturnError(errors.New(testErrText))

			err = checkResponce(modifyUserInDB(db, testId, testAppend, testRemove), fmt.Errorf("%s%s", startTransactionErrText, testErrText), mock, t)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("error while commiting transaction", func(t *testing.T) {
			mock.ExpectBegin()
			for _, segment := range testAppend {
				mock.ExpectExec(
					queries[0]).WithArgs(testId, segment).WillReturnResult(sqlmock.NewResult(0, rand.Int63n(2)))
			}
			for _, segment := range testRemove {
				mock.ExpectExec(
					queries[1]).WithArgs(testId, segment).WillReturnResult(sqlmock.NewResult(0, rand.Int63n(2)))
			}
			mock.ExpectCommit().WillReturnError(errors.New(testErrText))

			err = checkResponce(modifyUserInDB(db, testId, testAppend, testRemove), fmt.Errorf("%s%s", commitTransactionErrText, testErrText), mock, t)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func Test_GetUserRelationsInDB(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal("error creating mock database")
	}
	defer db.Close()

	for i := 0; i < 1; i++ {
		var (
			testId       = rand.Int()
			testErr      = errors.New("test error " + strconv.Itoa(testId))
			testSegments = make([]string, rand.Intn(15))
			query        = `SELECT slug FROM segments WHERE id IN (SELECT segment_id FROM user_segment_relations WHERE user_id = $1);`
		)

		for j := 0; j < len(testSegments); j++ {
			testSegments[j] = "TEST " + strconv.Itoa(rand.Int())
		}

		t.Run("normal case", func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"slug"})
			for _, segment := range testSegments {
				rows.AddRow(segment)
			}
			mock.ExpectQuery(query).WithArgs(testId).WillReturnRows(rows)

			segments, err := getUserRelationsInDB(db, testId)
			err = checkResponce(err, nil, mock, t)
			if err != nil {
				t.Error(err)
			}

			if reflect.DeepEqual(segments, testSegments) == false {
				t.Fatalf("got segments = %v, expected %v", segments, testSegments)
			}
		})

		t.Run("error while getting user's segments from the database", func(t *testing.T) {
			mock.ExpectQuery(query).WithArgs(testId).WillReturnError(testErr)

			_, err := getUserRelationsInDB(db, testId)
			err = checkResponce(err, fmt.Errorf("error while getting user %d's segments from the database: %s", testId, testErr), mock, t)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func checkResponce(got error, expected error, mock sqlmock.Sqlmock, t *testing.T) error {
	if expected != nil && (got == nil || got.Error() != expected.Error()) {
		return fmt.Errorf("got err = %s\nexpected err = %s\n", got, expected)
	}
	if expected == nil && got != nil {
		return fmt.Errorf("unexpected error: %s\n", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		return fmt.Errorf("unmet expectation error: %s\n", err)
	}
	return nil
}
