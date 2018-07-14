# go-seq

Painless alphabetical sequential IDs.

# Documentation

Visit [godoc](https://godoc.org/github.com/sugoiuguu/go-seq).

# Get this package

```
$ go get github.com/sugoiuguu/go-seq
```

# Example

## New

```go
// create a new sequence
seq := sequence.NewSeq()

fmt.Printf("%q\n", seq.Next())
```

## Encode

```go
// create a new sequence
seq := sequence.NewSeq()

// create a new encoder that outputs to stdout
dec := json.NewEncoder(os.Stdout)

if err := enc.Encode(seq); err != nil {
    panic(err)
}
```

## Decode

```go
// the sequence we'll initialize
var seq sequence.Seq

// create a new decoder that reads from stdin
dec := json.NewDecoder(os.Stdin)

if err := dec.Decode(&seq); err != nil {
    panic(err)
}

fmt.Printf("%q\n", seq.Next())
```
