// Package aesxts provides a Go interface to the AES-XTS encryption/decryption functions in the C library.
// The Go version of XTS is not used because Nitendo implemented a custom tweak not compatible with the standard.
// The C library is a modified version of mbedTLS by https://github.com/blawar/nut
package aesxts

// #cgo CFLAGS: -g -I mbedtls/include -D_BSD_SOURCE -D_POSIX_SOURCE -D_POSIX_C_SOURCE=200112L -D_DEFAULT_SOURCE -D__USE_MINGW_ANSI_STDIO=1 -D_FILE_OFFSET_BITS=64
// #cgo LDFLAGS: -L mbedtls/library -lmbedtls -lmbedx509 -lmbedcrypto
// #include <stdlib.h>
// #include "aes.h"
import "C"
import "unsafe"

func Decrypt(key []byte, data []byte, size, sectorSize int) ([]byte, error) {
	// Store the key as a C array
	ptrKey := unsafe.Pointer(&key[0])

	ptrData := unsafe.Pointer(&data[0])

	dec_data := make([]byte, len(data))
	ptrDecData := unsafe.Pointer(&dec_data[0])

	aes_ctx := C.new_aes_ctx(ptrKey, 32, 52) // AES_MODE_XTS
	C.aes_xts_decrypt(aes_ctx, ptrDecData, ptrData, C.ulong(size), 0, C.ulong(sectorSize))

	return dec_data, nil
}
