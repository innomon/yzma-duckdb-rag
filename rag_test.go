package main

import (
	"math"
	"testing"
)

const epsilon = 1e-6

func TestNormalizeVector(t *testing.T) {
	vec := []float32{3, 4}
	result := normalizeVector(vec)

	if len(result) != 2 {
		t.Fatalf("expected length 2, got %d", len(result))
	}
	if math.Abs(float64(result[0])-0.6) > epsilon {
		t.Errorf("expected result[0] ≈ 0.6, got %f", result[0])
	}
	if math.Abs(float64(result[1])-0.8) > epsilon {
		t.Errorf("expected result[1] ≈ 0.8, got %f", result[1])
	}

	// Verify magnitude is ~1.0
	var mag float64
	for _, v := range result {
		mag += float64(v * v)
	}
	mag = math.Sqrt(mag)
	if math.Abs(mag-1.0) > epsilon {
		t.Errorf("expected magnitude ≈ 1.0, got %f", mag)
	}
}

func TestNormalizeVector_ZeroVector(t *testing.T) {
	vec := []float32{0, 0, 0}
	result := normalizeVector(vec)

	if len(result) != 3 {
		t.Fatalf("expected length 3, got %d", len(result))
	}
	for i, v := range result {
		if v != 0 {
			t.Errorf("expected result[%d] = 0, got %f", i, v)
		}
	}
}

func TestNormalizeVector_SingleElement(t *testing.T) {
	vec := []float32{5}
	result := normalizeVector(vec)

	if len(result) != 1 {
		t.Fatalf("expected length 1, got %d", len(result))
	}
	if math.Abs(float64(result[0])-1.0) > epsilon {
		t.Errorf("expected result[0] ≈ 1.0, got %f", result[0])
	}
}

func TestNormalizeVector_Empty(t *testing.T) {
	vec := []float32{}
	result := normalizeVector(vec)

	if len(result) != 0 {
		t.Fatalf("expected length 0, got %d", len(result))
	}
}

func TestFloatArrayToSQL(t *testing.T) {
	arr := []float32{1.0, 2.0, 3.0}
	result := floatArrayToSQL(arr)
	expected := "[1.000000, 2.000000, 3.000000]"

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFloatArrayToSQL_Empty(t *testing.T) {
	arr := []float32{}
	result := floatArrayToSQL(arr)
	expected := "[]"

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFloatArrayToSQL_Single(t *testing.T) {
	arr := []float32{42.5}
	result := floatArrayToSQL(arr)
	expected := "[42.500000]"

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTruncate_ShortString(t *testing.T) {
	result := truncate("hello", 10)
	if result != "hello" {
		t.Errorf("expected %q, got %q", "hello", result)
	}
}

func TestTruncate_ExactLength(t *testing.T) {
	result := truncate("hello", 5)
	if result != "hello" {
		t.Errorf("expected %q, got %q", "hello", result)
	}
}

func TestTruncate_LongString(t *testing.T) {
	result := truncate("hello world", 8)
	expected := "hello..."
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTruncate_Empty(t *testing.T) {
	result := truncate("", 10)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}
