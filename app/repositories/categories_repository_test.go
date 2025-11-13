package repositories

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newGormWithMock creates a gorm DB backed by sqlmock.
func newGormWithMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	t.Helper()
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	dial := postgres.New(postgres.Config{
		Conn:       conn,
		DriverName: "postgres",
	})
	db, err := gorm.Open(dial, &gorm.Config{})
	if err != nil {
		conn.Close()
		t.Fatalf("failed to open gorm with sqlmock: %v", err)
	}
	cleanup := func() { _ = conn.Close() }
	return db, mock, cleanup
}

func TestCategoriesRepository_ListCategories_Success(t *testing.T) {
	db, mock, cleanup := newGormWithMock(t)
	defer cleanup()

	r := NewCategoriesRepository(db)

	rows := sqlmock.NewRows([]string{"id", "code", "name"}).
		AddRow(1, "clothing", "Clothing").
		AddRow(2, "shoes", "Shoes")

	mock.ExpectQuery(`SELECT\s+.*\s+FROM\s+"categories"\s+ORDER BY id ASC`).
		WillReturnRows(rows)

	ctx := context.Background()
	items, err := r.ListCategories(ctx)
	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, uint(1), items[0].ID)
	assert.Equal(t, "clothing", items[0].Code)
	assert.Equal(t, "Clothing", items[0].Name)
	assert.Equal(t, uint(2), items[1].ID)
	assert.Equal(t, "shoes", items[1].Code)
	assert.Equal(t, "Shoes", items[1].Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoriesRepository_ListCategories_DBError(t *testing.T) {
	db, mock, cleanup := newGormWithMock(t)
	defer cleanup()

	r := NewCategoriesRepository(db)

	mock.ExpectQuery(`SELECT\s+.*\s+FROM\s+"categories"\s+ORDER BY id ASC`).
		WillReturnError(assert.AnError)

	ctx := context.Background()
	items, err := r.ListCategories(ctx)
	assert.Error(t, err)
	assert.Nil(t, items)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoriesRepository_CreateCategory_Success(t *testing.T) {
	db, mock, cleanup := newGormWithMock(t)
	defer cleanup()

	r := NewCategoriesRepository(db)

	// Begin transaction
	mock.ExpectBegin()
	// MAX(id) query
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(MAX(id), 0) FROM "categories"`)).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(2))
	// INSERT returning id
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "categories" ("code","name","id") VALUES ($1,$2,$3) RETURNING "id"`)).
		WithArgs("new-code", "New Name", 3).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
	// Commit
	mock.ExpectCommit()

	ctx := context.Background()
	err := r.CreateCategory(ctx, models.Category{Code: "new-code", Name: "New Name"})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoriesRepository_CreateCategory_MaxQueryError(t *testing.T) {
	db, mock, cleanup := newGormWithMock(t)
	defer cleanup()

	r := NewCategoriesRepository(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(MAX(id), 0)`)).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	ctx := context.Background()
	err := r.CreateCategory(ctx, models.Category{Code: "c", Name: "n"})
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCategoriesRepository_CreateCategory_InsertError(t *testing.T) {
	db, mock, cleanup := newGormWithMock(t)
	defer cleanup()

	r := NewCategoriesRepository(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(MAX(id), 0)`)).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(10))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "categories" ("code","name","id") VALUES ($1,$2,$3) RETURNING "id"`)).
		WithArgs("c", "n", 11).
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	ctx := context.Background()
	err := r.CreateCategory(ctx, models.Category{Code: "c", Name: "n"})
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
