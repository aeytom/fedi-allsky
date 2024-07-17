package allsky

import (
	"testing"
	"time"
)

func Test_mustParseAngle(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		// TODO: Add test cases.
		{
			name: "full",
			args: args{
				val: `41deg 23' 30.2"`,
			},
			want: 41.39172222222222,
		},
		{
			name: "negative",
			args: args{
				val: `-06deg 15' 33.1"`,
			},
			want: -6.259194444444445,
		},
		{
			name: "large",
			args: args{
				val: `334deg 00' 42.5"`,
			},
			want: 334.01180555555555,
		},
		{
			name: "short",
			args: args{
				val: `-11.07°`,
			},
			want: -11.07,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mustParseAngle(tt.args.val); got != tt.want {
				t.Errorf("mustParseAngle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mustParseInt(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "null",
			args: args{
				val: "0",
			},
			want: 0,
		},
		{
			name: "negative",
			args: args{
				val: "-53",
			},
			want: -53,
		},
		{
			name: "high exposure",
			args: args{
				val: "30000000",
			},
			want: 30000000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mustParseInt(tt.args.val); got != tt.want {
				t.Errorf("mustParseInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mustParseBool(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "",
			args: args{
				val: "Yes",
			},
			want: true,
		},
		{
			name: "",
			args: args{
				val: "No",
			},
			want: false,
		},
		{
			name: "",
			args: args{
				val: "True",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mustParseBool(tt.args.val); got != tt.want {
				t.Errorf("mustParseBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mustParseDateTime(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "sunrise",
			args: args{
				val: "20240714 05:03:17",
			},
			want: "2024-07-14T05:03:17+02:00",
		},
		{
			name: "sunset",
			args: args{
				val: "20240713 21:23:11",
			},
			want: "2024-07-13T21:23:11+02:00",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, _ := time.Parse(time.RFC3339, tt.want)
			if got := mustParseDateTime(tt.args.val); !got.Equal(want) {
				t.Errorf("mustParseDateTime() = %v, want %v", got, want)
			}
		})
	}
}

func Test_mustParseDuration(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			name: "",
			args: args{
				val: "30000000us",
			},
			want: 30 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mustParseDuration(tt.args.val); got != tt.want {
				t.Errorf("mustParseDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
