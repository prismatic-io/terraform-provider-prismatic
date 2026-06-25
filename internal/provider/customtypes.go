package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// These custom string types carry semantic equality so that Required attributes
// whose values the API normalizes on read-back (trimming whitespace, expanding a
// definition) still round-trip without spurious diffs or "inconsistent result"
// errors. Plan modifiers cannot do this for non-Computed attributes.

// ---- normalizedString: equal when values match after trimming whitespace ----

type normalizedStringType struct{ basetypes.StringType }

var _ basetypes.StringTypable = normalizedStringType{}

func (t normalizedStringType) String() string { return "normalizedStringType" }

func (t normalizedStringType) Equal(o attr.Type) bool {
	other, ok := o.(normalizedStringType)
	return ok && t.StringType.Equal(other.StringType)
}

func (t normalizedStringType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return normalizedStringValue{StringValue: in}, nil
}

func (t normalizedStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	v, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type %T", attrValue)
	}
	return normalizedStringValue{StringValue: v}, nil
}

func (t normalizedStringType) ValueType(context.Context) attr.Value { return normalizedStringValue{} }

type normalizedStringValue struct{ basetypes.StringValue }

var _ basetypes.StringValuableWithSemanticEquals = normalizedStringValue{}

func (v normalizedStringValue) Type(context.Context) attr.Type { return normalizedStringType{} }

func (v normalizedStringValue) Equal(o attr.Value) bool {
	other, ok := o.(normalizedStringValue)
	return ok && v.StringValue.Equal(other.StringValue)
}

func (v normalizedStringValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	other, ok := newValuable.(normalizedStringValue)
	if !ok {
		return false, nil
	}
	return strings.TrimSpace(v.ValueString()) == strings.TrimSpace(other.ValueString()), nil
}

// ---- definitionString: equal when the YAML definitions are semantically the same ----

type definitionStringType struct{ basetypes.StringType }

var _ basetypes.StringTypable = definitionStringType{}

func (t definitionStringType) String() string { return "definitionStringType" }

func (t definitionStringType) Equal(o attr.Type) bool {
	other, ok := o.(definitionStringType)
	return ok && t.StringType.Equal(other.StringType)
}

func (t definitionStringType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return definitionStringValue{StringValue: in}, nil
}

func (t definitionStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	v, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type %T", attrValue)
	}
	return definitionStringValue{StringValue: v}, nil
}

func (t definitionStringType) ValueType(context.Context) attr.Value { return definitionStringValue{} }

type definitionStringValue struct{ basetypes.StringValue }

var _ basetypes.StringValuableWithSemanticEquals = definitionStringValue{}

func (v definitionStringValue) Type(context.Context) attr.Type { return definitionStringType{} }

func (v definitionStringValue) Equal(o attr.Value) bool {
	other, ok := o.(definitionStringValue)
	return ok && v.StringValue.Equal(other.StringValue)
}

func (v definitionStringValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	other, ok := newValuable.(definitionStringValue)
	if !ok {
		return false, nil
	}
	// The receiver is the canonical read-back and the argument is the submitted
	// value, so the comparison is one-directional: the submitted definition must be
	// a subset of the canonical one for the two to be considered equal.
	return definitionsEquivalent(other.ValueString(), v.ValueString()), nil
}
