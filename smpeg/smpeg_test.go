package smpeg

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

var testfile = filepath.Join("..", "testdata", "0800001.mpg")

func TestNew(t *testing.T) {
	mpg, err := New(testfile, true)
	if err != nil {
		t.Fatal(err.Error())
	}

	mpg.Delete()
}

func TestNewDescr(t *testing.T) {
	file, err := os.Open(testfile)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer file.Close()

	mpg, err := NewDescr(int(file.Fd()), true)
	if err != nil {
		t.Fatal(err.Error())
	}

	mpg.Delete()
}

func TestNewData(t *testing.T) {
	data, err := ioutil.ReadFile(testfile)
	if err != nil {
		t.Fatal(err.Error())
	}

	mpg, err := NewData(data, true)
	if err != nil {
		t.Fatal(err.Error())
	}

	mpg.Delete()
}

func TestNewRWops(t *testing.T) {
	data, err := ioutil.ReadFile(testfile)
	if err != nil {
		t.Fatal(err.Error())
	}

	src := sdl.RWFromMem(unsafe.Pointer(&data[0]), len(data))

	mpg, err := NewRWops(src, true, true)
	if err != nil {
		t.Fatal(err.Error())
	}

	mpg.Delete()
}

func TestInfo(t *testing.T) {
	mpg, err := New(testfile, true)
	if err != nil {
		t.Fatal(err.Error())
	}

	info := mpg.Info()

	if info.Width != 352 || info.Height != 240 {
		t.Fatal()
	}

	if info.TotalSize != 1959952 {
		t.Fatal()
	}

	mpg.Delete()
}
