package middleware

import (
	"compress/gzip"
	"github.com/JuliyaMS/gophermart/internal/logger"
	"io"
	"net/http"
	"strings"
)

type CompressWriter struct {
	res http.ResponseWriter
	zw  *gzip.Writer
}

func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		res: w,
		zw:  gzip.NewWriter(w),
	}
}

type CompressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &CompressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *CompressWriter) Header() http.Header {
	return c.res.Header()
}

func (c *CompressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.res.Header().Set("Content-Encoding", "gzip")
	}
	c.res.WriteHeader(statusCode)
}

func (c *CompressWriter) Close() error {
	return c.zw.Close()
}

func (c CompressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func CompressionGzip(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.NewLogger()

		log.Infow("Start middleware")
		ow := w

		log.Infow("Check client's Accept-Encoding")
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if supportsGzip {
			log.Info("Start compress data")
			cw := NewCompressWriter(w)
			ow = cw

			defer func(cw *CompressWriter) {
				err := cw.Close()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Error("Error while write compress data", err.Error())
				}
			}(cw)
		}

		log.Infow("Check client's Content-Encoding")
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			log.Infow("Read compress data")
			cr, err := NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Error("Error while create compress reader:", err.Error())
				return
			}
			r.Body = cr

			defer func(cr *CompressReader) {
				er := cr.Close()
				if er != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Error("Error while read compress data", er.Error())
				}
			}(cr)
		}
		log.Infow("End middleware")
		h.ServeHTTP(ow, r)
	}
}
