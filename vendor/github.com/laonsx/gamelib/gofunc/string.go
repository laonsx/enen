package gofunc

import (
	"crypto/md5"
	"encoding/hex"
)

//SubStr 截取字符串
func SubStr(s string, start int, length ...int) string {

	rs := []rune(s)
	var d []rune

	if len(length) == 1 {

		l := len(rs)

		if start < 0 {

			start = l + start
		}

		end := start + length[0]
		if end > l {

			end = l
		}

		d = rs[start:end]
	} else {

		if start < 0 {

			start = len(rs) + start
			d = rs[start:]
		}

		d = rs[start:]
	}

	return string(d)
}

//URLEncode 编码 URL 字符串, 按照RFC 3986
func URLEncode(v string) string {

	var buf []byte
	var x, y int
	hexchars := []byte("0123456789ABCDEF")
	val := []byte(v)
	length := len(val)

	for {

		buf = append(buf, val[x])

		if (buf[y] < '0' && buf[y] != '-' && buf[y] != '.') ||
			(buf[y] < 'A' && buf[y] > '9') ||
			(buf[y] > 'Z' && buf[y] < 'a' && buf[y] != '_') || (buf[y] > 'z') {

			buf[y] = '%'
			y++
			a := val[x] >> 4
			buf = append(buf, hexchars[a])
			y++
			b := val[x] & 15
			buf = append(buf, hexchars[b])
		}

		length--
		if length == 0 {

			break
		}

		x++
		y++
	}

	return string(buf)
}

func MD5(s string) string {

	h := md5.New()
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}

//StrLen 字符串长度
func StrLen(s string) int {

	var l int
	r := []rune(s)
	for _, v := range r {

		if v > 127 {

			l += 2
		} else {

			l++
		}
	}

	return l
}
