package v1

import (
	"reflect"
	"testing"

	"github.com/AlekSi/pointer"
)

func TestUnpointer(t *testing.T) {
	type args struct {
		offset   *int64
		limit    *int64
		pageNo   *int64
		pageSize *int64
	}
	tests := []struct {
		name string
		args args
		want *LimitAndOffset
	}{
		{
			name: "both offset and limit are not zero",
			args: args{
				offset: pointer.ToInt64(0),
				limit:  pointer.ToInt64(10),
			},
			want: &LimitAndOffset{
				Offset: 0,
				Limit:  10,
			},
		},
		{
			name: "both offset and limit are zero",
			want: &LimitAndOffset{
				Offset: 0,
				Limit:  1000,
			},
		},
		{
			name: "offset not zero and limit zero",
			args: args{
				offset: pointer.ToInt64(2),
			},
			want: &LimitAndOffset{
				Offset: 2,
				Limit:  1000,
			},
		},
		{
			name: "offset zero and limit not zero",
			args: args{
				limit: pointer.ToInt64(10),
			},
			want: &LimitAndOffset{
				Offset: 0,
				Limit:  10,
			},
		},
		{
			name: "both pageNo and pageSize are not zero",
			args: args{
				pageNo:   pointer.ToInt64(2),
				pageSize: pointer.ToInt64(10),
			},
			want: &LimitAndOffset{
				Offset: 10,
				Limit:  10,
			},
		},
		{
			name: "both pageNo and pageSize are zero",
			want: &LimitAndOffset{
				Offset: 0,
				Limit:  1000,
			},
		},
		{
			name: "pageNo not zero and pageSize zero",
			args: args{
				pageNo: pointer.ToInt64(1),
			},
			want: &LimitAndOffset{
				Offset: 0,
				Limit:  1000,
			},
		},
		{
			name: "pageNo zero and pageSize not zero",
			args: args{
				pageSize: pointer.ToInt64(10),
			},
			want: &LimitAndOffset{
				Offset: 0,
				Limit:  10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := ListOptions{Offset: tt.args.offset, Limit: tt.args.limit, PageNo: tt.args.pageNo, PageSize: tt.args.pageSize}
			if got := Unpointer(opt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unpointer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func FuzzUnpointer(f *testing.F) {
	testcases := []int64{1, 2, 3, 4, 5}
	for _, tc := range testcases {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}
	f.Fuzz(func(t *testing.T, in int64) {
		out := Unpointer(ListOptions{Offset: pointer.ToInt64(0), Limit: &in})
		want := &LimitAndOffset{
			Offset: 0,
			Limit:  int(in),
		}
		if !reflect.DeepEqual(out, want) {
			t.Errorf("got: %v, want: %v", out, want)
		}
	})
}
