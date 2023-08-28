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
			`SELECT COUNT(*) = 1 AS properUsers
		FROM information_schema.columns
		WHERE table_schema = 'public'
		AND table_name = 'users'
		AND column_name = 'id' AND data_type = 'integer';`,
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
			mock.ExpectQuery(queries[0]).WillReturnRows(sqlmock.NewRows([]string{"properUsers"}).AddRow("true"))
			mock.ExpectQuery(queries[1]).WillReturnRows(sqlmock.NewRows([]string{"properSegments"}).AddRow("true"))
			mock.ExpectQuery(queries[2]).WillReturnRows(sqlmock.NewRows([]string{"properRelations"}).AddRow("true"))

			err := checkDB(db)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
				return
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				fmt.Printf("unmet expectation error: %s", err)
			}
		})

		t.Run("db with wrong 'users' table", func(t *testing.T) {
			mock.ExpectQuery(queries[0]).WillReturnRows(sqlmock.NewRows([]string{"properUsers"}).AddRow("false"))
			mock.ExpectQuery(queries[1]).WillReturnRows(sqlmock.NewRows([]string{"properSegments"}).AddRow("true"))
			mock.ExpectQuery(queries[2]).WillReturnRows(sqlmock.NewRows([]string{"properRelations"}).AddRow("true"))

			expected := "'users' table is not ok: proper 'users' table is { id INTEGER }"

			err := checkDB(db)
			if err == nil || err.Error() != expected {
				t.Fatalf(`got err = "%s", expected "%s"`, err, expected)
				return
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				fmt.Printf("unmet expectation error: %s", err)
			}
		})

		t.Run("db with wrong 'segments' table", func(t *testing.T) {
			mock.ExpectQuery(queries[0]).WillReturnRows(sqlmock.NewRows([]string{"properUsers"}).AddRow("true"))
			mock.ExpectQuery(queries[1]).WillReturnRows(sqlmock.NewRows([]string{"properSegments"}).AddRow("false"))
			mock.ExpectQuery(queries[2]).WillReturnRows(sqlmock.NewRows([]string{"properRelations"}).AddRow("true"))

			expected := "'segments' table is not ok: proper 'segments' table is { id INTEGER; slug TEXT }"

			err := checkDB(db)
			if err == nil || err.Error() != expected {
				t.Fatalf(`got err = "%s", expected "%s"`, err, expected)
				return
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				fmt.Printf("unmet expectation error: %s", err)
			}
		})

		t.Run("db with wrong 'user_segment_relations' table", func(t *testing.T) {
			mock.ExpectQuery(queries[0]).WillReturnRows(sqlmock.NewRows([]string{"properUsers"}).AddRow("true"))
			mock.ExpectQuery(queries[1]).WillReturnRows(sqlmock.NewRows([]string{"properSegments"}).AddRow("true"))
			mock.ExpectQuery(queries[2]).WillReturnRows(sqlmock.NewRows([]string{"properRelations"}).AddRow("false"))

			expected := "'user_segment_relations' table is not ok: proper 'user_segment_relations' table is { user_id INTEGER; segment_id INTEGER }"

			err := checkDB(db)
			if err == nil || err.Error() != expected {
				t.Fatalf(`got err = "%s", expected "%s"`, err, expected)
				return
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				fmt.Printf("unmet expectation error: %s", err)
			}
		})

		t.Run("db with both wrong 'users' and 'segment' table", func(t *testing.T) {
			mock.ExpectQuery(queries[0]).WillReturnRows(sqlmock.NewRows([]string{"properUsers"}).AddRow("false"))
			mock.ExpectQuery(queries[1]).WillReturnRows(sqlmock.NewRows([]string{"properSegments"}).AddRow("false"))
			mock.ExpectQuery(queries[2]).WillReturnRows(sqlmock.NewRows([]string{"properRelations"}).AddRow("true"))

			expected := `'users' table is not ok: proper 'users' table is { id INTEGER }
'segments' table is not ok: proper 'segments' table is { id INTEGER; slug TEXT }`

			err := checkDB(db)
			if err == nil || err.Error() != expected {
				t.Fatalf(`got err = "%s", expected "%s"`, err, expected)
				return
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				fmt.Printf("unmet expectation error: %s", err)
			}
		})
	}
}

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
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if id != testId {
				t.Fatalf("got id = %d, expected %d", id, testId)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				fmt.Printf("unmet expectation error: %s", err)
			}
		})

		t.Run("wrong case", func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectQuery(query).WithArgs(testSlug).WillReturnError(errors.New(testErrText))
			mock.ExpectRollback()

			_, err := addSegmentToDB(db, testSlug)
			expectedErr := "error while adding segment to the database: " + testErrText

			if err == nil || err.Error() != expectedErr {
				t.Fatalf(`got err = "%s", expected err = "%s"`, err, expectedErr)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				fmt.Printf("unmet expectation error: %s", err)
			}
		})

		t.Run("error while starting transaction", func(t *testing.T) {
			mock.ExpectBegin().WillReturnError(errors.New(testErrText))

			_, err := addSegmentToDB(db, testSlug)
			expectedErr := "error while starting transaction: " + testErrText

			if err == nil || err.Error() != expectedErr {
				t.Fatalf(`got err = "%s", expected err = "%s"`, err, expectedErr)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				fmt.Printf("unmet expectation error: %s", err)
			}
		})

		t.Run("error while commiting transaction", func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectQuery(query).WithArgs(testSlug).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testId))
			mock.ExpectCommit().WillReturnError(errors.New(testErrText))

			_, err := addSegmentToDB(db, testSlug)
			expectedErr := "error while committing transaction: " + testErrText

			if err == nil || err.Error() != expectedErr {
				t.Fatalf(`got err = "%s", expected err = "%s"`, err, expectedErr)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				fmt.Printf("unmet expectation error: %s", err)
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
			query       = "DELETE FROM segments WHERE slug = $1;"
		)

		t.Run("normal case", func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectExec(query).WithArgs(testSlug).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()

			err := deleteSegmentFromDB(db, testSlug)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				fmt.Printf("unmet expectation error: %s", err)
			}
		})

		t.Run("wrong case", func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectExec(query).WillReturnError(errors.New(testErrText))
			mock.ExpectRollback()

			err := deleteSegmentFromDB(db, testSlug)
			expectedErr := fmt.Sprintf("error while deleting segment with slug = %s from the database: %s", testSlug, testErrText)

			if err == nil || err.Error() != expectedErr {
				t.Fatalf(`got err = "%s", expected err = "%s"`, err, expectedErr)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				fmt.Printf("unmet expectation error: %s", err)
			}
		})

		t.Run("error while starting transaction", func(t *testing.T) {
			mock.ExpectBegin().WillReturnError(errors.New(testErrText))

			err := deleteSegmentFromDB(db, testSlug)
			expectedErr := "error while starting transaction: " + testErrText

			if err == nil || err.Error() != expectedErr {
				t.Fatalf(`got err = "%s", expected err = "%s"`, err, expectedErr)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				fmt.Printf("unmet expectation error: %s", err)
			}
		})

		t.Run("error while commiting transaction", func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectExec(query).WithArgs(testSlug).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit().WillReturnError(errors.New(testErrText))

			err := deleteSegmentFromDB(db, testSlug)
			expectedErr := "error while committing transaction: " + testErrText

			if err == nil || err.Error() != expectedErr {
				t.Fatalf(`got err = "%s", expected err = "%s"`, err, expectedErr)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				fmt.Printf("unmet expectation error: %s", err)
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
				`INSERT INTO users (id) SELECT $1 WHERE NOT EXISTS (SELECT 1 FROM users WHERE id = $1);`,
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
			mock.ExpectExec(
				queries[0]).WithArgs(testId).WillReturnResult(sqlmock.NewResult(0, rand.Int63n(2)))
			for _, segment := range testAppend {
				mock.ExpectExec(
					queries[1]).WithArgs(testId, segment).WillReturnResult(sqlmock.NewResult(0, rand.Int63n(2)))
			}
			for _, segment := range testRemove {
				mock.ExpectExec(
					queries[2]).WithArgs(testId, segment).WillReturnResult(sqlmock.NewResult(0, rand.Int63n(2)))
			}
			mock.ExpectCommit()

			err := modifyUserInDB(db, testId, testAppend, testRemove)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				fmt.Printf("unmet expectation error: %s", err)
			}
		})

		t.Run("some relations already exist", func(t *testing.T) {
			expectedErrStr := ""

			mock.ExpectBegin()
			mock.ExpectExec(
				queries[0]).
				WithArgs(testId).
				WillReturnResult(sqlmock.NewResult(0, rand.Int63n(2)))
			for _, segment := range testAppend {
				if rand.Intn(2) == 0 {
					mock.ExpectExec(
						queries[1]).WithArgs(testId, segment).WillReturnResult(sqlmock.NewResult(0, rand.Int63n(2)))
				} else {
					err := fmt.Errorf("duplicate key value violates unique constraint %d", rand.Int())
					mock.ExpectExec(
						queries[1]).WithArgs(testId, segment).WillReturnError(err)

					expectedErrStr += fmt.Sprintf(`error while adding user %d to the segment "%s": %s`, testId, segment, err.Error())
					expectedErrStr += fmt.Sprintln()
				}
			}
			for _, segment := range testRemove {
				mock.ExpectExec(queries[2]).WithArgs(testId, segment).WillReturnResult(sqlmock.NewResult(0, rand.Int63n(2)))
			}
			mock.ExpectCommit()

			err := modifyUserInDB(db, testId, testAppend, testRemove)
			if err == nil && expectedErrStr != "" || err != nil && err.Error() != expectedErrStr {
				t.Fatalf("got err = %s, expected err = \n%s", err, expectedErrStr)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				fmt.Printf("unmet expectation error: %s", err)
			}
		})

		t.Run("error while starting transaction", func(t *testing.T) {
			mock.ExpectBegin().WillReturnError(errors.New(testErrText))

			err := modifyUserInDB(db, testId, testAppend, testRemove)
			expectedErr := fmt.Sprintf("error while starting transaction: %s", testErrText)

			if err == nil || err.Error() != expectedErr {
				t.Fatalf(`got err = "%s", expected err = "%s"`, err, expectedErr)
			}
		})

		t.Run("error while commiting transaction", func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectExec(
				queries[0]).WithArgs(testId).WillReturnResult(sqlmock.NewResult(0, rand.Int63n(2)))
			for _, segment := range testAppend {
				mock.ExpectExec(
					queries[1]).WithArgs(testId, segment).WillReturnResult(sqlmock.NewResult(0, rand.Int63n(2)))
			}
			for _, segment := range testRemove {
				mock.ExpectExec(
					queries[2]).WithArgs(testId, segment).WillReturnResult(sqlmock.NewResult(0, rand.Int63n(2)))
			}
			mock.ExpectCommit().WillReturnError(errors.New(testErrText))

			err := modifyUserInDB(db, testId, testAppend, testRemove)
			expectedErr := fmt.Sprintf("error while committing transaction: %s", testErrText)

			if err == nil || err.Error() != expectedErr {
				t.Fatalf(`got err = "%s", expected err = "%s"`, err, expectedErr)
			}
		})
	}
}

func Test_checkupUserInDB(t *testing.T) {
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

			segments, err := checkupUserInDB(db, testId)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if reflect.DeepEqual(segments, testSegments) == false {
				t.Fatalf("got segments = %v, expected %v", segments, testSegments)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				fmt.Printf("unmet expectation error: %s", err)
			}
		})

		t.Run("error while getting user's segments from the database", func(t *testing.T) {
			mock.ExpectQuery(query).WithArgs(testId).WillReturnError(testErr)

			_, err := checkupUserInDB(db, testId)
			expectedErr := fmt.Sprintf("error while getting user %d's segments from the database: %s", testId, testErr)
			if err == nil || err.Error() != expectedErr {
				t.Fatalf(`got err = "%s", expected err = "%s"`, err, expectedErr)
			}
		})
	}
}
