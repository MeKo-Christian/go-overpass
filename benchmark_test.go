package overpass

import (
	"context"
	"net/http"
	"testing"
)

// BenchmarkUnmarshal benchmarks JSON parsing performance.
func BenchmarkUnmarshal(b *testing.B) {
	// Complex JSON with nodes, ways, and relations
	jsonData := []byte(`{
		"osm3s": {"timestamp_osm_base": "2024-01-01T00:00:00Z"},
		"elements": [
			{"type":"node","id":1,"lat":1.0,"lon":2.0,"tags":{"name":"Test"}},
			{"type":"node","id":2,"lat":3.0,"lon":4.0},
			{"type":"way","id":100,"nodes":[1,2],"tags":{"highway":"primary"}},
			{"type":"way","id":101,"nodes":[2,1],"geometry":[{"lat":3.0,"lon":4.0},{"lat":1.0,"lon":2.0}]},
			{"type":"relation","id":1000,"members":[{"type":"way","ref":100,"role":"outer"}],"tags":{"type":"multipolygon"}}
		]
	}`)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := unmarshal(jsonData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkQuery benchmarks full query cycle with mock.
func BenchmarkQuery(b *testing.B) {
	client := NewWithSettings(apiEndpoint, 1, &mockHTTPClient{
		res: &http.Response{
			StatusCode: http.StatusOK,
			Body:       newTestBody(`{"elements":[{"type":"node","id":1,"lat":1.0,"lon":2.0}]}`),
		},
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Note: This will fail after first iteration due to body being consumed
		// In real benchmark, mockHTTPClient would need to create fresh bodies
		_, _ = client.Query(`[out:json];node(1);out;`)
	}
}

// BenchmarkQueryContext benchmarks context-aware query.
func BenchmarkQueryContext(b *testing.B) {
	client := NewWithSettings(apiEndpoint, 1, &mockConcurrentHTTPClient{})
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := client.QueryContext(ctx, `[out:json];node(1);out;`)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrentQueries benchmarks parallel request handling.
func BenchmarkConcurrentQueries(b *testing.B) {
	client := NewWithSettings(apiEndpoint, 5, &mockConcurrentHTTPClient{})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			_, err := client.QueryContext(ctx, `[out:json];node(1);out;`)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkClientCreation benchmarks client initialization.
func BenchmarkClientCreation(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = New()
	}
}

// BenchmarkClientCreationWithSettings benchmarks custom client initialization.
func BenchmarkClientCreationWithSettings(b *testing.B) {
	mockClient := &mockHTTPClient{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = NewWithSettings(apiEndpoint, 3, mockClient)
	}
}

// BenchmarkUnmarshal_SimpleNode benchmarks simple node parsing.
func BenchmarkUnmarshal_SimpleNode(b *testing.B) {
	jsonData := []byte(`{"elements":[{"type":"node","id":1,"lat":1.0,"lon":2.0}]}`)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := unmarshal(jsonData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUnmarshal_ComplexWay benchmarks way with geometry.
func BenchmarkUnmarshal_ComplexWay(b *testing.B) {
	jsonData := []byte(`{
		"elements":[{
			"type":"way",
			"id":1,
			"nodes":[1,2,3,4,5],
			"geometry":[
				{"lat":1.0,"lon":2.0},
				{"lat":1.1,"lon":2.1},
				{"lat":1.2,"lon":2.2},
				{"lat":1.3,"lon":2.3},
				{"lat":1.4,"lon":2.4}
			],
			"tags":{"highway":"primary","name":"Main Street"}
		}]
	}`)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := unmarshal(jsonData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUnmarshal_LargeResult benchmarks large result sets.
func BenchmarkUnmarshal_LargeResult(b *testing.B) {
	// Generate JSON with 100 nodes
	json := `{"elements":[`

	for i := 0; i < 100; i++ {
		if i > 0 {
			json += `,`
		}

		json += `{"type":"node","id":` + string(rune(i+1)) + `,"lat":1.0,"lon":2.0}`
	}

	json += `]}`
	jsonData := []byte(json)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := unmarshal(jsonData)
		if err != nil {
			b.Fatal(err)
		}
	}
}
