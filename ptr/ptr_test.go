package ptr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTo(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		val := 42
		result := To(val)
		assert.NotNil(t, result)
		assert.Equal(t, val, *result)
	})

	t.Run("string", func(t *testing.T) {
		val := "hello"
		result := To(val)
		assert.NotNil(t, result)
		assert.Equal(t, val, *result)
	})

	t.Run("struct", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		val := Person{Name: "Alice", Age: 30}
		result := To(val)
		assert.NotNil(t, result)
		assert.Equal(t, val, *result)
	})

	t.Run("modification doesn't affect original", func(t *testing.T) {
		val := 42
		result := To(val)
		*result = 100
		assert.Equal(t, 42, val)
	})
}

func TestDerefOr(t *testing.T) {
	t.Run("non-nil pointer", func(t *testing.T) {
		val := 42
		result := DerefOr(&val, 0)
		assert.Equal(t, 42, result)
	})

	t.Run("nil pointer", func(t *testing.T) {
		var val *int = nil
		result := DerefOr(val, 100)
		assert.Equal(t, 100, result)
	})

	t.Run("with string", func(t *testing.T) {
		str := "hello"
		result := DerefOr(&str, "default")
		assert.Equal(t, "hello", result)
	})

	t.Run("nil string pointer", func(t *testing.T) {
		var str *string = nil
		result := DerefOr(str, "default")
		assert.Equal(t, "default", result)
	})
}

func TestToIf(t *testing.T) {
	t.Run("condition true", func(t *testing.T) {
		result := ToIf(true, 42)
		assert.NotNil(t, result)
		assert.Equal(t, 42, *result)
	})

	t.Run("condition false", func(t *testing.T) {
		result := ToIf(false, 42)
		assert.Nil(t, result)
	})

	t.Run("with string", func(t *testing.T) {
		result := ToIf(true, "hello")
		assert.NotNil(t, result)
		assert.Equal(t, "hello", *result)
	})

	t.Run("false with string", func(t *testing.T) {
		result := ToIf(false, "hello")
		assert.Nil(t, result)
	})
}

func TestIsNil(t *testing.T) {
	t.Run("nil pointer", func(t *testing.T) {
		var val *int = nil
		assert.True(t, IsNil(val))
	})

	t.Run("non-nil pointer", func(t *testing.T) {
		val := 42
		assert.False(t, IsNil(&val))
	})

	t.Run("nil string pointer", func(t *testing.T) {
		var str *string = nil
		assert.True(t, IsNil(str))
	})
}

func TestIsNotNil(t *testing.T) {
	t.Run("nil pointer", func(t *testing.T) {
		var val *int = nil
		assert.False(t, IsNotNil(val))
	})

	t.Run("non-nil pointer", func(t *testing.T) {
		val := 42
		assert.True(t, IsNotNil(&val))
	})

	t.Run("nil string pointer", func(t *testing.T) {
		var str *string = nil
		assert.False(t, IsNotNil(str))
	})
}

func TestZero(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		result := Zero[int]()
		assert.NotNil(t, result)
		assert.Equal(t, 0, *result)
	})

	t.Run("string", func(t *testing.T) {
		result := Zero[string]()
		assert.NotNil(t, result)
		assert.Equal(t, "", *result)
	})

	t.Run("bool", func(t *testing.T) {
		result := Zero[bool]()
		assert.NotNil(t, result)
		assert.Equal(t, false, *result)
	})

	t.Run("struct", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		result := Zero[Person]()
		assert.NotNil(t, result)
		assert.Equal(t, Person{}, *result)
	})

	t.Run("slice", func(t *testing.T) {
		result := Zero[[]int]()
		assert.NotNil(t, result)
		assert.Nil(t, *result)
	})

	t.Run("map", func(t *testing.T) {
		result := Zero[map[string]int]()
		assert.NotNil(t, result)
		assert.Nil(t, *result)
	})
}

func TestDeref(t *testing.T) {
	t.Run("non-nil pointer", func(t *testing.T) {
		val := 42
		result := Deref(&val)
		assert.Equal(t, 42, result)
	})

	t.Run("nil pointer", func(t *testing.T) {
		var val *int = nil
		result := Deref(val)
		assert.Equal(t, 0, result)
	})

	t.Run("with string", func(t *testing.T) {
		str := "hello"
		result := Deref(&str)
		assert.Equal(t, "hello", result)
	})

	t.Run("nil string pointer", func(t *testing.T) {
		var str *string = nil
		result := Deref(str)
		assert.Equal(t, "", result)
	})

	t.Run("struct", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		person := Person{Name: "Alice", Age: 30}
		result := Deref(&person)
		assert.Equal(t, person, result)
	})

	t.Run("nil struct pointer", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		var person *Person = nil
		result := Deref(person)
		assert.Equal(t, Person{}, result)
	})
}

func TestTo_Generics(t *testing.T) {
	t.Run("float64", func(t *testing.T) {
		result := To(3.14)
		assert.NotNil(t, result)
		assert.Equal(t, 3.14, *result)
	})

	t.Run("bool", func(t *testing.T) {
		result := To(true)
		assert.NotNil(t, result)
		assert.Equal(t, true, *result)
	})

	t.Run("byte", func(t *testing.T) {
		result := To(byte('A'))
		assert.NotNil(t, result)
		assert.Equal(t, byte('A'), *result)
	})

	t.Run("rune", func(t *testing.T) {
		result := To(rune('中'))
		assert.NotNil(t, result)
		assert.Equal(t, rune('中'), *result)
	})

	t.Run("uintptr", func(t *testing.T) {
		result := To(uintptr(12345))
		assert.NotNil(t, result)
		assert.Equal(t, uintptr(12345), *result)
	})

	t.Run("complex64", func(t *testing.T) {
		result := To(complex(1, 2))
		assert.NotNil(t, result)
		assert.Equal(t, complex(1, 2), *result)
	})
}

func TestDerefOr_Generics(t *testing.T) {
	t.Run("float64", func(t *testing.T) {
		val := 3.14
		result := DerefOr(&val, 0.0)
		assert.Equal(t, 3.14, result)
	})

	t.Run("nil float64", func(t *testing.T) {
		var val *float64 = nil
		result := DerefOr(val, 2.71)
		assert.Equal(t, 2.71, result)
	})

	t.Run("bool", func(t *testing.T) {
		val := true
		result := DerefOr(&val, false)
		assert.Equal(t, true, result)
	})

	t.Run("nil bool", func(t *testing.T) {
		var val *bool = nil
		result := DerefOr(val, true)
		assert.Equal(t, true, result)
	})
}

func BenchmarkTo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		To(42)
	}
}

func BenchmarkDerefOr_NonNil(b *testing.B) {
	val := 42
	for i := 0; i < b.N; i++ {
		DerefOr(&val, 0)
	}
}

func BenchmarkDerefOr_Nil(b *testing.B) {
	var val *int = nil
	for i := 0; i < b.N; i++ {
		DerefOr(val, 0)
	}
}

func BenchmarkZero(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Zero[int]()
	}
}

func BenchmarkDeref_NonNil(b *testing.B) {
	val := 42
	for i := 0; i < b.N; i++ {
		Deref(&val)
	}
}

func BenchmarkDeref_Nil(b *testing.B) {
	var val *int = nil
	for i := 0; i < b.N; i++ {
		Deref(val)
	}
}