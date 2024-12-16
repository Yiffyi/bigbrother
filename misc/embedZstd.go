package misc

import (
	"archive/tar"
	"bytes"
	"io"

	"github.com/DataDog/zstd"
)

type CompressedTarCloser struct {
	io.ReadCloser

	tarReader  io.Reader
	baseReader io.ReadCloser
}

func (c *CompressedTarCloser) Read(p []byte) (n int, err error) {
	return c.tarReader.Read(p)
}

func (c *CompressedTarCloser) Close() error {
	return c.baseReader.Close()
}

func NewReaderFromTarZstd(embededBytes []byte, fileName string) (io.ReadCloser, error) {
	bReader := bytes.NewReader(embededBytes)
	zstdReader := zstd.NewReader(bReader)
	tarReader := tar.NewReader(zstdReader)

	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			zstdReader.Close()
			return nil, err
		}

		// yes, this is how it works
		if hdr.Name == fileName || hdr.Name == "./"+fileName {
			return &CompressedTarCloser{tarReader: tarReader, baseReader: zstdReader}, nil
		}
	}

	zstdReader.Close()
	return nil, io.EOF
}
