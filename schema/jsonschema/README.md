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

## Authored field constraints

Go field tags can describe existing bounds in MCP input and output schemas.

```go
type Search struct {
    Names []string `json:"names,omitempty" minItems:"1" maxItems:"50"`
    Code  *string  `json:"code,omitempty" minLength:"1" maxLength:"64" pattern:"^[A-Z0-9]+$"`
    Score float64  `json:"score" minimum:"0" maximum:"1"`
}
```

Use `minItems` and `maxItems` for arrays, `minLength`, `maxLength`, and
`pattern` for strings, and `minimum` and `maximum` for numbers and integers.
Bounds are applied to the compiled JSON representation. Optional pointers
retain their nullability. Patterns use the JSON Schema regular expression
dialect and are carried into discovery as authored.

The reflector reads the effective field tags after visibility selection.
Malformed bounds, inverted ranges, and constraints incompatible with the
visible wire type fail compilation with the field path. The owner remains
responsible for validating requests.

The legacy `StructToProperties` adapter uses the same tags. Its signature has
no error return, so invalid metadata produces an unsatisfiable field schema.
