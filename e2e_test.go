package influxdb_test

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/influxdata/influxdb-client-go"
)

var e2e bool
var fuzz bool

func init() {
	flag.BoolVar(&e2e, "e2e", false, "run the end tests (requires a working influxdb instance on 127.0.0.1)")
	flag.BoolVar(&fuzz, "fuzz", false, "(Not Implemented) run the end tests (requires a working influxdb instance on 127.0.0.1)")
	flag.Parse()
}

// headerMap := map[[2]string][]*influxdb.Field{}
// standardHeaders := []string{"", "result", "table", "_start", "_stop", "_time", "_value", "_measurement"}

// fields := []*influxdb.Field{}

// for i := range rm {
// 	for j := range rm[i].Fields {
// 		key := rm[i].Fields[j].Key
// 		val := rm[i].Fields[j].Value
// 		k := [2]string{rm[i].Name(), key}
// 		l := 0
// 		q := headerMap[k]
// 		for ; l < len(q) && (q[l].Value != val || q[l].Key != rm[i].Fields[j].Key); l++ { //l := range headerMap[k][i] {
// 		}
// 		if l == len(q) {
// 			fields = append(q, rm[i].Fields[j])
// 		}
// 	}
// }
// _ = standardHeaders
// _ = fields
// for i := range rm {
// 	for j := range rm[i].Tags {
// 		if _, ok := headerMap[rm[i].Tags[j].Key]; !ok {
// 			headerMap[rm[i].Tags[j].Key] = struct{}{}
// 			headers = append(standardHeaders, rm[i].Tags[j].Key)
// 		}
// 	}
// }

// w := csv.NewWriter(writer)
// if err := w.Write(headers); err != nil {
// }
// return nil
//}

func TestE2E(t *testing.T) {
	if !e2e {
		t.Skipf("skipping end to end testing, spin up a copy of influxdb 2.x.x on 127.0.0.1 and run tests with --e2e to test")
	}
	influx, err := influxdb.New(nil, influxdb.WithAddress("http://127.0.0.1:9999"), influxdb.WithUserAndPass("e2e-test-user", "e2e-test-password"))
	if err != nil {
		t.Fatal(err)
	}
	// set up the bucket and org and get the token
	sRes, err := influx.Setup(context.Background(), "e2e-test-bucket", "e2e-test-org", 0)
	if err != nil {
		t.Fatal(err)
	}

	if sRes.User.Name != "e2e-test-user" {
		t.Fatalf("expected user to be %s, but was %s", "e2e-test-user", sRes.User.Name)
	}
	now := time.Now()
	err = influx.Write(context.Background(), "e2e-test-bucket", "e2e-test-org",
		&influxdb.RowMetric{
			NameStr: "test",
			Tags: []*influxdb.Tag{
				{Key: "ktest1", Value: "k-test1"},
				{Key: "ktest2,k-test3", Value: "k-test2, k-test3"}},
			Fields: []*influxdb.Field{
				{Key: "ftest1", Value: 3},
				{Key: "ftest2", Value: "kftest2"}},
			TS: now.Add(-time.Minute)},
		&influxdb.RowMetric{
			NameStr: "test",
			Tags: []*influxdb.Tag{
				{Key: "ktest1", Value: "k-test1"},
				{Key: "ktest2", Value: "k-test2"},
				{Key: "ktest3", Value: "k-test3"}},
			Fields: []*influxdb.Field{
				{Key: "ftest1", Value: 3},
				{Key: "ftest2", Value: "kftest2"}},
			TS: now.Add(-time.Minute)},
		&influxdb.RowMetric{
			NameStr: "test",
			Tags: []*influxdb.Tag{
				{Key: "ktest1", Value: "k-test1"},
				{Key: "ktest2", Value: "k-test2"}},
			Fields: []*influxdb.Field{
				{Key: "ftest1", Value: 3},
				{Key: "ftest2", Value: "kftest2"}},
			TS: now.Add(-time.Minute)},
		&influxdb.RowMetric{
			NameStr: "test",
			Tags: []*influxdb.Tag{
				{Key: "ktest1", Value: "k-test1"},
				{Key: "ktest3", Value: "k-test3"}},
			Fields: []*influxdb.Field{
				{Key: "ftest1", Value: 4},
				{Key: "ftest2", Value: "kftest2"}},
			TS: now.Add(-time.Minute)},
		&influxdb.RowMetric{
			NameStr: "test",
			Tags: []*influxdb.Tag{
				{Key: "ktest1", Value: "k-test1"},
				{Key: "ktest3", Value: "k-test3"}},
			Fields: []*influxdb.Field{
				{Key: "ftest1", Value: 3},
				{Key: "ftest2", Value: "kftest2"}},
			TS: now.Add(-time.Second * 30)},
		&influxdb.RowMetric{
			NameStr: "test",
			Tags: []*influxdb.Tag{
				{Key: "ktest1", Value: "k-test1"},
				{Key: "ktest3", Value: "k-test3"}},
			Fields: []*influxdb.Field{
				{Key: "ftest1", Value: 6},
				{Key: "ftest2", Value: "kftest3"}},
			TS: now.Add(-time.Second * 25)},
		&influxdb.RowMetric{
			NameStr: "tes0",
			Tags: []*influxdb.Tag{
				{Key: "ktest1", Value: "k-test1"},
				{Key: "ktest3", Value: "k-test3"}},
			Fields: []*influxdb.Field{
				{Key: "ftest1", Value: 5},
				{Key: "ftest2", Value: "kftest3"}},
			TS: now.Add(-time.Second * 25).Truncate(time.Microsecond).Add(1)},
		&influxdb.RowMetric{
			NameStr: "tes0",
			Tags: []*influxdb.Tag{
				{Key: "ktest1", Value: "k-test1"},
				{Key: "ktest3", Value: "k-test3"}},
			Fields: []*influxdb.Field{
				{Key: "ftest1", Value: 5},
				{Key: "ftest2", Value: "kftest3"}},
			TS: now.Add(-time.Second * 25).Truncate(time.Microsecond).Add(time.Microsecond)},
	)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(sRes.Auth.Token)
	time.Sleep(5 * time.Second)
	r, err := influx.QueryCSV(context.Background(), `from(bucket:"e2e-test-bucket")|>range(start:-1000h)|>group()`, `e2e-test-org`)
	if err != nil {
		t.Fatal(err)
		t.Fatal(err)
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("output:\n\n", string(b), "\n", int(b[0]))
}
