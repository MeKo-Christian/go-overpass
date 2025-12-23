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

	for i, tc := range testCases {
		tc := tc // capture range variable

		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			t.Parallel()

			got, err := unmarshal([]byte(tc.input))
			if err != nil {
				t.Fatal(err)
			}

			if tc.want.Nodes == nil {
				tc.want.Nodes = map[int64]*Node{}
			} else {
				for id, n := range tc.want.Nodes {
					n.ID = id
				}
			}

			if tc.want.Ways == nil {
				tc.want.Ways = map[int64]*Way{}
			} else {
				for id, w := range tc.want.Ways {
					w.ID = id
					if w.Nodes == nil {
						w.Nodes = []*Node{}
					}

					if w.Geometry == nil {
						w.Geometry = []Point{}
					}
				}
			}

			if tc.want.Relations == nil {
				tc.want.Relations = map[int64]*Relation{}
			} else {
				for id, r := range tc.want.Relations {
					r.ID = id
					if r.Members == nil {
						r.Members = []RelationMember{}
					}
				}
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("%v != %v", got, tc.want)
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

	for i, tc := range testCases {
		tc := tc // capture range variable

		t.Run(fmt.Sprintf("test case %d", i), func(t *testing.T) {
			t.Parallel()

			cli := NewWithSettings(apiEndpoint, 1, &mockHTTPClient{tc.res, tc.err})

			_, err := cli.Query("")
			if err == nil {
				t.Fatal("unexpected success")
			}

			if err.Error() != tc.want {
				t.Fatalf("%s != %s", err.Error(), tc.want)
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

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.res, m.err
}

// newTestBody creates an io.ReadCloser from a string for testing.
func newTestBody(s string) io.ReadCloser {
	return io.NopCloser(bytes.NewReader([]byte(s)))
}
