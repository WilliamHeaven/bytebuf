package bytebuf

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

var sinkString string

func BenchmarkString(b *testing.B) {
	sizes := []int{0, 5, 64, 1024}

	for _, size := range sizes {
		arg := strings.Repeat("x", size)
		name := "empty"
		if arg != "" {
			name = fmt.Sprint(len(arg))
		}

		b.Run("bytebuf.Buffer/"+name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buf := New()
				buf.WriteString(arg)
				sinkString = buf.String()
			}
		})
		b.Run("bytes.Buffer/"+name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buf := bytes.NewBuffer(make([]byte, 0, 64))
				buf.WriteString(arg)
				sinkString = buf.String()
			}
		})
	}
}

func BenchmarkStringPointerBuffer(b *testing.B) {
	sizes := []int{0, 5, 64, 1024}

	for _, size := range sizes {
		arg := strings.Repeat("x", size)
		name := "empty"
		if arg != "" {
			name = fmt.Sprint(len(arg))
		}

		b.Run("bytebuf.Buffer/"+name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buf := NewPointer()
				buf.WriteString(arg)
				sinkString = buf.String()
			}
		})
		b.Run("bytes.Buffer/"+name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buf := new(bytes.Buffer)
				buf.WriteString(arg)
				sinkString = buf.String()
			}
		})
	}
}
