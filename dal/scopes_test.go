package dal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type testProduct struct {
	ID    uint    `gorm:"primarykey"`
	Name  string  `gorm:"size:255"`
	Price float64 `gorm:"size:255"`
}

func setupTestDBForScopes(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(&testProduct{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestEscapeLike(t *testing.T) {
	assert.Equal(t, "hello", escapeLike("hello"))
	assert.Equal(t, `hello\%world`, escapeLike("hello%world"))
	assert.Equal(t, `hello\_world`, escapeLike("hello_world"))
}

func TestPaginate(t *testing.T) {
	db := setupTestDBForScopes(t)
	for i := 1; i <= 50; i++ {
		db.Create(&testProduct{Name: "Product", Price: float64(i)})
	}

	scope := Paginate(1, 10)
	result := scope(db)
	var products []testProduct
	result.Find(&products)
	assert.Equal(t, 10, len(products))
}

func TestEqual(t *testing.T) {
	db := setupTestDBForScopes(t)
	db.Create(&testProduct{Name: "Apple", Price: 1.5})
	db.Create(&testProduct{Name: "Apple", Price: 2.0})

	scope := Equal("name", "Apple")
	result := scope(db)
	var products []testProduct
	result.Find(&products)
	assert.Len(t, products, 2)
}

func TestNotEqual(t *testing.T) {
	db := setupTestDBForScopes(t)
	db.Create(&testProduct{Name: "Apple", Price: 1.5})
	db.Create(&testProduct{Name: "Banana", Price: 0.5})

	scope := NotEqual("name", "Apple")
	result := scope(db)
	var products []testProduct
	result.Find(&products)
	assert.Len(t, products, 1)
}

func TestGreaterThan(t *testing.T) {
	db := setupTestDBForScopes(t)
	db.Create(&testProduct{Name: "Cheap", Price: 1.0})
	db.Create(&testProduct{Name: "Expensive", Price: 10.0})

	scope := GreaterThan("price", 5.0)
	result := scope(db)
	var products []testProduct
	result.Find(&products)
	assert.Len(t, products, 1)
}

func TestLessThan(t *testing.T) {
	db := setupTestDBForScopes(t)
	db.Create(&testProduct{Name: "Cheap", Price: 1.0})
	db.Create(&testProduct{Name: "Expensive", Price: 10.0})

	scope := LessThan("price", 5.0)
	result := scope(db)
	var products []testProduct
	result.Find(&products)
	assert.Len(t, products, 1)
}

func TestIn(t *testing.T) {
	db := setupTestDBForScopes(t)
	db.Create(&testProduct{Name: "Apple", Price: 1.5})
	db.Create(&testProduct{Name: "Banana", Price: 0.5})

	scope := In("name", []string{"Apple", "Banana"})
	result := scope(db)
	var products []testProduct
	result.Find(&products)
	assert.Len(t, products, 2)
}

func TestNotIn(t *testing.T) {
	db := setupTestDBForScopes(t)
	db.Create(&testProduct{Name: "Apple", Price: 1.5})
	db.Create(&testProduct{Name: "Banana", Price: 0.5})
	db.Create(&testProduct{Name: "Orange", Price: 2.0})

	scope := NotIn("name", []string{"Apple", "Banana"})
	result := scope(db)
	var products []testProduct
	result.Find(&products)
	assert.Len(t, products, 1)
}

func TestBetween(t *testing.T) {
	db := setupTestDBForScopes(t)
	db.Create(&testProduct{Name: "Item1", Price: 1.0})
	db.Create(&testProduct{Name: "Item2", Price: 5.0})
	db.Create(&testProduct{Name: "Item3", Price: 10.0})

	scope := Between("price", 1.0, 5.0)
	result := scope(db)
	var products []testProduct
	result.Find(&products)
	assert.Len(t, products, 2)
}

func TestOrderBy(t *testing.T) {
	db := setupTestDBForScopes(t)
	db.Create(&testProduct{Name: "C", Price: 3.0})
	db.Create(&testProduct{Name: "A", Price: 1.0})
	db.Create(&testProduct{Name: "B", Price: 2.0})

	scope := OrderBy("name", "asc")
	result := scope(db)
	var products []testProduct
	result.Find(&products)
	assert.Equal(t, "A", products[0].Name)
}

func TestLimit(t *testing.T) {
	db := setupTestDBForScopes(t)
	db.Create(&testProduct{Name: "Item1", Price: 1.0})
	db.Create(&testProduct{Name: "Item2", Price: 2.0})
	db.Create(&testProduct{Name: "Item3", Price: 3.0})

	scope := Limit(2)
	result := scope(db)
	var products []testProduct
	result.Find(&products)
	assert.Len(t, products, 2)
}

func TestLike(t *testing.T) {
	db := setupTestDBForScopes(t)
	db.Create(&testProduct{Name: "Apple", Price: 1.5})
	db.Create(&testProduct{Name: "ApplePie", Price: 5.0})

	scope := Like("name", "Apple")
	result := scope(db)
	var products []testProduct
	result.Find(&products)
	assert.Len(t, products, 2)
}

func TestPaginationConstants(t *testing.T) {
	assert.Equal(t, 10, DefaultPageSize)
	assert.Equal(t, 100, MaxPageSize)
}