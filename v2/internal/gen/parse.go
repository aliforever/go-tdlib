package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// field is one TL constructor field: name:type plus its doc text.
type field struct {
	name string // TL name, snake_case (used verbatim as the JSON tag)
	typ  string // TL type, e.g. int32, vector<MessageSender>, formattedText
	doc  string
}

// ctor is one TL constructor line: name f:t ... = Result;
type ctor struct {
	name   string // TL name, camelCase (e.g. messageText, getMe)
	result string // TL boxed result type (e.g. MessageContent, Ok)
	doc    string
	fields []field
}

// scheme is the parsed td_api.tl model.
type scheme struct {
	entities    []ctor            // types section, minus skipped plumbing
	functions   []ctor            // functions section (RPC requests)
	classes     map[string]string // declared //@class name -> description
	updateTypes []string          // sorted @type names of Update variants
	byName      map[string]ctor   // entity constructor name -> ctor
}

// TL scalar -> Go type mapping, kept explicit on purpose: int53 and int64
// both map to Go int64 but are distinct TL types.
var scalarTypes = map[string]string{
	"int32":  "int32",
	"int53":  "int64",
	"int64":  "int64",
	"double": "float64",
	"string": "string",
	"bytes":  "[]byte",
	"Bool":   "bool",
}

// result types handled by the hand-written core or skipped entirely.
var skipResults = map[string]bool{
	"Double": true, "String": true, "Int32": true, "Int53": true,
	"Int64": true, "Bytes": true, "Bool": true, "Vector": true,
}

// constructors skipped by name: error lives in core.go, boolFalse/boolTrue
// are Bool plumbing.
var skipCtors = map[string]bool{
	"error": true, "boolFalse": true, "boolTrue": true,
}

var (
	classRe    = regexp.MustCompile(`^//@class (\w+)(?:\s+@description\s+(.*))?$`)
	ctorRe     = regexp.MustCompile(`^(\w+)((?:\s+\w+:[\w<>]+)*)\s*=\s*(\w+)(?:\s+\w+)?\s*;$`)
	fieldRe    = regexp.MustCompile(`^(\w+):([\w<>]+)$`)
	plumbingRe = regexp.MustCompile(`^(\w+ \? = \w+|vector \{t:Type\} # \[ t \] = Vector t);$`)
)

