package aesxts

// #cgo CFLAGS: -g -Wall
// #cgo LDFLAGS: -lmbedtls -lmbedx509 -lmbedcrypto
// #include "mbedtls/cipher.h"
// #include "mbedtls/cmac.h"
import "C"
import (
	"fmt"
	"unsafe"
)

func NewAESCtx(key []byte) (*C.mbedtls_cipher_context_t, error) {
	var aes_ctx C.mbedtls_cipher_context_t
	C.mbedtls_cipher_init(&aes_ctx)

	if C.mbedtls_cipher_setup(&aes_ctx, C.mbedtls_cipher_info_from_type(C.MBEDTLS_CIPHER_AES_128_XTS)) != 0 {
		return nil, fmt.Errorf("failed to set up AES context")
	}

	ptrKey := unsafe.Pointer(&key[0])

	if C.mbedtls_cipher_setkey(&aes_ctx, (*C.uchar)(ptrKey), 32*8, C.MBEDTLS_DECRYPT) != 0 {
		return nil, fmt.Errorf("failed to set key for AES context")
	}

	return &aes_ctx, nil
}

func FreeAESCtx(aes_ctx *C.mbedtls_cipher_context_t) {
	C.mbedtls_cipher_free(aes_ctx)
}

func AESDecrypt(aes_ctx *C.mbedtls_cipher_context_t, dst, src []byte, size int) error {
	C.mbedtls_cipher_reset(aes_ctx)

	if C.mbedtls_cipher_get_cipher_mode(aes_ctx) != C.MBEDTLS_MODE_XTS {
		return fmt.Errorf("AES context is not in XTS mode")
	}

	C.mbedtls_cipher_update(aes_ctx, (*C.uchar)(&src[0]), C.size_t(size), (*C.uchar)(&dst[0]), (*C.size_t)(unsafe.Pointer(&size)))

	return nil
}

func AESXTSDecrypt(aes_ctx *C.mbedtls_cipher_context_t, dst, src []byte, size, sector, sectorSize int) error {
	if size%sectorSize != 0 {
		return fmt.Errorf("size must be a multiple of sectorSize")
	}

	tweak := make([]byte, 16)
	for i := 0; i < size; i += sectorSize {
		getTweak(tweak, sector)
		setIV(aes_ctx, (*C.uchar)(&tweak[0]), 16)
		AESDecrypt(aes_ctx, dst[i:], src[i:], sectorSize)
		sector++
	}

	return nil
}

func getTweak(tweak []byte, sector int) {
	for i := 0xF; i >= 0; i-- {
		tweak[i] = byte(sector & 0xFF)
		sector >>= 8
	}
}

func setIV(aes_ctx *C.mbedtls_cipher_context_t, iv *C.uchar, ivLen int) error {
	if C.mbedtls_cipher_set_iv(aes_ctx, iv, C.size_t(ivLen)) != 0 {
		return fmt.Errorf("failed to set IV for AES context")
	}
	return nil
}
