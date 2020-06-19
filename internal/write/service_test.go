// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package write

import (
	"github.com/influxdata/influxdb-client-go/api/write"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAddDefaultTags(t *testing.T) {
	opts := write.DefaultOptions()
	assert.Len(t, opts.DefaultTags(), 0)

	opts.AddDefaultTag("dt1", "val1")
	opts.AddDefaultTag("zdt", "val2")
	srv := NewService("org", "buc", nil, opts)

	p := write.NewPointWithMeasurement("test")
	p.AddTag("id", "101")

	p.AddField("float32", float32(80.0))

	s, err := srv.EncodePoints(p)
	require.Nil(t, err)
	assert.Equal(t, "test,dt1=val1,id=101,zdt=val2 float32=80\n", s)
	assert.Len(t, p.TagList(), 1)

	p = write.NewPointWithMeasurement("x")
	p.AddTag("xt", "1")
	p.AddField("i", 1)

	s, err = srv.EncodePoints(p)
	require.Nil(t, err)
	assert.Equal(t, "x,dt1=val1,xt=1,zdt=val2 i=1i\n", s)
	assert.Len(t, p.TagList(), 1)

	p = write.NewPointWithMeasurement("d")
	p.AddTag("id", "1")
	// do not overwrite point tag
	p.AddTag("zdt", "val10")
	p.AddField("i", -1)

	s, err = srv.EncodePoints(p)
	require.Nil(t, err)
	assert.Equal(t, "d,dt1=val1,id=1,zdt=val10 i=-1i\n", s)

	assert.Len(t, p.TagList(), 2)
}
