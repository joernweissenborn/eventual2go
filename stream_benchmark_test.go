package eventual2go

import "testing"

func BenchmarkStream(b *testing.B) {
	sc := NewStreamController()
	c, _ := sc.Stream().AsChan()

	for i := 0; i < b.N; i++ {
		sc.Add("test")
		<-c
	}
}
