package ir

import (
	"bytes"
	"fmt"
)

// ModuleBuilder is used to build an LLVM module and generate its LLVM bitcode
// representation.
type ModuleBuilder struct {
	// The byte data making up the module.
	buff bytes.Buffer

	// The bit vector used for the byte currently being built: the LLVM bitcode
	// format often uses bit counts which aren't a multiple of 8 (ie. occupy
	// less than a single) byte.  So, we need to introduce a bit vector used to
	// build bytes.
	bitVecData byte
	bitVecLen  byte
}

// NewModuleBuilder creates a new module builder for a module named name.
func NewModuleBuilder(name string) (mb *ModuleBuilder) {
	mb = &ModuleBuilder{}

	mb.writeHeader()

	return
}

// CompileBitcode returns the bitcode representation of the built module.
func (mb *ModuleBuilder) CompileBitcode() []byte {
	return mb.buff.Bytes()
}

/* -------------------------------------------------------------------------- */

// writeHeader writes the LLVM bitcode header to the buffer.
func (mb *ModuleBuilder) writeHeader() {
	// Write the magic number to the header file.

}

/* -------------------------------------------------------------------------- */

// writeFixedInt writes a fixed width integer to the byte buffer.
func (mb *ModuleBuilder) writeFixedInt(n uint64, bitWidth byte) {
	var mask uint64 = 0xff

	var i byte = 0
	for ; i < bitWidth; i += 8 {
		bits := byte((n & mask) >> i)

		if bitWidth-i >= 8 {
			mb.writeBits(bits, 8)
		} else {
			mb.writeBits(bits, bitWidth-i)
		}

		mask <<= 8
	}
}

// writeVarInt writes a variable width integer to the byte buffer.
func (mb *ModuleBuilder) writeVarInt(n uint64, intBitWidth, fieldBitWidth byte) {
	//var mask uint64 =
}

// write6Char writes a 6 bit character to a fixed 6-bit field.
func (mb *ModuleBuilder) write6Char(b byte) {
	if b == '_' {
		mb.writeBits(63, 6)
	} else if b == '.' {
		mb.writeBits(62, 6)
	} else if 'a' <= b && b <= 'z' {
		mb.writeBits(b-'a', 6)
	} else if 'A' <= b && b <= 'A' {
		mb.writeBits(b-'A'+26, 6)
	} else if '0' <= b && b <= '9' {
		mb.writeBits(b-'0'+52, 6)
	} else {
		panic(fmt.Sprintf("error: cannot encode %c as a 6-bit character", b))
	}
}

// writeBits writes n <= 8 bits to stored in b to the stream.
func (mb *ModuleBuilder) writeBits(b byte, n byte) {
	for n > 0 {
		toWrite := n - mb.bitVecLen
		writeMask := ^(byte(0xff) << toWrite)
		mb.bitVecData |= (b & writeMask) << mb.bitVecLen
		mb.bitVecLen += toWrite

		if mb.bitVecLen == 8 {
			mb.buff.WriteByte(mb.bitVecData)
			mb.bitVecData = 0
			mb.bitVecLen = 0
		}

		n -= toWrite
		b >>= toWrite
	}
}
