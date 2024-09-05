package assert

import "testing"

func Equal(t *testing.T, actual any, expected any) {
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func NotEqual(t *testing.T, actual any, expected any) {
	if expected == actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

func True(t *testing.T, actual bool) {
	if !actual {
		t.Errorf("expected %v, got %v", true, actual)
	}
}

func False(t *testing.T, actual bool) {
	if actual {
		t.Errorf("expected %v, got %v", false, actual)
	}
}

func Nil(t *testing.T, actual any) {
	if actual != nil {
		t.Errorf("expected %v, got %v", nil, actual)
	}
}

func NotNil(t *testing.T, actual any) {
	if actual == nil {
		t.Errorf("expected not %v, got %v", nil, actual)

	}
}
