package configui

import (
	"slices"
	"strconv"

	"github.com/lost-melody/yuman/fcitx5"
)

const (
	OptionTypeUnknown OptionType = ""
	OptionTypeBool    OptionType = "Boolean"
	OptionTypeInt     OptionType = "Integer"
	OptionTypeString  OptionType = "String"
	OptionTypeEnum    OptionType = "Enum"
	OptionTypeStrings OptionType = "List|String"
)

const (
	OptionKeyIntMax   = "IntMax"
	OptionKeyIntMin   = "IntMin"
	OptionKeyEnum     = "Enum"
	OptionKeyEnumI18n = "EnumI18n"
)

type OptionType string

type OptionValue struct {
	Option  *fcitx5.ConfigOption
	Type    OptionType
	Bool    ValueBool
	Int     ValueInt
	String  ValueString
	Enum    ValueEnum
	Strings ValueStrings
	Unknown string
}

type ValueBool struct {
	Value bool
}

type ValueInt struct {
	Value int
	Min   *int
	Max   *int
}

type ValueString struct {
	Value string
}

type ValueEnum struct {
	Index int
	Enums []string
	Names []string
}

type ValueStrings struct {
	Index int
	Value []string
}

func (v *OptionValue) Set(option *fcitx5.ConfigOption, value any) {
	if v.Option == option {
		return
	}
	v.Option = option

	v.Type = OptionType(option.Type)
	switch v.Type {
	case OptionTypeBool:
		v.Bool.Parse(option, value)
	case OptionTypeInt:
		v.Int.Parse(option, value)
	case OptionTypeString:
		v.String.Parse(option, value)
	case OptionTypeEnum:
		v.Enum.Parse(option, value)
	case OptionTypeStrings:
		v.Strings.Parse(option, value)
	default:
	}
}

func (v *ValueBool) Parse(option *fcitx5.ConfigOption, value any) {
	val, _ := value.(string)
	v.Value, _ = strconv.ParseBool(val)
}

func (v *ValueBool) Toggle(options map[string]any, option *fcitx5.ConfigOption) {
	v.Value = !v.Value
	options[option.Name] = strconv.FormatBool(v.Value)
}

func (v *ValueInt) Parse(option *fcitx5.ConfigOption, value any) {
	val, _ := value.(string)
	i, _ := strconv.ParseInt(val, 10, 64)
	v.Value = int(i)

	intMin, ok := option.Extras[OptionKeyIntMin].(string)
	if ok {
		i, _ := strconv.ParseInt(intMin, 10, 64)
		v.Min = new(int(i))
	} else {
		v.Min = nil
	}
	intMax, ok := option.Extras[OptionKeyIntMax].(string)
	if ok {
		i, _ := strconv.ParseInt(intMax, 10, 64)
		v.Max = new(int(i))
	} else {
		v.Max = nil
	}
}

func (v *ValueInt) Inc(options map[string]any, option *fcitx5.ConfigOption) {
	if v.Max == nil || v.Value < *v.Max {
		v.Value++
		options[option.Name] = strconv.FormatInt(int64(v.Value), 10)
	}
}

func (v *ValueInt) Dec(options map[string]any, option *fcitx5.ConfigOption) {
	if v.Min == nil || v.Value > *v.Min {
		v.Value--
		options[option.Name] = strconv.FormatInt(int64(v.Value), 10)
	}
}

func (v *ValueInt) SetMin(options map[string]any, option *fcitx5.ConfigOption) {
	if v.Min != nil {
		v.Value = *v.Min
		options[option.Name] = strconv.FormatInt(int64(v.Value), 10)
	}
}

func (v *ValueInt) SetMax(options map[string]any, option *fcitx5.ConfigOption) {
	if v.Max != nil {
		v.Value = *v.Max
		options[option.Name] = strconv.FormatInt(int64(v.Value), 10)
	}
}

func (v *ValueString) Parse(option *fcitx5.ConfigOption, value any) {
	v.Value, _ = value.(string)
}

