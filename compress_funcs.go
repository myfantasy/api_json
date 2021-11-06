package api_json

import "github.com/myfantasy/compress"

var zipOutPreferCompress = []compress.CompressionType{compress.Zip, compress.NoCompression}

var ZipCompressFunc GetCompressionFunc = func(preferCompress []compress.CompressionType,
) (outCompType compress.CompressionType, outPreferCompress []compress.CompressionType) {
	if len(preferCompress) != 0 {
		ze := false
		for _, ct := range preferCompress {
			if ct == compress.Zip || ct == compress.Zip1 || ct == compress.Zip9 {
				ze = true
			}
		}
		if !ze {
			outCompType = compress.NoCompression
			outPreferCompress = zipOutPreferCompress
		}
	}
	outCompType = compress.Zip
	outPreferCompress = zipOutPreferCompress
	return outCompType, outPreferCompress
}
