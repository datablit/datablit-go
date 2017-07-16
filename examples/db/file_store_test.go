package db

import (
	"io/ioutil"
	"testing"

	"os"

	"github.com/c2stack/c2g/c2"
	"github.com/c2stack/c2g/node"
)

func Test_FileStore(t *testing.T) {
	fs := FileStore{VarDir: "./var"}
	b, _ := node.BirdBrowser("../../node", `{"bird":[{
		"name" : "robin",
		"wingspan" : 10
	}]}`)
	err := fs.DbWrite("x", "m", b)
	if err != nil {
		t.Error(err)
	}
	f, err := os.Open("./var/x:m.json")
	if err != nil {
		t.Error(err)
	}
	actual, err := ioutil.ReadAll(f)
	if err != nil {
		t.Error(err)
	}
	expected := `{
"bird":[
  {
    "name":"robin",
    "wingspan":10}]}`
	if err := c2.CheckEqual(expected, string(actual)); err != nil {
		t.Error(err)
	}
}
