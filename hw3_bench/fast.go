package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

import (
	json "encoding/json"

	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

type User struct {
	Browsers []string
	Email    string
	Name     string
}

func easyjson9e1087fdDecodeGithubComSinimaWathCourseraHwHw3BenchObject(in *jlexer.Lexer, out *User) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "email":
			out.Email = string(in.String())
		case "name":
			out.Name = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson9e1087fdEncodeGithubComSinimaWathCourseraHwHw3BenchObject(out *jwriter.Writer, in User) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"browsers\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.Browsers == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.Browsers {
				if v2 > 0 {
					out.RawByte(',')
				}
				out.String(string(v3))
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"email\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Email))
	}
	{
		const prefix string = ",\"name\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Name))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v User) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson9e1087fdEncodeGithubComSinimaWathCourseraHwHw3BenchObject(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v User) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9e1087fdEncodeGithubComSinimaWathCourseraHwHw3BenchObject(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *User) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9e1087fdDecodeGithubComSinimaWathCourseraHwHw3BenchObject(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *User) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9e1087fdDecodeGithubComSinimaWathCourseraHwHw3BenchObject(l, v)
}

const (
	userDelimiter       string = "\n"
	androidBrowserRegex        = "Android"
	msieBrowserRegex           = "MSIE"
)

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	seenBrowsers := make(map[string]struct{})
	user := &User{}

	isAndroid := false
	isMSIE := false

	scanner := bufio.NewScanner(file)
	fmt.Fprintln(out, "found users:")

	for i := 0; scanner.Scan(); i++ {

		if err := user.UnmarshalJSON(scanner.Bytes()); err != nil {
			panic(err)
		}

		isAndroid = false
		isMSIE = false

		for _, browser := range user.Browsers {

			if strings.Contains(browser, androidBrowserRegex) {
				isAndroid = true
				if _, ok := seenBrowsers[browser]; !ok {
					seenBrowsers[browser] = struct{}{}

				}
			} else if strings.Contains(browser, msieBrowserRegex) {
				isMSIE = true
				if _, ok := seenBrowsers[browser]; !ok {
					seenBrowsers[browser] = struct{}{}
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		fmt.Fprintf(out, "[%d] %s <%s>\n", i, user.Name, strings.Replace(user.Email, "@", " [at] ", 1))
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	fmt.Fprintf(out, "\n")
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}

func main() {
}
