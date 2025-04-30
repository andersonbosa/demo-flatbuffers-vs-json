package main

import (
	"testing"
)

func BenchmarkJSON(b *testing.B) {
	data, _ := generateUserData(100_000)
	for i := 0; i < b.N; i++ {
		benchmarkJson(data)
	}
}

func BenchmarkFlatBuffers(b *testing.B) {
	_, names := generateUserData(100_000)
	for i := 0; i < b.N; i++ {
		benchmarkFlatBuffers(names)
	}
}