func parseScheme(path string) (*scheme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	s := &scheme{classes: map[string]string{}}
	section := "types"

	// pending doc state for the next constructor
	var desc []string
	fieldDocs := map[string][]string{}
	// continuation target for //- lines: either the description or one
	// field doc entry
	lastIsDesc := false
	lastFieldKey := ""

	flush := func(c *ctor) {
		c.doc = strings.Join(desc, "\n")
		for i := range c.fields {
			if d, ok := fieldDocs[c.fields[i].name]; ok {
				c.fields[i].doc = strings.Join(d, "\n")
			}
		}
		desc = nil
		fieldDocs = map[string][]string{}
		lastIsDesc = false
		lastFieldKey = ""
	}

	for ln, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimRight(raw, "\r")
		switch {
		case strings.HasPrefix(line, "---functions---"):
			section = "functions"
			desc, fieldDocs, lastIsDesc, lastFieldKey = nil, map[string][]string{}, false, ""
		case classRe.MatchString(line):
			m := classRe.FindStringSubmatch(line)
			s.classes[m[1]] = strings.TrimSpace(m[2])
			lastIsDesc, lastFieldKey = false, ""
		case strings.HasPrefix(line, "//@"):
			body := strings.TrimPrefix(line, "//@")
			idx := strings.IndexAny(body, " \t")
			var key, text string
			if idx < 0 {
				key = body
			} else {
				key, text = body[:idx], strings.TrimSpace(body[idx+1:])
			}
			if key == "description" {
				desc = append(desc, text)
				lastIsDesc, lastFieldKey = true, ""
			} else {
				fieldDocs[key] = append(fieldDocs[key], text)
				lastIsDesc, lastFieldKey = false, key
			}
		case strings.HasPrefix(line, "//-"):
			text := strings.TrimSpace(strings.TrimPrefix(line, "//-"))
			if lastIsDesc && len(desc) > 0 {
				desc[len(desc)-1] += "\n" + text
			} else if lastFieldKey != "" {
				d := fieldDocs[lastFieldKey]
				if len(d) > 0 {
					d[len(d)-1] += "\n" + text
					fieldDocs[lastFieldKey] = d
				}
			}
		case strings.HasPrefix(line, "//") || strings.TrimSpace(line) == "":
			// free comments and blank lines carry no doc
		default:
			m := ctorRe.FindStringSubmatch(line)
			if m == nil {
				// Known non-constructor plumbing: "double ? = Double;",
				// "vector {t:Type} # [ t ] = Vector t;" and friends.
				if plumbingRe.MatchString(line) {
					continue
				}
				return nil, fmt.Errorf("line %d: unparsable: %q", ln+1, line)
			}
			c := ctor{name: m[1], result: m[3]}
			for _, fs := range strings.Fields(m[2]) {
				fm := fieldRe.FindStringSubmatch(fs)
				if fm == nil {
					return nil, fmt.Errorf("line %d: unparsable field %q", ln+1, fs)
				}
				c.fields = append(c.fields, field{name: fm[1], typ: fm[2]})
			}
			flush(&c)
			if skipCtors[c.name] || skipResults[c.result] {
				continue
			}
			if section == "functions" {
				s.functions = append(s.functions, c)
			} else {
				s.entities = append(s.entities, c)
			}
		}
	}

	sort.Slice(s.entities, func(i, j int) bool { return s.entities[i].name < s.entities[j].name })
	sort.Slice(s.functions, func(i, j int) bool { return s.functions[i].name < s.functions[j].name })

	for _, c := range s.entities {
		if c.result == "Update" {
			s.updateTypes = append(s.updateTypes, c.name)
		}
	}
	sort.Strings(s.updateTypes)

	s.byName = map[string]ctor{}
	for _, c := range s.entities {
		s.byName[c.name] = c
	}
	return s, nil
}

func pascal(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// goFieldName converts a snake_case TL field name to an exported Go name:
// user_id -> UserId, can_be_edited -> CanBeEdited.
func goFieldName(s string) string {
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		b.WriteString(pascal(p))
	}
	return b.String()
}

// goType maps a TL field/result type to its Go type.
//   - scalars per scalarTypes
//   - vector<T> -> []T (recursively)
//   - declared abstract class C -> the C union container (value)
//   - lowercase constructor ref c -> resolved through its result type
//   - concrete boxed type T -> *T
func (s *scheme) goType(typ string) string { return s.goTypeIn(typ, "") }

// goTypeIn is goType with an optional package qualifier ("" inside the td
// package, "td." for the facade in the root tdlib package). Scalars are
// never qualified; every td-defined type name is.
func (s *scheme) goTypeIn(typ, qual string) string {
	// vector<int64> needs the named Int64List type: TDLib quotes int64
	// values in JSON, and the ,string tag option does not reach slice
	// elements. Deeper nesting does not occur in td_api.tl and is rejected
	// by checkType.
	if typ == "vector<int64>" {
		return qual + "Int64List"
	}
	if strings.HasPrefix(typ, "vector<") && strings.HasSuffix(typ, ">") {
		return "[]" + s.goTypeIn(typ[len("vector<"):len(typ)-1], qual)
	}
	if g, ok := scalarTypes[typ]; ok {
		return g
	}
	if _, ok := s.classes[typ]; ok {
		return qual + typ
	}
	if typ == "error" {
		return "*" + qual + "Error" // bare ref to the hand-written error constructor
	}
	if typ[0] >= 'a' && typ[0] <= 'z' {
		// bare constructor reference (e.g. text:formattedText)
		if c, ok := s.byName[typ]; ok {
			return s.goTypeIn(c.result, qual)
		}
	}
	return "*" + qual + typ
}

// goResult maps a function result type R for use in Req[R]: union containers
// by value, concrete structs by pointer, scalars as-is. This is exactly
// goType; the separate name documents intent at the call site.
func (s *scheme) goResult(typ string) string { return s.goType(typ) }

