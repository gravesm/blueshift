package services

import (
	"github.com/dhowden/tag"
	"github.com/google/uuid"
	"io"
	"log"
	"os"
	"path/filepath"
)

func FileMetadata(f io.ReadSeeker) tag.Metadata {
	f.Seek(0, io.SeekStart)
	m, err := tag.ReadFrom(f)
	if err != nil {
		log.Fatal(err)
	}
	f.Seek(0, io.SeekStart)
	return m
}

type StreamHandler interface {
	Store(data io.Reader) string
	Get(path string) io.ReadCloser
}

type FileStreamHandler struct {
	Directory string
}

func (sh FileStreamHandler) Store(d io.Reader) string {
	p := sh.path()
	fp, err := os.Create(p)
	if err != nil {
		log.Fatalf("Could not create file %s: %v", p, err)
	}
	defer fp.Close()

	_, err = io.Copy(fp, d)
	if err != nil {
		log.Fatalf("Could not copy file %s: %v", p, err)
	}
	return p
}

func (sh FileStreamHandler) Get(path string) io.ReadCloser {
	fp, err := os.Open(path)
	if err != nil {
		log.Fatalf("Could not open file %s: %v", path, err)
	}
	return fp
}

func (sh FileStreamHandler) path() string {
	root, err := filepath.Abs(sh.Directory)
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(root, uuid.New().String())
}
