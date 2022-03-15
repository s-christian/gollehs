package evasion

import "bytes"

// ---

type NoMagicError string

func (e NoMagicError) Error() string {
	return string(e)
}

// ---

const (
	ErrNoMagic = NoMagicError("packet does not contain magic")
)

var (
	// These can be changed to whatever you want
	magicBytes = []byte{0x59, 0x41, 0x02}
	xorKey     = []byte{0x14, 0x23}
)

// ---

/*
	Encrypt or decrypt input with the evasion package's internal xorKey.
	Returns the xor-encrypted/decrypted byte slice.
*/
func XorEncryptDecryptBytes(input []byte) (output []byte) {
	for i, b := range input {
		output = append(output, b^xorKey[i%len(xorKey)])
	}

	return output
}

/*
	Check for magic bytes (evasion.magicBytes) in encryptedBytes. If present,
	return DATA.

	encryptedBytes format: "[...] magic DATA"
*/
func DecryptBytesAfterMagic(encryptedBytes []byte) ([]byte, error) {
	var decryptedBytes []byte

	if bytes.Contains(encryptedBytes, magicBytes) {
		decryptedBytes = XorEncryptDecryptBytes(bytes.Split(encryptedBytes, magicBytes)[1])
	} else {
		return nil, ErrNoMagic
	}

	return decryptedBytes, nil
}
