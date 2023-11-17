package util_test

import (
	"bytes"
	"testing"

	"github.com/rohitxdev/go-api-template/internal/util"
)

func TestReverseString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Case 1", args: args{s: "123456789"}, want: "987654321"},
		{name: "Case 2", args: args{s: "John Jonah Jameson"}, want: "nosemaJ hanoJ nhoJ"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := util.ReverseString(tt.args.s); got != tt.want {
				t.Errorf("ReverseString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAES(t *testing.T) {
	key := []byte("1234567812345678")
	plainText := []byte("Lorem ipsum dolor sit amet, consectetur adipisicing elit. Iusto itaque error, voluptates molestiae at consequuntur minima, doloremque consequatur dolores ipsam voluptatem quaerat aliquid, adipisci rem est quia nobis ducimus neque distinctio debitis. Quo exercitationem earum, possimus velit non ullam tempora, architecto maxime rerum accusantium aliquam. Fugit laborum omnis non distinctio.")

	encryptedData := util.EncryptAES(plainText, key)
	decryptedData := util.DecryptAES(encryptedData, key)

	if !bytes.Equal(plainText, decryptedData) {
		t.Error("decryption failed")
	}
}
