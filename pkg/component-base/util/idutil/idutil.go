package idutil

import (
	"crypto/rand"
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/sony/sonyflake"
	hashids "github.com/speps/go-hashids"

	"business-dev-bone/pkg/component-base/util/iputil"
	"business-dev-bone/pkg/component-base/util/stringutil"
)

// Defiens alphabet.
const (
	Alphabet16 = "abcdef1234567890"
	Alphabet36 = "abcdefghijklmnopqrstuvwxyz1234567890"
	Alphabet62 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
)

var sf *sonyflake.Sonyflake

func init() {
	var st sonyflake.Settings
	st.MachineID = func() (uint16, error) {
		ip := iputil.GetLocalIP()

		return uint16([]byte(ip)[2])<<8 + uint16([]byte(ip)[3]), nil
	}

	sf = sonyflake.NewSonyflake(st)
}

// GetIntID returns uint64 uniq id.
func GetIntID() uint64 {
	id, err := sf.NextID()
	if err != nil {
		panic(err)
	}

	return id
}

// GetInstanceID returns id format like: secret-2v69o5
func GetInstanceID(uid uint64, prefix string) string {
	hd := hashids.NewData()
	hd.Alphabet = Alphabet36
	hd.MinLength = 6
	hd.Salt = "x20k5x"

	h, err := hashids.NewWithData(hd)
	if err != nil {
		panic(err)
	}

	i, err := h.Encode([]int{int(uid)})
	if err != nil {
		panic(err)
	}

	return prefix + stringutil.Reverse(i)
}

// GetUUID36 returns id format like: 300m50zn91nwz5.
func GetUUID36(prefix string) string {
	id := GetIntID()
	hd := hashids.NewData()
	hd.Alphabet = Alphabet36

	h, err := hashids.NewWithData(hd)
	if err != nil {
		panic(err)
	}

	i, err := h.Encode([]int{int(id)})
	if err != nil {
		panic(err)
	}

	return prefix + stringutil.Reverse(i)
}

func randString(letters string, n int) string {
	output := make([]byte, n)

	// We will take n bytes, one byte for each character of output.
	randomness := make([]byte, n)

	// read all random
	_, err := rand.Read(randomness)
	if err != nil {
		panic(err)
	}

	l := len(letters)
	// fill output
	for pos := range output {
		// get random item
		random := randomness[pos]

		// random % 64
		randomPos := random % uint8(l)

		// put into output
		output[pos] = letters[randomPos]
	}

	return string(output)
}

// NewRoomCodeID returns a room code.
func NewRoomCodeID() string {
	return randString(Alphabet62, 6)
}

// NewSecretID returns a secretID.
func NewSecretID() string {
	return randString(Alphabet62, 36)
}

// NewSecretKey returns a secretKey or password.
func NewSecretKey() string {
	return randString(Alphabet62, 32)
}

// NewNonce returns a nonce.
func NewNonce() string {
	return randString(Alphabet36, 36)
}

func NewUUID() string {
	return uuid.Must(uuid.NewV4()).String()
}

// 没有中缸线
func NewNoCylinderLineID() string {
	uid := NewUUID()
	return strings.Replace(uid, "-", "", -1)
}

// NewTraceID16 returns a 16-character trace ID.
func NewTraceID16() string {
	return randString(Alphabet16, 16)
}
