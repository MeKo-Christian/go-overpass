# go-overpass Examples

This directory contains example programs demonstrating how to use the go-overpass library.

## Running Examples

Each example is in its own directory and can be run using:

```bash
cd example/<example-name>
go run main.go
```

## Examples

### 1. Basic Query ([basic/](basic/))

The simplest example showing how to:

- Create a default client
- Query for a specific node
- Access node properties and tags

```bash
cd example/basic
go run main.go
```

### 2. Custom Client ([custom_client/](custom_client/))

Demonstrates advanced client configuration:

- Custom HTTP client with timeout
- Custom Overpass API endpoint
- Parallel request limiting
- Querying for multiple elements

```bash
cd example/custom_client
go run main.go
```

### 3. Ways and Relations ([ways_and_relations/](ways_and_relations/))

Shows how to work with complex OSM elements:

- Querying ways (streets, paths, etc.)
- Accessing way nodes and geometry
- Working with relations
- Understanding member relationships

```bash
cd example/ways_and_relations
go run main.go
```

### 4. Area Query ([area_query/](area_query/))

Demonstrates bounding box queries:

- Using bbox parameter in queries
- Querying multiple element types
- Handling query bounds
- Working with center coordinates for ways

```bash
cd example/area_query
go run main.go
```

## Query Format

All queries must include `[out:json]` to get JSON responses:

```
[out:json];
node(52.5,13.4);
out;
```

## Common Query Patterns

### Query by ID

```
[out:json];
node(123456);
out;
```

### Query by Bounding Box

```
[out:json];
node(south,west,north,east);
out;
```

### Query by Tag

```
[out:json];
node["amenity"="restaurant"];
out;
```

### Query with Geometry

```
[out:json];
way(123456);
out geom;
```

## Learn More

- [Overpass QL Documentation](https://wiki.openstreetmap.org/wiki/Overpass_API/Language_Guide)
- [Overpass Turbo](https://overpass-turbo.eu/) - Interactive query builder
- [go-overpass Documentation](https://pkg.go.dev/github.com/MeKo-Christian/go-overpass)
