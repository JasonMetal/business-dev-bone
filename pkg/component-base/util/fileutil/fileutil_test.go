package fileutil

import "testing"

func TestCheckPathIsDot(t *testing.T) {
	testCases := []struct {
		path   string
		result bool
	}{
		{".", true},
		{"./", true},
		{"test", false},
		{"../", false},
		{".\\", false},
	}

	for _, testCase := range testCases {
		result := CheckPathIsDot(testCase.path)
		if result != testCase.result {
			t.Fatalf("check path is dot failed")
		}
	}
}
