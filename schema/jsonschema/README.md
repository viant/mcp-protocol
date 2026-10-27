# Custom scalar wire schemas

A field whose type owns custom JSON/text serialization can declare its scalar
wire representation without executing its methods during discovery:

```go
type Filter struct {
    Since *FlexibleTime `json:"since,omitempty" wireSchema:"{\"type\":[\"string\",\"integer\"]}"`
}
```

The declaration describes the existing wire contract; it does not change
marshaling, decoding or validation. Here `FlexibleTime` must actually support the
declared text and integer forms. Pointer fields retain nullability.

`wireSchema` accepts `type` (one scalar or a nonempty union of `string`, `integer`,
`number`, `boolean`, `null`) and an optional string `format`. A format requires a
string member. Ordinary Go fields and standard `time.Time` keep their canonical
schemas. Duplicate declarations/types, quoted JSON conflicts, object/array
schemas, references and other schema keywords fail discovery. Field visibility
is applied first. This tag cannot rename members or expose excluded fields.

`ScalarWireSchema` provides the same declaration parsing for output-schema
owners. No custom serializer or decoder is executed while interpreting the tag.
