package structor

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"math"
	"net"
	"net/mail"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/illikainen/go-utils/src/stringx"

	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

// ------
// Base64
// ------
type validateBase64 struct {
	opts *Options
}

func (v *validateBase64) Init(opts *Options) error {
	v.opts = opts
	return nil
}

func (v *validateBase64) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	_, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return errors.Wrapf(ErrValidation, "%s: invalid base64", name)
	}

	return nil
}

func (v *validateBase64) NumArgs() (int, int) {
	return 0, 0
}

// ------
// Either
// ------
type validateEither struct {
	opts *Options
}

func (v *validateEither) Init(opts *Options) error {
	v.opts = opts
	return nil
}

func (v *validateEither) Validate(name string, value reflect.Value, args []string, root reflect.Value) error {
	if value.IsValid() && !value.IsZero() {
		return nil
	}

	for _, arg := range args {
		v := root.FieldByName(arg)
		if v.IsValid() && !v.IsZero() {
			return nil
		}
	}

	return errors.Wrapf(ErrValidation, "one of the following fields are required: %s, %s",
		name, strings.Join(args, ", "))
}

func (v *validateEither) NumArgs() (int, int) {
	return 1, -1
}

// -----
// Email
// -----
type validateEmail struct {
	opts *Options
}

func (v *validateEmail) Init(opts *Options) error {
	v.opts = opts
	return nil
}

func (v *validateEmail) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	_, err := mail.ParseAddress(s)
	if err != nil {
		return errors.Wrapf(ErrValidation, "%s: '%s' not a valid email", name, s)
	}

	return nil
}

func (v *validateEmail) NumArgs() (int, int) {
	return 0, 0
}

// ------
// Exists
// ------
type validateExists struct {
	opts *Options
}

func (v *validateExists) Init(opts *Options) error {
	v.opts = opts
	return nil
}

func (v *validateExists) Validate(name string, value reflect.Value, args []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	info, err := os.Stat(s)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.Wrapf(ErrValidation, "%s: '%s' does not exist", name, s)
		}
		return errors.WithStack(err)
	}

	if slices.Contains(args, "dir") {
		if !info.IsDir() {
			return errors.Wrapf(ErrValidation, "%s: '%s' is not a directory", name, s)
		}
	} else if slices.Contains(args, "file") {
		if info.IsDir() {
			return errors.Wrapf(ErrValidation, "%s: '%s' is not a file", name, s)
		}
	}

	return nil
}

func (v *validateExists) NumArgs() (int, int) {
	return 0, 1
}

// -----
// Group
// -----
type validateGroup struct {
	opts *Options
	rx   *regexp.Regexp
}

func (v *validateGroup) Init(opts *Options) error {
	rx, err := regexp.Compile(`^[a-z_][a-z0-9_-]{0,31}$`)
	if err != nil {
		return errors.WithStack(err)
	}
	v.rx = rx

	v.opts = opts
	return nil
}

func (v *validateGroup) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	if !v.rx.MatchString(s) {
		return errors.Wrapf(ErrValidation, "%s: '%s' is not valid group name", name, s)
	}

	return nil
}

func (v *validateGroup) NumArgs() (int, int) {
	return 0, 0
}

// ---
// Hex
// ---
type validateHex struct {
	opts *Options
}

func (v *validateHex) Init(opts *Options) error {
	v.opts = opts
	return nil
}

func (v *validateHex) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	_, err := hex.DecodeString(s)
	if err != nil {
		return errors.Wrapf(ErrValidation, "%s: value is not valid hex", name)
	}

	return nil
}

func (v *validateHex) NumArgs() (int, int) {
	return 0, 0
}

// ----
// Host
// ----
type validateHost struct {
	opts *Options
	rx   *regexp.Regexp
}

