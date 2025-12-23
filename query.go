package overpass

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type overpassResponse struct {
	OSM3S struct {
		TimestampOSMBase time.Time `json:"timestamp_osm_base"`
	} `json:"osm3s"`
	Elements []overpassResponseElement `json:"elements"`
}

type overpassResponseElement struct {
	Type      ElementType `json:"type"`
	ID        int64       `json:"id"`
	Lat       float64     `json:"lat"`
	Lon       float64     `json:"lon"`
	Timestamp *time.Time  `json:"timestamp"`
	Version   int64       `json:"version"`
	Changeset int64       `json:"changeset"`
	User      string      `json:"user"`
	UID       int64       `json:"uid"`
	Nodes     []int64     `json:"nodes"`
	Members   []struct {
		Type     ElementType `json:"type"`
		Ref      int64       `json:"ref"`
		Role     string      `json:"role"`
		Geometry []struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"geometry,omitempty"`
	} `json:"members"`
	Geometry []struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"geometry"`
	Bounds *struct {
		MinLat float64 `json:"minlat"`
		MinLon float64 `json:"minlon"`
		MaxLat float64 `json:"maxlat"`
		MaxLon float64 `json:"maxlon"`
	} `json:"bounds"`
	Tags map[string]string `json:"tags"`
}

// httpPost sends HTTP POST request with context support.
func (c *Client) httpPost(ctx context.Context, query string) ([]byte, error) {
	<-c.semaphore

	defer func() { c.semaphore <- struct{}{} }()

	// Create POST request with context
	data := url.Values{"data": []string{query}}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiEndpoint,
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use Do instead of PostForm to support context
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil && err == nil {
			err = fmt.Errorf("http error: %w", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("overpass engine error: %w", &ServerError{resp.StatusCode, body})
	}

	return body, nil
}

func unmarshal(body []byte) (Result, error) {
	var overpassRes overpassResponse
	err := json.Unmarshal(body, &overpassRes)
	if err != nil {
		return Result{}, fmt.Errorf("overpass engine error: %w", err)
	}

	result := Result{
		Timestamp: overpassRes.OSM3S.TimestampOSMBase,
		Count:     len(overpassRes.Elements),
		Nodes:     make(map[int64]*Node),
		Ways:      make(map[int64]*Way),
		Relations: make(map[int64]*Relation),
	}

	for _, el := range overpassRes.Elements {
		meta := Meta{
			ID:        el.ID,
			Timestamp: el.Timestamp,
			Version:   el.Version,
			Changeset: el.Changeset,
			User:      el.User,
			UID:       el.UID,
			Tags:      el.Tags,
		}
		switch el.Type {
		case ElementTypeNode:
			node := result.getNode(el.ID)
			*node = Node{
				Meta: meta,
				Lat:  el.Lat,
				Lon:  el.Lon,
			}
		case ElementTypeWay:
			way := result.getWay(el.ID)

			*way = Way{
				Meta:     meta,
				Nodes:    make([]*Node, len(el.Nodes)),
				Geometry: make([]Point, len(el.Geometry)),
			}
			for idx, nodeID := range el.Nodes {
				way.Nodes[idx] = result.getNode(nodeID)
			}

			if el.Bounds != nil {
				way.Bounds = &Box{
					Min: Point{
						Lat: el.Bounds.MinLat,
						Lon: el.Bounds.MinLon,
					},
					Max: Point{
						Lat: el.Bounds.MaxLat,
						Lon: el.Bounds.MaxLon,
					},
				}
			}

			for idx, geo := range el.Geometry {
				way.Geometry[idx].Lat = geo.Lat
				way.Geometry[idx].Lon = geo.Lon
			}
		case ElementTypeRelation:
			relation := result.getRelation(el.ID)

			*relation = Relation{
				Meta:    meta,
				Members: make([]RelationMember, len(el.Members)),
			}
			for idx, member := range el.Members {
				relationMember := RelationMember{
					Type: member.Type,
					Role: member.Role,
				}
				switch member.Type {
				case ElementTypeNode:
					relationMember.Node = result.getNode(member.Ref)
				case ElementTypeWay:
					// Get or create the way from the result
					way := result.getWay(member.Ref)
					relationMember.Way = way
					// If inline geometry is provided (from "out geom"), populate the way's geometry
					// This is needed for multipolygon relations where member ways may not be
					// returned as separate elements but have their geometry embedded in the relation
					if len(member.Geometry) > 0 && len(way.Geometry) == 0 {
						way.Geometry = make([]Point, len(member.Geometry))
						for i, g := range member.Geometry {
							way.Geometry[i] = Point{Lat: g.Lat, Lon: g.Lon}
						}
					}
				case ElementTypeRelation:
					relationMember.Relation = result.getRelation(member.Ref)
				}

				relation.Members[idx] = relationMember
			}

			if el.Bounds != nil {
				relation.Bounds = &Box{
					Min: Point{
						Lat: el.Bounds.MinLat,
						Lon: el.Bounds.MinLon,
					},
					Max: Point{
						Lat: el.Bounds.MaxLat,
						Lon: el.Bounds.MaxLon,
					},
				}
			}
		}
	}

	return result, nil
}

// QueryContext runs query with context using default client.
func QueryContext(ctx context.Context, query string) (Result, error) {
	return DefaultClient.QueryContext(ctx, query)
}

// Query is deprecated: use QueryContext instead.
// It runs query with default client using context.Background().
func Query(query string) (Result, error) {
	return QueryContext(context.Background(), query)
}

type ServerError struct {
	StatusCode int
	Body       []byte
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("%d %s", e.StatusCode, http.StatusText(e.StatusCode))
}
