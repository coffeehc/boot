package transport

import (
	"compress/gzip"
	"io"

	"google.golang.org/grpc/encoding"
)

func init() {
	encoding.RegisterCompressor(&gzipEncoding{})
}

type gzipEncoding struct {
}

func (impl *gzipEncoding) Compress(w io.Writer) (io.WriteCloser, error) {
	return gzip.NewWriter(w), nil
}

func (impl *gzipEncoding) Decompress(r io.Reader) (io.Reader, error) {
	return gzip.NewReader(r)
}

func (impl *gzipEncoding) Name() string {
	return "gzip"
}