func (v *validateHost) Init(opts *Options) error {
	rx, err := regexp.Compile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$`)
	if err != nil {
		return errors.WithStack(err)
	}
	v.rx = rx

	v.opts = opts
	return nil
}

func (v *validateHost) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	if len(s) > 253 {
		return errors.Wrapf(ErrValidation, "%s: '%s' is not valid hostname", name, s)
	}

	for _, elt := range strings.Split(s, ".") {
		if !v.rx.MatchString(elt) {
			return errors.Wrapf(ErrValidation, "%s: '%s' is not valid hostname", name, s)
		}
	}

	return nil
}

func (v *validateHost) NumArgs() (int, int) {
	return 0, 0
}

// --
// In
// --
type validateIn struct {
	opts *Options
}

func (v *validateIn) Init(opts *Options) error {
	v.opts = opts
	return nil
}

func (v *validateIn) Validate(name string, value reflect.Value, args []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	elemType := value.Type()
	if elemType.Kind() == reflect.Slice {
		elemType = elemType.Elem()
	}

	valid := reflect.New(reflect.SliceOf(elemType))
	err := json.Unmarshal([]byte(args[0]), valid.Interface())
	if err != nil {
		return err
	}

	if value.Type().Kind() == reflect.Slice { // revive:disable-line
		for i := 0; i < value.Len(); i++ {
			ok := false
			for j := 0; j < valid.Elem().Len() && !ok; j++ {
				if reflect.DeepEqual(value.Index(i), valid.Elem().Index(j)) {
					ok = true
				}
			}

			if !ok {
				return errors.Wrapf(ErrValidation, "%s: '%v' is not among %s", name, value, args[0])
			}
		}
	} else {
		for i := 0; i < valid.Elem().Len(); i++ {
			if reflect.DeepEqual(value.Interface(), valid.Elem().Index(i).Interface()) {
				return nil
			}
		}
		return errors.Wrapf(ErrValidation, "%s: '%v' is not among %s", name, value, args[0])
	}

	return nil
}

func (v *validateIn) NumArgs() (int, int) {
	return 1, 1
}

// --
// IP
// --
type validateIP struct {
	opts *Options
}

func (v *validateIP) Init(opts *Options) error {
	v.opts = opts
	return nil
}

func (v *validateIP) Validate(name string, value reflect.Value, args []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	var ip net.IP
	if slices.Contains(args, "autocidr") {
		tmp, _, err := net.ParseCIDR(s)
		if err == nil {
			ip = tmp
		} else {
			ip = net.ParseIP(s)
		}
	} else if slices.Contains(args, "cidr") {
		tmp, _, err := net.ParseCIDR(s)
		if err != nil {
			return errors.Wrapf(ErrValidation, "%s: '%s' not a valid CIDR", name, s)
		}
		ip = tmp
	} else {
		ip = net.ParseIP(s)
	}

	if ip == nil {
		return errors.Wrapf(ErrValidation, "%s: '%s' not a valid IP", name, s)
	}

	if slices.Contains(args, "v4") && ip.To4() == nil {
		return errors.Wrapf(ErrValidation, "%s: '%s' not a valid IPv4", name, s)
	} else if slices.Contains(args, "v6") && ip.To4() != nil {
		return errors.Wrapf(ErrValidation, "%s: '%s' not a valid IPv6", name, s)
	}

	return nil
}

func (v *validateIP) NumArgs() (int, int) {
	return 0, -1
}

// ---
// Max
// ---
type validateMax struct {
	opts *Options
}

func (v *validateMax) Init(opts *Options) error {
	v.opts = opts
	return nil
}

func (v *validateMax) Validate(name string, value reflect.Value, args []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		max, err := strconv.ParseInt(args[0], 10, reflect.TypeOf(int64(0)).Bits())
		if err != nil {
			return err
		}

		if value.Int() > max {
			return errors.Wrapf(ErrValidation, "%s: %d must be less than %d", name, value.Int(), max)
		}
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		max, err := strconv.ParseUint(args[0], 10, reflect.TypeOf(int64(0)).Bits())
		if err != nil {
			return err
		}

		if value.Uint() > max {
			return errors.Wrapf(ErrValidation, "%s: %d must be less than %d", name, value.Uint(), max)
		}
		return nil
	}

	return errors.Errorf("%s: %s cannot be used with min", name, value.Type())
}

func (v *validateMax) NumArgs() (int, int) {
	return 1, 1
}

// ---
// Min
// ---
type validateMin struct {
	opts *Options
}

func (v *validateMin) Init(opts *Options) error {
	v.opts = opts
	return nil
}

func (v *validateMin) Validate(name string, value reflect.Value, args []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		min, err := strconv.ParseInt(args[0], 10, reflect.TypeOf(int64(0)).Bits())
		if err != nil {
			return err
		}

		if value.Int() < min {
			return errors.Wrapf(ErrValidation, "%s: %d must be more than %d", name, value.Int(), min)
		}
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		min, err := strconv.ParseUint(args[0], 10, reflect.TypeOf(int64(0)).Bits())
		if err != nil {
			return err
		}

		if value.Uint() < min {
			return errors.Wrapf(ErrValidation, "%s: %d must be more than %d", name, value.Uint(), min)
		}
		return nil
	}

	return errors.Errorf("%s: %s cannot be used with min", name, value.Type())
}

func (v *validateMin) NumArgs() (int, int) {
	return 1, 1
}

// ----
// Port
// ----
type validatePort struct {
	opts *Options
}

func (v *validatePort) Init(opts *Options) error {
	v.opts = opts
	return nil
}

func (v *validatePort) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v := value.Int()
		if v < 0 || v > math.MaxUint16 {
			return errors.Wrapf(ErrValidation, "%s: %d is not a valid port", name, v)
		}
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v := value.Uint()
		if v > math.MaxUint16 {
			return errors.Wrapf(ErrValidation, "%s: %d is not a valid port", name, v)
		}
		return nil
	case reflect.String:
		v := value.String()
		elts := strings.SplitN(v, "-", 2)

		start, err := strconv.ParseUint(elts[0], 10, reflect.TypeOf(int64(0)).Bits())
		if err != nil {
			return errors.WithStack(err)
		}
		if start > math.MaxUint16 {
			return errors.Wrapf(ErrValidation, "%s: %s is not a valid port", name, v)
		}

		if len(elts) == 2 {
			end, err := strconv.ParseUint(elts[1], 10, reflect.TypeOf(int64(0)).Bits())
			if err != nil {
				return errors.WithStack(err)
			}

			if end < start || end > math.MaxUint16 {
				return errors.Wrapf(ErrValidation, "%s: %s is not a valid port rage", name, v)
			}
		}
		return nil
	}

	return errors.Errorf("%s: %s cannot be used with port", name, value.Type())
}

func (v *validatePort) NumArgs() (int, int) {
	return 0, 0
}

// ---------
// Printable
// ---------
type validatePrintable struct {
	opts *Options
}

func (v *validatePrintable) Init(opts *Options) error {
	v.opts = opts
	return nil
}

func (v *validatePrintable) Validate(name string, value reflect.Value, args []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	if slices.Contains(args, "unicode") {
		for _, r := range s {
			if !unicode.IsPrint(r) {
				return errors.Wrapf(ErrValidation, "%s: value is not printable unicode", name)
			}
		}
	} else {
		if !stringx.IsPrintable(s) {
			return errors.Wrapf(ErrValidation, "%s: value is not printable ascii", name)
		}
	}

	return nil
}

func (v *validatePrintable) NumArgs() (int, int) {
	return 0, 1
}

// --------
// Required
// --------
type validateRequired struct {
	opts *Options
}

func (v *validateRequired) Init(opts *Options) error {
	v.opts = opts
	return nil
}

func (v *validateRequired) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return errors.Wrapf(ErrValidation, "missing required field: %s", name)
	}
	return nil
}

func (v *validateRequired) NumArgs() (int, int) {
	return 0, 0
}

// ------
// Regexp
// ------
type validateRegexp struct {
	opts *Options
}

func (v *validateRegexp) Init(opts *Options) error {
	v.opts = opts
	return nil
}

func (v *validateRegexp) Validate(name string, value reflect.Value, args []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	rx, err := regexp.Compile(args[0])
	if err != nil {
		return errors.WithStack(err)
	}

	if !rx.MatchString(s) {
		return errors.Wrapf(ErrValidation, "%s: '%s' does not match pattern '%s'", name, s, args[0])
	}
	return nil
}

func (v *validateRegexp) NumArgs() (int, int) {
	return 1, 1
}

// ---
// URI
// ---
type validateURI struct {
	opts *Options
}

func (v *validateURI) Init(opts *Options) error {
	v.opts = opts
	return nil
}

func (v *validateURI) Validate(name string, value reflect.Value, args []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	var uri *url.URL
	if u, ok := value.Interface().(*url.URL); ok {
		uri = u
	} else if s, ok := value.Interface().(string); ok {
		u, err := url.ParseRequestURI(s)
		if err != nil {
			return errors.Wrapf(ErrValidation, "%s: '%s' not a valid URI", name, s)
		}
		uri = u
	} else {
		return errors.Errorf("%s: %s is not a string or a *net.URL", name, value.Type())
	}

	schemes := argsByName("scheme", args)
	if schemes != nil && !slices.Contains(schemes, uri.Scheme) {
		return errors.Wrapf(ErrValidation, "%s: '%s' does not have a valid URI scheme", name, uri)
	}

	return nil
}

func (v *validateURI) NumArgs() (int, int) {
	return 0, -1
}

// ----
// User
// ----
type validateUser struct {
	opts *Options
	rx   *regexp.Regexp
}

func (v *validateUser) Init(opts *Options) error {
	rx, err := regexp.Compile(`^[a-z_][a-z0-9_-]{0,31}$`)
	if err != nil {
		return errors.WithStack(err)
	}
	v.rx = rx

	v.opts = opts
	return nil
}

func (v *validateUser) Validate(name string, value reflect.Value, _ []string, _ reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	s, ok := value.Interface().(string)
	if !ok {
		return errors.Errorf("%s: %s is not a string", name, value.Type())
	}

	if !v.rx.MatchString(s) {
		return errors.Wrapf(ErrValidation, "%s: '%s' is not valid username", name, s)
	}

	return nil
}

func (v *validateUser) NumArgs() (int, int) {
	return 0, 0
}
