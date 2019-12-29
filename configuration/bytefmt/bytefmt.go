// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2020  InfoMark.org
// Authors: Patrick Wieschollek
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package bytefmt

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

type ByteSize int64

// ToBytes parses a string formatted by ByteSize as bytes.
// Note, this is based on the basis of 2 (no KiB, ...).
func FromString(s string) (ByteSize, error) {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	i := strings.IndexFunc(s, unicode.IsLetter)

	if i == -1 {
		return 0, invalidByteQuantityError
	}

	bytesString, unit := s[:i], s[i:]
	bytes, err := strconv.ParseFloat(bytesString, 64)
	if err != nil || bytes < 0 {
		return 0, invalidByteQuantityError
	}

	switch unit {
	case "eb":
		return ByteSize(bytes * EXABYTE), nil
	case "pb":
		return ByteSize(bytes * PETABYTE), nil
	case "tb":
		return ByteSize(bytes * TERABYTE), nil
	case "gb":
		return ByteSize(bytes * GIGABYTE), nil
	case "mb":
		return ByteSize(bytes * MEGABYTE), nil
	case "kb":
		return ByteSize(bytes * KILOBYTE), nil
	case "b":
		return ByteSize(bytes), nil
	default:
		return 0, invalidByteQuantityError
	}
}

const (
	BYTE_UNIT     = "b"
	KILOBYTE_UNIT = "kb"
	MEGABYTE_UNIT = "mb"
	GIGABYTE_UNIT = "gb"
	TERABYTE_UNIT = "tb"
	PETABYTE_UNIT = "pb"
	EXABYTE_UNIT  = "eb"
)

const (
	BYTE     = 1
	KILOBYTE = 1024
	MEGABYTE = 1024 * 1024
	GIGABYTE = 1024 * 1024 * 1024
	TERABYTE = 1024 * 1024 * 1024 * 1024
	PETABYTE = 1024 * 1024 * 1024 * 1024 * 1024
	EXABYTE  = 1024 * 1024 * 1024 * 1024 * 1024 * 1024
)

var invalidByteQuantityError = errors.New("byte quantity must be a positive integer with a unit of measurement like b, kb, mb, gb, tb, pb or eb")

// ByteSize returns a human-readable byte string of the form 10M, 12.5K, and so forth.  The following units are available:
// The unit that results in the smallest number greater than or equal to 1 is always chosen.
func ToString(bytes ByteSize) string {
	unit := ""
	value := float64(bytes)

	switch {
	case bytes >= EXABYTE:
		unit = EXABYTE_UNIT
		value = value / EXABYTE
	case bytes >= PETABYTE:
		unit = PETABYTE_UNIT
		value = value / PETABYTE
	case bytes >= TERABYTE:
		unit = TERABYTE_UNIT
		value = value / TERABYTE
	case bytes >= GIGABYTE:
		unit = GIGABYTE_UNIT
		value = value / GIGABYTE
	case bytes >= MEGABYTE:
		unit = MEGABYTE_UNIT
		value = value / MEGABYTE
	case bytes >= KILOBYTE:
		unit = KILOBYTE_UNIT
		value = value / KILOBYTE
	case bytes >= BYTE:
		unit = BYTE_UNIT
	case bytes == 0:
		return "0b"
	}

	result := strconv.FormatFloat(value, 'f', 1, 64)
	result = strings.TrimSuffix(result, ".0")
	return result + unit
}

func (t ByteSize) MarshalYAML() (interface{}, error) {
	return ToString(t), nil
}

func (f *ByteSize) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var fm string
	var err error
	if err := unmarshal(&fm); err != nil {
		return err
	}
	*f, err = FromString(fm)
	return err
}
