package util_test

import (
	"testing"

	"github.com/rohitxdev/go-api-template/internal/util"
)

func BenchmarkReverseLong(b *testing.B) {
	for i := 0; i < b.N; i++ {
		util.ReverseString("My name is Walter Hartwell White. I live at 308 Negra Arroyo Lane, Albuquerque, New Mexico, 87104. This is my confession. If you're watching this tape, I'm probably dead- murdered by my brother-in-law, Hank Schrader. Hank has been building a meth empire for over a year now, and using me as his chemist. Shortly after my 50th birthday, he asked that I use my chemistry knowledge to cook methamphetamine, which he would then sell using connections that he made through his career with the DEA.")
	}
}
