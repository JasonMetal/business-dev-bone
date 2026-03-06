package timeutil

import "testing"

func TestStrToTime(t *testing.T) {
	type args struct {
		datetime   string
		timeLayout string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		// TODO: Add test cases.
		{"TestStrToTime", args{"2022-01-01 00:00:00", ""}, 1640966400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrToTime(tt.args.datetime, tt.args.timeLayout); got != tt.want {
				t.Errorf("StrToTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
