package allsky

import "testing"

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
