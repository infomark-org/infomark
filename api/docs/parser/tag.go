package parser

import (
    "strconv"
    "strings"
)

type rawTag struct {
    // Key is the tag key, such as json, xml, etc..
    // i.e: `json:"foo,omitempty". Here key is: "json"
    Key string

    // Name is a part of the value
    // i.e: `json:"foo,omitempty". Here name is: "foo"
    Name string

    // Options is a part of the value. It contains a slice of tag options i.e:
    // `json:"foo,omitempty". Here options is: ["omitempty"]
    Options []string
}

type Tag struct {
    Name     string
    Example  string
    Required bool
}

func parseTag(tag string) (*Tag, error) {
    var tags []*rawTag

    tag = tag[1 : len(tag)-1]

    // NOTE(arslan) following code is from reflect and vet package with some
    // modifications to collect all necessary information and extend it with
    // usable methods
    for tag != "" {
        // fmt.Println("parse:", tag)
        if len(tag) < 3 {
            return nil, nil
        }
        // Skip leading space.
        i := 0
        for i < len(tag) && tag[i] == ' ' {
            i++
        }
        tag = tag[i:]
        if tag == "" {
            return nil, nil
        }

        // Scan to colon. A space, a quote or a control character is a syntax
        // error. Strictly speaking, control chars include the range [0x7f,
        // 0x9f], not just [0x00, 0x1f], but in practice, we ignore the
        // multi-byte control characters as it is simpler to inspect the tag's
        // bytes than the tag's runes.
        i = 0
        for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
            i++
        }

        if i == 0 {
            return nil, errTagKeySyntax
        }
        if i+1 >= len(tag) || tag[i] != ':' {
            return nil, errTagSyntax
        }
        if tag[i+1] != '"' {
            return nil, errTagValueSyntax
        }

        key := string(tag[:i])
        tag = tag[i+1:]

        // Scan quoted string to find value.
        i = 1
        for i < len(tag) && tag[i] != '"' {
            if tag[i] == '\\' {
                i++
            }
            i++
        }
        if i >= len(tag) {
            return nil, errTagValueSyntax
        }

        qvalue := string(tag[:i+1])
        tag = tag[i+1:]

        value, err := strconv.Unquote(qvalue)
        if err != nil {
            return nil, errTagValueSyntax
        }

        res := strings.Split(value, ",")
        name := res[0]
        options := res[1:]
        if len(options) == 0 {
            options = nil
        }
        // fmt.Println("got ", key, name)
        tags = append(tags, &rawTag{
            Key:     key,
            Name:    name,
            Options: options,
        })
    }

    field := &Tag{}
    for _, tag := range tags {

        if tag.Key == "json" {
            field.Name = tag.Name
        }

        if tag.Key == "example" {
            field.Example = tag.Name
        }

        if tag.Key == "required" {
            if tag.Name == "false" {
                field.Required = false
            } else {
                field.Required = true
            }
        }
    }

    return field, nil
}
