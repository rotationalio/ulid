package ulid

import "encoding/binary"

type uint80 struct {
	Hi uint16
	Lo uint64
}

func (u *uint80) SetBytes(bs []byte) {
	u.Hi = binary.BigEndian.Uint16(bs[:2])
	u.Lo = binary.BigEndian.Uint64(bs[2:])
}

func (u *uint80) AppendTo(bs []byte) {
	binary.BigEndian.PutUint16(bs[:2], u.Hi)
	binary.BigEndian.PutUint64(bs[2:], u.Lo)
}

func (u *uint80) Add(n uint64) (overflow bool) {
	lo, hi := u.Lo, u.Hi
	if u.Lo += n; u.Lo < lo {
		u.Hi++
	}
	return u.Hi < hi
}

func (u uint80) IsZero() bool {
	return u.Hi == 0 && u.Lo == 0
}
