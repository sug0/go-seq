package sequence

import (
    "bytes"
    "errors"
    //"encoding/json"
)

type Seq struct {
    free   []byte // 0, x_0, x_1, x_2, ..., x_n-1, N
    free2  [][]byte
    value  []byte
    value2 [256]byte
}

type seqJSON = struct {
    Curr string   `json:"curr"`
    Free []string `json:"free,omitempty"`
}

const (
    bigEncode = 'A' - 26
    maxSmall  = 'z' - 'a'
    maxSmall2 = maxSmall + 1
    maxLarge  = (maxSmall + 1) + ('Z' - 'A')
    maxLarge2 = maxLarge + 1
)

var (
    ErrInvalidByte = errors.New("sequence: invalid byte in initializer")
    ErrNotFreeable = errors.New("sequence: can't free value")
)

// Compare two IDs, x and y; if x == y, this
// function returns 0; if x > y, it returns 1;
// if x < y, it returns -1.
func Cmp(x, y []byte) int {
    switch {
    case len(x) > len(y):
        return 1
    case len(y) > len(x):
        return -1
    }

    for i := range x {
        xi, yi := decode(x[i]), decode(y[i])
        switch {
        case xi > yi:
            return 1
        case yi > xi:
            return -1
        }
    }
    return 0
}

// Returns a new sequence starting at 'a'.
func NewSeq() *Seq {
    var s Seq
    s.value2[0] = 'a'
    s.value = s.value2[:1]
    return &Seq{value: []byte{'a'}}
}

// Returns a new sequence starting at a
// particular value.
func NewSeqFrom(value []byte) (*Seq, error) {
    for _,b := range value {
        if (b < 'a' || b > 'z') && (b < 'A' || b > 'Z') {
            return nil, ErrInvalidByte
        }
    }
    return &Seq{value: value}, nil
}

// Returns the next ID in the sequence.
func (s *Seq) Next(nextBuf []byte) []byte {
    var value []byte

    if len(s.value) < 256 && len(s.free) != 0 {
        lim := int(s.free[len(s.free) - 1])
        value = s.free[len(s.free)-lim-1:len(s.free)-1]
        s.free = s.free[:len(s.free)-lim-2]
    } else if len(s.value) > 255 && len(s.free2) != 0 {
        lim := len(s.free2) - 1
        value = s.free2[lim]
        s.free2 = s.free2[:lim]
        if len(s.free2) == 0 {
            s.free = nil // attempt to gc mem
            s.free2 = nil
        }
    } else {
        value = s.value
    }

    nextBuf = append(nextBuf, value...)
    s.sum()

    return nextBuf
}

// Frees an ID that has already been returned with
// Next(); the free'd ID will be returned on the
// next Next() call.
func (s *Seq) Free(value []byte) error {
    if Cmp(value, s.value) >= 0 || s.beenFreed(value) {
        return ErrNotFreeable
    }
    if len(s.value) < 256 {
        s.free = append(s.free, 0)
        s.free = append(s.free, value...)
        s.free = append(s.free, byte(len(value)))
    } else {
        s.free2 = append(s.free2, value)
    }
    return nil
}

//     // Implements json.Marshaler.
//     func (s *Seq) MarshalJSON() ([]byte, error) {
//         var buf bytes.Buffer
//     
//         buf.Write([]byte(`{"curr":"`))
//         buf.Write(s.value)
//         buf.WriteByte('"')
//     
//         if s.free != nil {
//             sz := len(s.free) - 1
//             buf.Write([]byte(`,"free":[`))
//             for i := 0; i < sz; i++ {
//                 buf.WriteByte('"')
//                 buf.Write(s.free[i])
//                 buf.WriteByte('"')
//                 buf.WriteByte(',')
//             }
//             buf.WriteByte('"')
//             buf.Write(s.free[sz])
//             buf.WriteByte('"')
//             buf.WriteByte(']')
//         }
//         buf.WriteByte('}')
//     
//         return buf.Bytes(), nil
//     }
//     
//     // Implements json.Unmarshaler.
//     func (s *Seq) UnmarshalJSON(p []byte) error {
//         var obj seqJSON
//     
//         if err := json.Unmarshal(p, &obj); err != nil {
//             return err
//         }
//     
//         s.value = []byte(obj.Curr)
//         if obj.Free != nil && len(obj.Free) > 0 {
//             if s.free != nil {
//                 s.free = s.free[:0]
//             }
//             sz := len(obj.Free)
//             for i := 0; i < sz; i++ {
//                 s.free = append(s.free, []byte(obj.Free[i]))
//             }
//         } else {
//             s.free = nil
//         }
//     
//         return nil
//     }

func (s *Seq) sum() {
    d := s.value

    for {
        if d[0] != 'Z' {
            break
        }
        if len(d) == 1 {
            d[0] = 'a'
            s.value = append(s.value, 'a')
            return
        }
        d[0] = 'a'
        d = d[1:]
    }

    d[0] = encode((decode(d[0]) + 1) % maxLarge2)
}

func (s *Seq) beenFreed(value []byte) bool {
    if len(s.value) < 256 {
        for i := 0; i < len(s.free); {
            var j int
            for j = i+1; j < len(s.free); j++ {
                if s.free[j] == 0 {
                    break
                }
            }
            thisValue := s.free[i+1:j-1]
            if bytes.Equal(value, thisValue) {
                return true
            }
            i += int(s.free[j-1])
        }
        return false
    }
    for i := 0; i < len(s.free2); i++ {
        if bytes.Equal(value, s.free2[i]) {
            return true
        }
    }
    return false
}

func decode(d byte) byte {
    if d >= 'a' {
        return d - 'a'
    }
    return maxSmall2 + (d - 'A')
}

func encode(d byte) byte {
    if d < 26 {
        return d + 'a'
    }
    return d + bigEncode
}
