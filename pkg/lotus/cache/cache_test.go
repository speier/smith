package cache

import (
	"testing"
)

func TestCSSCaching(t *testing.T) {
	Clear()

	css1 := `
		#test {
			color: #fff;
			background: #000;
		}
	`

	styles1 := GetStyles(css1)
	if styles1 == nil {
		t.Fatal("Expected styles to be parsed")
	}

	if cacheSize := Size(); cacheSize != 1 {
		t.Errorf("Expected cache to have 1 entry, got %d", cacheSize)
	}

	styles2 := GetStyles(css1)
	if styles2 == nil {
		t.Fatal("Expected styles to be parsed")
	}

	if cacheSize := Size(); cacheSize != 1 {
		t.Errorf("Expected cache to still have 1 entry, got %d", cacheSize)
	}

	css2 := `
		#other {
			color: #f00;
		}
	`
	styles3 := GetStyles(css2)
	if styles3 == nil {
		t.Fatal("Expected styles to be parsed")
	}

	if cacheSize := Size(); cacheSize != 2 {
		t.Errorf("Expected cache to have 2 entries, got %d", cacheSize)
	}
}

func TestSetEnabled(t *testing.T) {
	Clear()
	css := `#test { color: #fff; }`
	SetEnabled(true)
	_ = GetStyles(css)
	if cacheSize := Size(); cacheSize != 1 {
		t.Errorf("Expected cache to have 1 entry with caching enabled, got %d", cacheSize)
	}
	SetEnabled(false)
	if cacheSize := Size(); cacheSize != 0 {
		t.Errorf("Expected cache to be cleared when disabled, got %d entries", cacheSize)
	}
	_ = GetStyles(css)
	if cacheSize := Size(); cacheSize != 0 {
		t.Errorf("Expected cache to remain empty with caching disabled, got %d", cacheSize)
	}
	SetEnabled(true)
}

func TestClear(t *testing.T) {
	css := `#test { color: #fff; }`
	_ = GetStyles(css)
	if cacheSize := Size(); cacheSize == 0 {
		t.Error("Expected cache to have entries before clearing")
	}
	Clear()
	if cacheSize := Size(); cacheSize != 0 {
		t.Errorf("Expected cache to be empty after clearing, got %d entries", cacheSize)
	}
}

func BenchmarkCachedCSS(b *testing.B) {
	Clear()
	SetEnabled(true)
	css := `
		#root { display: flex; flex-direction: column; }
		#header { height: 3; color: #5af; }
		#messages { flex: 1; color: #ddd; }
		#input { height: 3; color: #fff; }
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetStyles(css)
	}
}

func BenchmarkUncachedCSS(b *testing.B) {
	SetEnabled(false)
	defer SetEnabled(true)
	css := `
		#root { display: flex; flex-direction: column; }
		#header { height: 3; color: #5af; }
		#messages { flex: 1; color: #ddd; }
		#input { height: 3; color: #fff; }
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetStyles(css)
	}
}
