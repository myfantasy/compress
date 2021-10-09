package compress

import (
	"context"

	"github.com/myfantasy/mfs"
	"github.com/myfantasy/mft"
)

// Compression or encriprion type
type CompressionType string

// Algs
const (
	NoCompression CompressionType = ""
	Zip           CompressionType = "gzip"
	Zip1          CompressionType = "gzip1"
	Zip9          CompressionType = "gzip9"

	// Aes - AES 256 alg
	Aes CompressionType = "aes"
)

type CompressFunc func(ctx context.Context, algorithm CompressionType, body []byte, encryptKey []byte) (algorithmUsed CompressionType, result []byte, err *mft.Error)
type RestoreFunc func(ctx context.Context, algorithm CompressionType, body []byte, decryptKey []byte) (algorithmUsed CompressionType, result []byte, err *mft.Error)

// Generator - compressor
type Generator struct {
	mx          mfs.PMutex
	compressors map[CompressionType]CompressFunc
	restores    map[CompressionType]RestoreFunc
}

func (g *Generator) Add(name CompressionType, compressor CompressFunc, restore RestoreFunc) {
	g.mx.Lock()
	defer g.mx.Unlock()

	g.compressors[name] = compressor
	g.restores[name] = restore
}

func (g *Generator) Init() {
	g.mx.Lock()
	defer g.mx.Unlock()
	g.compressors = make(map[CompressionType]CompressFunc)
	g.restores = make(map[CompressionType]RestoreFunc)
}

// GeneratorCreate - generate with default algs
func GeneratorCreate(gzipDefaultLevel int) *Generator {
	g := &Generator{}
	g.Init()

	g.Add(Zip, GZipCompressGenerator(gzipDefaultLevel), GZipRestore)
	g.Add(Zip1, GZipCompressGenerator(1), GZipRestore)
	g.Add(Zip9, GZipCompressGenerator(9), GZipRestore)
	g.Add(Aes, AesEncrypt, AesDecrypt)

	return g
}

func (g *Generator) Compress(ctx context.Context, must bool, algorithm CompressionType, body []byte, encryptKey []byte) (algorithmUsed CompressionType, result []byte, err *mft.Error) {
	g.mx.RLock()
	defer g.mx.RUnlock()

	if algorithm == "" {
		return "", body, nil
	}

	c, ok := g.compressors[algorithm]
	if !ok || c == nil {
		if len(encryptKey) == 0 && !must {
			return "", body, nil
		}

		return "", nil, mft.GenerateError(10200000, algorithm)
	}

	return c(ctx, algorithm, body, encryptKey)
}
func (g *Generator) Restore(ctx context.Context, algorithm CompressionType, body []byte, decryptKey []byte) (algorithmUsed CompressionType, result []byte, err *mft.Error) {
	g.mx.RLock()
	defer g.mx.RUnlock()

	if algorithm == "" {
		return "", body, nil
	}

	r, ok := g.restores[algorithm]
	if !ok || r == nil {
		return "", nil, mft.GenerateError(10200001, algorithm)
	}

	return r(ctx, algorithm, body, decryptKey)
}
