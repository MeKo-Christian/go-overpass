package overpass

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"
	"testing/iotest"
)

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string
		want  Result
	}{
		{
			`{"elements":[{"type":"way","id":1,
				"bounds":{"minlat":-37.9,"minlon":144.6,"maxlat":-37.8,"maxlon":144.7}
			}]}`,
			Result{
				Count: 1,
				Ways: map[int64]*Way{1: {
					Bounds: &Box{Min: Point{-37.9, 144.6}, Max: Point{-37.8, 144.7}},
				}},
			},
		},
		{
			`{"elements":[{"type":"way","id":1,
				"geometry":[{"lat":-37.9,"lon":144.6},{"lat":-37.8,"lon":144.7}]
			}]}`,
			Result{
				Count: 1,
				Ways: map[int64]*Way{1: {
					Geometry: []Point{{-37.9, 144.6}, {-37.8, 144.7}},
				}},
			},
		},
		{
			`{"elements":[{"type":"relation","id":1,
				"bounds":{"minlat":-37.9,"minlon":144.6,"maxlat":-37.8,"maxlon":144.7}
			}]}`,
			Result{
				Count: 1,
				Relations: map[int64]*Relation{1: {
					Bounds: &Box{Min: Point{-37.9, 144.6}, Max: Point{-37.8, 144.7}},
				}},
			},
		},
	}

	for i, testCase := range testCases {
		testCase := testCase // capture range variable

		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			t.Parallel()

			got, err := unmarshal([]byte(testCase.input))
			if err != nil {
				t.Fatal(err)
			}

			if testCase.want.Nodes == nil {
				testCase.want.Nodes = map[int64]*Node{}
			} else {
				for id, n := range testCase.want.Nodes {
					n.ID = id
				}
			}

			if testCase.want.Ways == nil {
				testCase.want.Ways = map[int64]*Way{}
			} else {
				for id, way := range testCase.want.Ways {
					way.ID = id
					if way.Nodes == nil {
						way.Nodes = []*Node{}
					}

					if way.Geometry == nil {
						way.Geometry = []Point{}
					}
				}
			}

			if testCase.want.Relations == nil {
				testCase.want.Relations = map[int64]*Relation{}
			} else {
				for id, r := range testCase.want.Relations {
					r.ID = id
					if r.Members == nil {
						r.Members = []RelationMember{}
					}
				}
			}

			if !reflect.DeepEqual(got, testCase.want) {
				t.Fatalf("%v != %v", got, testCase.want)
			}
		})
	}
}

func TestQueryErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		res  *http.Response
		err  error
		want string
	}{
		{
			nil,
			errors.New("request fail"),
			"http error: request fail",
		},
		{
			&http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(bytes.NewReader(nil))},
			nil,
			"overpass engine error: 400 Bad Request",
		},
		{
			&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(iotest.ErrReader(errors.New("read fail"))),
			},
			nil,
			"http error: read fail",
		},
		{
			&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(nil))},
			nil,
			"overpass engine error: unexpected end of JSON input",
		},
	}

	for i, testCase := range testCases {
		testCase := testCase // capture range variable

		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			t.Parallel()

			cli := NewWithSettings(apiEndpoint, 1, &mockHTTPClient{testCase.res, testCase.err})

			_, err := cli.Query("")
			if err == nil {
				t.Fatal("unexpected success")
			}

			if err.Error() != testCase.want {
				t.Fatalf("%s != %s", err.Error(), testCase.want)
			}

			if errors.Unwrap(err) == nil {
				t.Fatal("expected wrapped error")
			}
		})
	}
}

type mockHTTPClient struct {
	res *http.Response
	err error
}

func (m *mockHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	return m.res, m.err
}

// newTestBody creates an io.ReadCloser from a string for testing.
func newTestBody(s string) io.ReadCloser {
	return io.NopCloser(bytes.NewReader([]byte(s)))
}
