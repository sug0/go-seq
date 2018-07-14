# go-seq

Painless alphabetical sequential IDs.

# Example

## New

```go
seq := sequence.NewSeq()
fmt.Printf("%q\n", seq.Next())
```

## Decode

```go
var seq sequence.Seq

dec := json.NewDecoder(os.Stdin)
if err := dec.Decode(&seq); err != nil {
    panic(err)
}

fmt.Printf("%q\n", seq.Next())
```
