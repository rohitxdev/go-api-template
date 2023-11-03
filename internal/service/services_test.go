package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/rohitxdev/go-api-template/internal/service"
)

func TestJWT(t *testing.T) {
	t.Run("Generate JWT", func(t *testing.T) {
		if _, err := service.GenerateJWT(1, time.Hour); err != nil {
			t.Error(err)
		}
	})

	t.Run("Verify JWT", func(t *testing.T) {})

}
func BenchmarkLogIn(b *testing.B) {
	for i := 0; i < b.N; i++ {
		service.LogIn(context.TODO(), "rohit@gmail.com", "rohit")
	}
	b.ReportAllocs()
}
