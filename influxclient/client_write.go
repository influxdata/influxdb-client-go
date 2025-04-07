package influxclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/influxdata/influxdb-client-go/v3/influxclient/gzip"
)

// WritePoints writes all the given points to the server into the given bucket.
// The points are written synchronously. For a higher throughput
// API that buffers individual points and writes them asynchronously,
// use the PointsWriter method.
func (c *Client) WritePoints(ctx context.Context, bucket string, points ...*Point) error {
	var buff []byte
	for _, p := range points {
		bts, err := p.MarshalBinary(c.params.WriteParams.Precision, c.params.WriteParams.DefaultTags)
		if err != nil {
			return err
		}
		buff = append(buff, bts...)
	}
	return c.Write(ctx, bucket, buff)
}

// Write writes line protocol record(s) to the server into the given bucket.
// Multiple records must be separated by the new line character (\n)
// Data are written synchronously. For a higher throughput
// API that buffers individual points and writes them asynchronously,
// use the PointsWriter method.
func (c *Client) Write(ctx context.Context, bucket string, buff []byte) error {
	var body io.Reader
	var err error
	u, _ := c.apiURL.Parse("write")
	params := u.Query()
	params.Set("org", c.params.Organization)
	params.Set("bucket", bucket)
	params.Set("precision", c.params.WriteParams.Precision.String())
	if c.params.WriteParams.Consistency != "" {
		params.Set("consistency", string(c.params.WriteParams.Consistency))
	}
	u.RawQuery = params.Encode()
	body = bytes.NewReader(buff)
	headers := http.Header{"Content-Type": {"application/json"}}
	if c.params.WriteParams.GzipThreshold > 0 && len(buff) >= c.params.WriteParams.GzipThreshold {
		body, err = gzip.CompressWithGzip(body)
		if err != nil {
			return fmt.Errorf("unable to compress write body: %w", err)
		}
		headers["Content-Encoding"] = []string{"gzip"}
	}
	_, err = c.makeAPICall(ctx, httpParams{
		endpointURL: u,
		httpMethod:  "POST",
		headers:     headers,
		queryParams: u.Query(),
		body:        body,
	})
	return err
}

// WriteData encodes fields of custom points into line protocol
// and writes line protocol record(s) to the server into the given bucket.
// Each custom point must be annotated with 'lp' prefix and values measurement,tag, field or timestamp.
// Valid point must contain measurement and at least one field.
//
// A field with timestamp must be of a type time.Time
//
//	 type TemperatureSensor struct {
//		  Measurement string `lp:"measurement"`
//		  Sensor string `lp:"tag,sensor"`
//		  ID string `lp:"tag,device_id"`
//		  Temp float64 `lp:"field,temperature"`
//		  Hum int	`lp:"field,humidity"`
//		  Time time.Time `lp:"timestamp"`
//		  Description string `lp:"-"`
//	 }
//
// The points are written synchronously. For a higher throughput
// API that buffers individual points and writes them asynchronously,
// use the PointsWriter method.
func (c *Client) WriteData(ctx context.Context, bucket string, points ...interface{}) error {
	var buff []byte
	for _, p := range points {
		byts, err := encode(p, c.params.WriteParams)
		if err != nil {
			return fmt.Errorf("error encoding point: %w", err)
		}
		buff = append(buff, byts...)
	}

	return c.Write(ctx, bucket, buff)
}

func encode(x interface{}, params WriteParams) ([]byte, error) {
	if err := checkContainerType(x, false, "point"); err != nil {
		return nil, err
	}
	t := reflect.TypeOf(x)
	v := reflect.ValueOf(x)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	fields := reflect.VisibleFields(t)

	var point Point

	for _, f := range fields {
		name := f.Name
		if tag, ok := f.Tag.Lookup("lp"); ok {
			if tag == "-" {
				continue
			}
			parts := strings.Split(tag, ",")
			if len(parts) > 2 {
				return nil, fmt.Errorf("multiple tag attributes are not supported")
			}
			typ := parts[0]
			if len(parts) == 2 {
				name = parts[1]
			}
			switch typ {
			case "measurement":
				if point.Measurement != "" {
					return nil, fmt.Errorf("multiple measurement fields")
				}
				point.Measurement = v.FieldByIndex(f.Index).String()
			case "tag":
				point.AddTag(name, v.FieldByIndex(f.Index).String())
			case "field":
				point.AddField(name, v.FieldByIndex(f.Index).Interface())
			case "timestamp":
				if f.Type != timeType {
					return nil, fmt.Errorf("cannot use field '%s' as a timestamp", f.Name)
				}
				point.Timestamp = v.FieldByIndex(f.Index).Interface().(time.Time)
			default:
				return nil, fmt.Errorf("invalid tag %s", typ)
			}
		}
	}
	if point.Measurement == "" {
		return nil, fmt.Errorf("no struct field with tag 'measurement'")
	}
	if len(point.Fields) == 0 {
		return nil, fmt.Errorf("no struct field with tag 'field'")
	}
	return point.MarshalBinary(params.Precision, params.DefaultTags)
}

// PointsWriter returns a PointsWriter value that support fast asynchronous
// writing of points to Influx. All the points are written into the given bucket.
//
// The returned PointsWriter must be closed after use to release resources
// and flush any buffered points.
func (c *Client) PointsWriter(bucket string) *PointsWriter {
	return NewPointsWriter(c.Write, bucket, c.params.WriteParams)
}
