package dal

import (
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

func TestNew(t *testing.T) {
	repo := New[testUser]()
	assert.NotNil(t, repo)
}

func TestRepo_Insert(t *testing.T) {
	db := setupTestDB(t)
	repo := New[testUser]()

	user := &testUser{Name: "Alice", Age: 25}
	err := repo.Insert(t.Context(), db, user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
}

func TestRepo_Insert_NilDB(t *testing.T) {
	repo := New[testUser]()
	err := repo.Insert(t.Context(), nil, &testUser{})
	assert.Error(t, err)
}

func TestRepo_Insert_NilValue(t *testing.T) {
	db := setupTestDB(t)
	repo := New[testUser]()
	err := repo.Insert(t.Context(), db, nil)
	assert.Error(t, err)
}

func TestRepo_BatchInsert(t *testing.T) {
	db := setupTestDB(t)
	repo := New[testUser]()

	users := []*testUser{
		{Name: "User1", Age: 20},
		{Name: "User2", Age: 21},
	}
	err := repo.BatchInsert(t.Context(), db, users, 2)
	assert.NoError(t, err)
}

func TestRepo_UpdateFields(t *testing.T) {
	db := setupTestDB(t)
	repo := New[testUser]()

	user := &testUser{Name: "Test", Age: 25}
	require.NoError(t, db.Create(user).Error)

	updates := map[string]any{"name": "NewName"}
	err := repo.UpdateFields(t.Context(), db, updates, Equal("id", user.ID))
	assert.NoError(t, err)
}

func TestRepo_Update_UsesDBConditionsWithoutScopes(t *testing.T) {
	db := setupTestDB(t)
	repo := New[testUser]()

	user := &testUser{Name: "Before", Age: 25}
	require.NoError(t, db.Create(user).Error)

	err := repo.Update(t.Context(), db.Where("id = ?", user.ID), &testUser{Name: "After"})
	require.NoError(t, err)

	var result testUser
	require.NoError(t, db.First(&result, user.ID).Error)
	assert.Equal(t, "After", result.Name)
}

func TestRepo_Update_BareDBIsRejectedByGORM(t *testing.T) {
	db := setupTestDB(t)
	repo := New[testUser]()

	require.NoError(t, db.Create(&testUser{Name: "Before", Age: 25}).Error)

	err := repo.Update(t.Context(), db, &testUser{Name: "Unsafe"})
	assert.Error(t, err)
}

func TestRepo_UpdateFields_UsesDBConditionsWithoutScopes(t *testing.T) {
	db := setupTestDB(t)
	repo := New[testUser]()

	user := &testUser{Name: "Before", Age: 25}
	require.NoError(t, db.Create(user).Error)

	err := repo.UpdateFields(t.Context(), db.Where("id = ?", user.ID), map[string]any{"name": "After"})
	require.NoError(t, err)

	var result testUser
	require.NoError(t, db.First(&result, user.ID).Error)
	assert.Equal(t, "After", result.Name)
}

func TestRepo_UpdateFields_BareDBIsRejectedByGORM(t *testing.T) {
	db := setupTestDB(t)
	repo := New[testUser]()

	require.NoError(t, db.Create(&testUser{Name: "Before", Age: 25}).Error)

	err := repo.UpdateFields(t.Context(), db, map[string]any{"name": "Unsafe"})
	assert.Error(t, err)
}

func TestRepo_QueryOne(t *testing.T) {
	db := setupTestDB(t)
	repo := New[testUser]()

	require.NoError(t, db.Create(&testUser{Name: "FindMe", Age: 20}).Error)

	result, err := repo.QueryOne(t.Context(), db, Equal("name", "FindMe"))
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "FindMe", result.Name)
}

func TestRepo_QueryOne_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := New[testUser]()

	result, err := repo.QueryOne(t.Context(), db, Equal("name", "NonExistent"))
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestRepo_QueryOne_UsesExplicitOrder(t *testing.T) {
	db := setupTestDB(t)
	repo := New[testUser]()

	require.NoError(t, db.Create(&testUser{Name: "Young", Age: 20}).Error)
	require.NoError(t, db.Create(&testUser{Name: "Old", Age: 30}).Error)

	result, err := repo.QueryOne(t.Context(), db, Order("age", "desc"))
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Old", result.Name)
}

func TestRepo_Query(t *testing.T) {
	db := setupTestDB(t)
	repo := New[testUser]()

	require.NoError(t, db.Create(&testUser{Name: "User1", Age: 20}).Error)
	require.NoError(t, db.Create(&testUser{Name: "User2", Age: 25}).Error)

	results, err := repo.Query(t.Context(), db)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestRepo_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := New[testUser]()

	require.NoError(t, db.Create(&testUser{Name: "User1", Age: 20}).Error)

	count, err := repo.Count(t.Context(), db)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := New[testUser]()

	require.NoError(t, db.Create(&testUser{Name: "DeleteMe", Age: 20}).Error)

	err := repo.Delete(t.Context(), db, Equal("name", "DeleteMe"))
	assert.NoError(t, err)
}

func TestRepo_Delete_UsesDBConditionsWithoutScopes(t *testing.T) {
	db := setupTestDB(t)
	repo := New[testUser]()

	require.NoError(t, db.Create(&testUser{Name: "DeleteMe", Age: 20}).Error)

	err := repo.Delete(t.Context(), db.Where("name = ?", "DeleteMe"))
	require.NoError(t, err)

	var count int64
	require.NoError(t, db.Model(&testUser{}).Where("name = ?", "DeleteMe").Count(&count).Error)
	assert.Equal(t, int64(0), count)
}

func TestRepo_Delete_BareDBIsRejectedByGORM(t *testing.T) {
	db := setupTestDB(t)
	repo := New[testUser]()

	require.NoError(t, db.Create(&testUser{Name: "DeleteMe", Age: 20}).Error)

	err := repo.Delete(t.Context(), db)
	assert.Error(t, err)
}

func TestRepo_Raw(t *testing.T) {
	db := setupTestDB(t)
	repo := New[testUser]()

	require.NoError(t, db.Create(&testUser{Name: "Raw1", Age: 20}).Error)

	results, err := repo.Raw(t.Context(), db, "SELECT * FROM test_users WHERE name = ?", "Raw1")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestExec(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.Create(&testUser{Name: "ToDelete", Age: 20}).Error)
	err := Exec(t.Context(), db, "DELETE FROM test_users WHERE name = ?", "ToDelete")
	assert.NoError(t, err)
}

func TestExec_AllowsZeroRowsAffected(t *testing.T) {
	db := setupTestDB(t)
	err := Exec(t.Context(), db, "DELETE FROM test_users WHERE name = ?", "missing")
	require.NoError(t, err)
}

func TestExec_AllowsDDL(t *testing.T) {
	db := setupTestDB(t)
	err := Exec(t.Context(), db, "CREATE TABLE audit_logs (id integer primary key, message text)")
	require.NoError(t, err)
}

func TestExec_NilDB(t *testing.T) {
	err := Exec(t.Context(), nil, "DELETE FROM test_users")
	assert.Error(t, err)
}

func TestRepositoryInterface(t *testing.T) {
	var repo Repository[testUser] = New[testUser]()
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
