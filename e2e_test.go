package influxdb_test

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	influxdb "github.com/influxdata/influxdb-client-go"
)

var e2e bool
var fuzz bool

func init() {
	flag.BoolVar(&e2e, "e2e", false, "run the end tests (requires a working influxdb instance on 127.0.0.1)")
	flag.BoolVar(&fuzz, "fuzz", false, "(Not Implemented) run the end tests (requires a working influxdb instance on 127.0.0.1)")
	flag.Parse()
}

func TestE2E(t *testing.T) {
	if !e2e {
		t.Skipf("skipping end to end testing, spin up a copy of influxdb 2.x.x on 127.0.0.1 and run tests with --e2e")
	}
	influx, err := influxdb.New("", "", influxdb.WithUserAndPass("e2e-test-user", "e2e-test-password"))
	if err != nil {
		t.Fatal(err)
	}
	// set up the bucket and org and get the token
	sRes, err := influx.Setup(context.Background(), "e2e-test-bucket", "e2e-test-org", 0)
	if err != nil {
		t.Fatal(err)
	}

	err = influx.Ping(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if sRes.User.Name != "e2e-test-user" {
		t.Fatalf("expected user to be %s, but was %s", "e2e-test-user", sRes.User.Name)
	}
	now := time.Now()
	_, err = influx.Write(context.Background(), "e2e-test-bucket", "e2e-test-org",
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

	time.Sleep(5 * time.Second)

	r, err := influx.QueryCSV(
		context.Background(),
		`from(bucket:bucket)|>range(start:-10000h)|>group(columns:["_field"])`,
		`e2e-test-org`,
		struct {
			Bucket string `flux:"bucket"`
		}{Bucket: "e2e-test-bucket"})
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(b))
}