// jsonTagFor returns the JSON struct tag for a field. Scalar int64 fields
// get the ,string option because TDLib quotes int64 values in JSON; int53
// (also Go int64) stays a plain number, so the option is keyed on the exact
// TL type "int64". vector<int64> maps to Int64List instead (see goTypeIn).
func jsonTagFor(f field) string {
	if f.typ == "int64" {
		return f.name + ",string"
	}
	return f.name
}

// validate performs consistency checks that would otherwise surface as
// broken generated code.
func (s *scheme) validate() error {
	reserved := map[string]bool{
		"Req": true, "Request": true, "Union": true, "Error": true,
		"NewUnion": true, "RegisterUnion": true,
		"MarshalWithType": true, "MarshalRequest": true,
		"Int64List": true,
	}
	goNames := map[string]string{} // Go name -> origin
	ctorByName := map[string]ctor{}
	for _, c := range s.entities {
		ctorByName[c.name] = c
	}
	for _, c := range s.entities {
		gn := pascal(c.name)
		if reserved[gn] {
			return fmt.Errorf("constructor %s maps to reserved Go name %s", c.name, gn)
		}
		if prev, dup := goNames[gn]; dup {
			return fmt.Errorf("Go name collision: %s and %s both -> %s", prev, c.name, gn)
		}
		goNames[gn] = c.name

		seen := map[string]bool{}
		for _, f := range c.fields {
			gf := goFieldName(f.name)
			if gf == "TDType" {
				return fmt.Errorf("%s: field %s collides with TDType method", c.name, f.name)
			}
			if seen[gf] {
				return fmt.Errorf("%s: duplicate Go field name %s", c.name, gf)
			}
			seen[gf] = true
			if err := s.checkType(f.typ, ctorByName); err != nil {
				return fmt.Errorf("%s: field %s: %w", c.name, f.name, err)
			}
		}
	}
	for _, c := range s.functions {
		gn := pascal(c.name)
		if prev, dup := goNames[gn]; dup {
			return fmt.Errorf("Go name collision: %s and %s both -> %s", prev, c.name, gn)
		}
		goNames[gn] = c.name
		for _, f := range c.fields {
			if err := s.checkType(f.typ, ctorByName); err != nil {
				return fmt.Errorf("%s: field %s: %w", c.name, f.name, err)
			}
		}
		if err := s.checkType(c.result, ctorByName); err != nil {
			return fmt.Errorf("function %s: result: %w", c.name, err)
		}
	}
	for class := range s.classes {
		if reserved[class] {
			return fmt.Errorf("class %s collides with core identifier", class)
		}
		if prev, dup := goNames[class]; dup {
			return fmt.Errorf("class %s collides with struct %s", class, prev)
		}
		goNames[class] = class
	}
	// The exported variant aliases emitted for the facade/root package must
	// not collide with any generated type either.
	for class := range exportedVariantAliases {
		alias := class + "Variant"
		if prev, dup := goNames[alias]; dup {
			return fmt.Errorf("variant alias %s collides with %s", alias, prev)
		}
		goNames[alias] = class + " variant alias"
	}
	return nil
}

func (s *scheme) checkType(typ string, ctors map[string]ctor) error {
	if strings.HasPrefix(typ, "vector<") && strings.HasSuffix(typ, ">") {
		inner := typ[len("vector<") : len(typ)-1]
		if inner == "int64" {
			return nil // maps to the named Int64List type
		}
		if strings.Contains(inner, "int64") {
			return fmt.Errorf("nested vector<int64> %q is not supported", typ)
		}
		return s.checkType(inner, ctors)
	}
	if _, ok := scalarTypes[typ]; ok {
		return nil
	}
	if _, ok := s.classes[typ]; ok {
		return nil
	}
	if typ == "Error" || typ == "error" {
		return nil // hand-written in core.go
	}
	if typ[0] >= 'a' && typ[0] <= 'z' {
		if _, ok := ctors[typ]; ok {
			return nil
		}
		return fmt.Errorf("unknown bare constructor reference %q", typ)
	}
	for _, c := range s.entities {
		if c.result == typ {
			return nil
		}
	}
	return fmt.Errorf("unknown type %q", typ)
}
