package groundcontrol

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeID(t *testing.T) {
	type args struct {
		identifiers []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		"one string",
		args{[]string{"User"}},
		"VXNlcg==",
	}, {
		"two string",
		args{[]string{"User", "0"}},
		"VXNlcjow",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeID(tt.args.identifiers...); got != tt.want {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}

func TestDecodeID(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{{
		"one string",
		args{"VXNlcg=="},
		[]string{"User"},
		false,
	}, {
		"two string",
		args{"VXNlcjow"},
		[]string{"User", "0"},
		false,
	}, {
		"invalid id",
		args{"*"},
		nil,
		true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeID(tt.args.id)
			if (err != nil) != tt.wantErr {
				assert.Equal(t, got, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}