func (v *ValueString) Apply(options map[string]any, option *fcitx5.ConfigOption) {
	options[option.Name] = v.Value
}

func (v *ValueEnum) Parse(option *fcitx5.ConfigOption, value any) {
	val, _ := value.(string)
	enums, _ := option.Extras[OptionKeyEnum].(map[string]any)
	names, _ := option.Extras[OptionKeyEnumI18n].(map[string]any)
	if enums == nil || names == nil || len(enums) != len(names) {
		v.Index = 0
		v.Enums = []string{"Unknown"}
		v.Names = []string{"Unknown"}
	}
	v.Enums = ParseFcitx5List(enums)
	v.Names = ParseFcitx5List(names)
	for i, e := range v.Enums {
		if e == val {
			v.Index = i
		}
	}
}

func (v *ValueEnum) Prev(options map[string]any, option *fcitx5.ConfigOption) {
	if v.Index > 0 {
		v.Index--
		options[option.Name] = v.Enums[v.Index]
	}
}

func (v *ValueEnum) Next(options map[string]any, option *fcitx5.ConfigOption) {
	if v.Index < len(v.Enums)-1 {
		v.Index++
		options[option.Name] = v.Enums[v.Index]
	}
}

func (v *ValueStrings) Parse(option *fcitx5.ConfigOption, value any) {
	dict, _ := value.(map[string]any)
	v.Value = ParseFcitx5List(dict)
}

func (v *ValueStrings) Prev(options map[string]any, option *fcitx5.ConfigOption) {
	if v.Index > 0 {
		v.Index--
	}
}

func (v *ValueStrings) Next(options map[string]any, option *fcitx5.ConfigOption) {
	if v.Index < len(v.Value)-1 {
		v.Index++
	}
}

func (v *ValueStrings) SwapPrev(options map[string]any, option *fcitx5.ConfigOption) {
	if v.Index > 0 && v.Index < len(v.Value) {
		v.Value[v.Index], v.Value[v.Index-1] = v.Value[v.Index-1], v.Value[v.Index]
		v.Apply(options, option, v.Index-1, v.Index)
		v.Index--
	}
}

func (v *ValueStrings) SwapNext(options map[string]any, option *fcitx5.ConfigOption) {
	if v.Index >= 0 && v.Index < len(v.Value)-1 {
		v.Value[v.Index], v.Value[v.Index+1] = v.Value[v.Index+1], v.Value[v.Index]
		v.Apply(options, option, v.Index, v.Index+1)
		v.Index++
	}
}

func (v *ValueStrings) Insert(options map[string]any, option *fcitx5.ConfigOption) {
	v.Value = slices.Insert(v.Value, v.Index, "")
	v.Apply(options, option, v.Index, len(v.Value)-1)
}

func (v *ValueStrings) Delete(options map[string]any, option *fcitx5.ConfigOption) {
	if v.Index >= len(v.Value) {
		return
	}
	v.Value = slices.Delete(v.Value, v.Index, v.Index+1)
	v.Apply(options, option, v.Index, len(v.Value))
	if v.Index > 0 && v.Index == len(v.Value) {
		v.Index--
	}
}

func (v *ValueStrings) Apply(options map[string]any, option *fcitx5.ConfigOption, i, j int) {
	dict, _ := options[option.Name].(map[string]any)
	if dict == nil {
		dict = map[string]any{}
		options[option.Name] = dict
	}
	if i > j {
		i, j = j, i
	}
	for idx := i; idx <= j; idx++ {
		key := strconv.FormatInt(int64(idx), 10)
		if idx < len(v.Value) {
			dict[key] = v.Value[idx]
		} else {
			delete(dict, key)
		}
	}
}

func ParseFcitx5List(dict map[string]any) (list []string) {
	if dict == nil {
		return
	}
	list = make([]string, len(dict))
	for k, v := range dict {
		idx, _ := strconv.ParseInt(k, 10, 64)
		value, _ := v.(string)
		list[int(idx)] = value
	}
	return
}
