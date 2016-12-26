package eventual2go

import "testing"

func BenchmarkStream(b *testing.B) {
	sc := NewStreamController()
	defer sc.Close()
	c := sc.Stream().AsChan()

	for i := 0; i < b.N; i++ {
		sc.Add("test")
		<-c
	}
}
