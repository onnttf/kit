package dal

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type testUser struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"size:255"`
	Age  int    `gorm:"size:255"`
}

func TestNewRepo(t *testing.T) {
	repo := NewRepo[testUser]()
	assert.NotNil(t, repo)
}

func TestRepo_Insert(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo[testUser]()

	user := &testUser{Name: "Alice", Age: 25}
	err := repo.Insert(context.Background(), db, user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
}

func TestRepo_Insert_NilDB(t *testing.T) {
	repo := NewRepo[testUser]()
	err := repo.Insert(context.Background(), nil, &testUser{})
	assert.Error(t, err)
}

func TestRepo_Insert_NilValue(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo[testUser]()
	err := repo.Insert(context.Background(), db, nil)
	assert.Error(t, err)
}

func TestRepo_BatchInsert(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo[testUser]()

	users := []*testUser{
		{Name: "User1", Age: 20},
		{Name: "User2", Age: 21},
	}
	err := repo.BatchInsert(context.Background(), db, users, 2)
	assert.NoError(t, err)
}

func TestRepo_UpdateFields(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo[testUser]()

	user := &testUser{Name: "Test", Age: 25}
	require.NoError(t, db.Create(user).Error)

	updates := map[string]any{"name": "NewName"}
	err := repo.UpdateFields(context.Background(), db, updates, Equal("id", user.ID))
	assert.NoError(t, err)
}

func TestRepo_Update_NoScope(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo[testUser]()

	err := repo.Update(context.Background(), db, &testUser{Name: "Unsafe"})
	assert.Error(t, err)
}

func TestRepo_UpdateFields_NoScope(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo[testUser]()

	err := repo.UpdateFields(context.Background(), db, map[string]any{"name": "Unsafe"})
	assert.Error(t, err)
}

func TestRepo_QueryOne(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo[testUser]()

	require.NoError(t, db.Create(&testUser{Name: "FindMe", Age: 20}).Error)

	result, err := repo.QueryOne(context.Background(), db, Equal("name", "FindMe"))
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "FindMe", result.Name)
}

func TestRepo_QueryOne_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo[testUser]()

	result, err := repo.QueryOne(context.Background(), db, Equal("name", "NonExistent"))
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrDatabase))
	assert.Nil(t, result)
}

func TestRepo_Query(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo[testUser]()

	require.NoError(t, db.Create(&testUser{Name: "User1", Age: 20}).Error)
	require.NoError(t, db.Create(&testUser{Name: "User2", Age: 25}).Error)

	results, err := repo.Query(context.Background(), db)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestRepo_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo[testUser]()

	require.NoError(t, db.Create(&testUser{Name: "User1", Age: 20}).Error)

	count, err := repo.Count(context.Background(), db)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo[testUser]()

	require.NoError(t, db.Create(&testUser{Name: "DeleteMe", Age: 20}).Error)

	err := repo.Delete(context.Background(), db, Equal("name", "DeleteMe"))
	assert.NoError(t, err)
}

func TestRepo_Delete_NoScope(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo[testUser]()

	err := repo.Delete(context.Background(), db)
	assert.Error(t, err)
}

func TestRepo_Raw(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo[testUser]()

	require.NoError(t, db.Create(&testUser{Name: "Raw1", Age: 20}).Error)

	results, err := repo.Raw(context.Background(), db, "SELECT * FROM test_users WHERE name = ?", "Raw1")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestExec(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.Create(&testUser{Name: "ToDelete", Age: 20}).Error)
	err := Exec(context.Background(), db, "DELETE FROM test_users WHERE name = ?", "ToDelete")
	assert.NoError(t, err)
}

func TestExec_AllowsZeroRowsAffected(t *testing.T) {
	db := setupTestDB(t)
	err := Exec(context.Background(), db, "DELETE FROM test_users WHERE name = ?", "missing")
	require.NoError(t, err)
}

func TestExec_AllowsDDL(t *testing.T) {
	db := setupTestDB(t)
	err := Exec(context.Background(), db, "CREATE TABLE audit_logs (id integer primary key, message text)")
	require.NoError(t, err)
}

func TestExec_NilDB(t *testing.T) {
	err := Exec(context.Background(), nil, "DELETE FROM test_users")
	assert.Error(t, err)
}

func TestRepositoryInterface(t *testing.T) {
	var repo Repository[testUser] = NewRepo[testUser]()
	assert.NotNil(t, repo)
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(&testUser{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}
