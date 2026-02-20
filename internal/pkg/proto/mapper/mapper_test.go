package mapper

import (
	"gophkeeper/internal/dto/model"
	pb "gophkeeper/internal/pkg/proto"
	"reflect"
	"testing"
)

func TestProtoSecertInSecert(t *testing.T) {
	type args struct {
		pbSecret *pb.Secret
	}
	tests := []struct {
		name string
		args args
		want *model.Secret
	}{
		{
			name: "TestProtoSecertInSecert",
			args: args{
				pbSecret: pb.Secret_builder{
					Name:          "secret",
					SecretType:    pb.SecretType_SECRET_TYPE_TEXT,
					Description:   "secret description",
					EncryptedData: []byte{1, 2, 3},
				}.Build(),
			},
			want: &model.Secret{
				Name:        "secret",
				SecretType:  "SECRET_TYPE_TEXT",
				Description: "secret description",
				Data:        []byte{1, 2, 3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ProtoSecertInSecert(tt.args.pbSecret); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProtoSecertInSecert() = %v, want %v", got, tt.want)
			}
		})
	}
}
