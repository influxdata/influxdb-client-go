package test

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

// GenPoints generates num points
func GenPoints(num int) []*write.Point {
	points := make([]*write.Point, num)
	rand.Seed(321)

	t := time.Now()
	for i := 0; i < len(points); i++ {
		points[i] = write.NewPoint(
			"test",
			map[string]string{
				"id":       fmt.Sprintf("rack_%v", i%10),
				"vendor":   "AWS",
				"hostname": fmt.Sprintf("host_%v", i%100),
			},
			map[string]interface{}{
				"temperature": rand.Float64() * 80.0,
				"disk_free":   rand.Float64() * 1000.0,
				"disk_total":  (i/10 + 1) * 1000000,
				"mem_total":   (i/100 + 1) * 10000000,
				"mem_free":    rand.Uint64(),
			},
			t)
		if i%10 == 0 {
			t = t.Add(time.Second)
		}
	}
	return points
}

// GenRecords generates num points
func GenRecords(num int) []string {
	lines := make([]string, num)
	rand.Seed(321)

	t := time.Now()
	for i := 0; i < len(lines); i++ {
		lines[i] = fmt.Sprintf("test,id=rack_%v,vendor=AWS,hostname=host_%v temperature=%v,disk_free=%v,disk_total=%vi,mem_total=%vi,mem_free=%vu %v",
			i%10, i%100, rand.Float64()*80.0, rand.Float64()*1000.0, (i/10+1)*1000000, (i/100+1)*10000000, rand.Uint64(), t.UnixNano())
		if i%10 == 0 {
			t = t.Add(time.Nanosecond)
		}
	}
	return lines
}
