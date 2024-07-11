package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	input   = flag.String("i", "", "input package")
	output  = flag.String("o", "", "output directory")
	exclude []string
)

func init() {
	flag.Func("e", "name patterns for files to be excluded. will be 1st arg of path.Match. comma-separated string", func(s string) error {
		exclude = strings.Split(s, ",")
		for i, ex := range exclude {
			_, err := path.Match(ex, "")
			if err != nil {
				return fmt.Errorf("malformed pattern at %d. not compliant to path.Math. err = %v", i, err)
			}
		}
		return nil
	})
}

func main() {
	flag.Parse()

	fsys := os.DirFS(*input)

	if _, err := os.Stat(*output); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			panic(err)
		}
		err := os.MkdirAll(*output, fs.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if p == "." {
			return err
		}
		if d.IsDir() || err != nil {
			return err
		}

		for _, pat := range exclude {
			if matched, _ := path.Match(pat, p); matched {
				return nil
			}
		}

		dstName := filepath.Join(*output, filepath.FromSlash(p))

		err = os.MkdirAll(filepath.Dir(dstName), fs.ModePerm)
		if err != nil {
			return err
		}

		dst, err := os.Create(dstName)
		if err != nil {
			return err
		}
		defer func() { _ = dst.Close() }() // generally ignores close error.

		src, err := os.Open(filepath.Join(*input, filepath.FromSlash(p)))
		if err != nil {
			return err
		}
		defer func() { _ = src.Close() }()

		_, err = io.Copy(dst, src)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		panic(err)
	}
}
