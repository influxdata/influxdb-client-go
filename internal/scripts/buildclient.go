package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

const swaggerurl = "https://raw.githubusercontent.com/influxdata/influxdb/master/http/swagger.yml"

func main() {
	f, err := ioutil.TempFile("", "swagger*.yml")
	if err != nil {
		log.Fatal("failed to create temporary file")
	}
	defer func() {
		name := f.Name()
		_ = f.Close()
		_ = os.Remove(name)
	}()
	resp, err := http.DefaultClient.Get(swaggerurl)
	if err != nil {
		log.Fatal("failed to download swagger")
	}
	defer resp.Body.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		log.Fatalf("failed to download swagger %v", err)
	}
	_ = f.Close()
	if err != nil {
		log.Fatalf("failed to flush swagger to temp file %v", err)
	}

	cmd := exec.Command("oapi-codegen", "-generate=client", "-package=genclient", f.Name())
	r, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("oapi-codegen execution failed %v", err)
	}
	fout, err := os.Create("generated/generatedclient.go")
	if err != nil {
		log.Fatalf("failed to generate client code file %v", err)
	}
	cmd.Stderr = os.Stderr
	_, err = fout.WriteString(`// Code generated for client DO NOT EDIT.
// TODO(docmerlin): modify generator so we don't need to edit the generated code.
// The generated code is modified to remove external dependencies and the GetSwagger function is removed.`)
	if err != nil {
		log.Fatalf("unable to write generated comment %v", err)
	}
	go func() {
		out, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatalf("failed to read client code from generator %v", err)
		}
		_, err = fout.Write(out)
		if err != nil {
			log.Fatalf("failed to write to client code file %v", err)
		}
		err = fout.Close()
		if err != nil {
			log.Fatalf("failed to flush client code file %v", err)
		}
	}()

	err = cmd.Run()
	if err != nil {
		log.Fatalf("failed when running oapi-codegen %v", err)
	}
}
