package sequence

import (
    "sync"
    "bytes"
    "errors"
    "encoding/json"
)

type Seq struct {
    sync.Mutex
    value []byte
    free  [][]byte
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
    s,_ := NewSeqFrom([]byte{'a'})
    return s
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
func (s *Seq) Next() []byte {
    s.Lock()
    defer s.Unlock()

    if s.free != nil {
        lim := len(s.free) - 1
        pop := s.free[lim]
        s.free = s.free[:lim]
        if len(s.free) == 0 {
            s.free = nil // attempt to gc mem
        }
        return pop
    }

    next := make([]byte, len(s.value))
    copy(next, s.value)

    s.sum()

    return next
}

// Frees an ID that has already been returned with
// Next(); the free'd ID will be returned on the
// next Next() call.
func (s *Seq) Free(value []byte) error {
    s.Lock()
    defer s.Unlock()

    if cmp := Cmp(value, s.value); cmp >= 0 || s.beenFreed(value) {
        return ErrNotFreeable
    }

    s.free = append(s.free, value)
    return nil
}

// Implements json.Marshaler.
func (s *Seq) MarshalJSON() ([]byte, error) {
    var buf bytes.Buffer

    s.Lock()
    buf.Write([]byte(`{"curr":"`))
    buf.Write(s.value)
    buf.WriteByte('"')

    if s.free != nil {
        sz := len(s.free) - 1
        buf.Write([]byte(`,"free":[`))
        for i := 0; i < sz; i++ {
            buf.WriteByte('"')
            buf.Write(s.free[i])
            buf.WriteByte('"')
            buf.WriteByte(',')
        }
        buf.WriteByte('"')
        buf.Write(s.free[sz])
        buf.WriteByte('"')
        buf.WriteByte(']')
    }
    s.Unlock()
    buf.WriteByte('}')

    return buf.Bytes(), nil
}

// Implements json.Unmarshaler.
func (s *Seq) UnmarshalJSON(p []byte) error {
    var obj seqJSON

    if err := json.Unmarshal(p, &obj); err != nil {
        return err
    }

    s.Lock()
    s.value = []byte(obj.Curr)
    if obj.Free != nil && len(obj.Free) > 0 {
        if s.free != nil {
            s.free = s.free[:0]
        }
        sz := len(obj.Free)
        for i := 0; i < sz; i++ {
            s.free = append(s.free, []byte(obj.Free[i]))
        }
    } else {
        s.free = nil
    }
    s.Unlock()

    return nil
}

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
    for i := range s.free {
        if bytes.Equal(value, s.free[i]) {
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
