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

## Authored field nullability

The native reflector honors existing Tagly `format` metadata independently of
presence requirements:

```go
type Options struct {
    Preview *bool `json:"preview,omitempty" format:"nullable=false"`
}
```

This publishes an optional boolean, without a null alternative. Untagged pointers
retain their existing nullable schemas. `nullable=true` permits null for the
annotated field; `nullable=false` excludes it. Neither setting changes nested
fields, collection-item policies, referenced definitions, or requiredness.
Enum alternatives retain their existing values while following the field's null
policy. References and compositions use local constraints rather than modifying
shared schema definitions.

As with `wireSchema`, these declarations describe the owner's contract. Schema
compilation does not add a request validator or change decoding. The owning
runtime/application must enforce its declared non-null contract; native presence
metadata can distinguish an omitted pointer from an explicit null. Metadata is
prepared during compilation, without examining requests or executing a decoder.

For enum inputs, `enumNull:"reject"` follows the enum plan's
`AllowsNullInput()` policy. It removes the null alternative even for pointer
storage or `format:"nullable=true"`; output nullability and omission remain
unchanged. This keeps discovery aligned with the selected decoder rather than
using storage pointers as an unconditional input-null grant.
