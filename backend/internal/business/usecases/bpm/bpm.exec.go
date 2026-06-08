// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package bpm

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Энэ slice-ийн илэрхийллийн давхарга нь зориуд хялбар: serviceTask-ийн url/body
// дотор `${var}` орлуулга, exclusive gateway-ийн нөхцөлд `<var> <op> <literal>`
// харьцуулалт. Дараагийн increment-д FEEL / JSONata болгон өргөтгөж болно.

var placeholderRe = regexp.MustCompile(`\$\{([^}]+)\}`)

// substitute нь текст доторх `${name}`-уудыг хувьсагчийн утгаар (мөр хэлбэрт)
// солино. Олдоогүй хувьсагч → хоосон мөр.
func substitute(s string, vars map[string]interface{}) string {
	return placeholderRe.ReplaceAllStringFunc(s, func(m string) string {
		name := strings.TrimSpace(m[2 : len(m)-1])
		if v, ok := vars[name]; ok {
			return valueToString(v)
		}
		return ""
	})
}

// valueToString нь JSON-оос задарсан утгыг (float64/bool/string/…) текст болгоно.
func valueToString(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		return fmt.Sprint(t)
	}
}

// condOps нь дэмжигдсэн харьцуулах операторууд (уртаар нь эрэмбэлсэн — ">="-ийг
// ">"-аас өмнө шалгана).
var condOps = []string{">=", "<=", "==", "!=", "=", ">", "<"}

// evalCondition нь gateway-ийн нэг салааны нөхцөлийг хувьсагчид дээр тооцно.
// "<var> <op> <literal>" хэлбэр; оператор олдохгүй бол `<var>`-ийг truthy гэж
// үзнэ.
func evalCondition(expr string, vars map[string]interface{}) bool {
	e := strings.TrimSpace(expr)
	if e == "" {
		return false
	}
	for _, op := range condOps {
		if idx := strings.Index(e, op); idx >= 0 {
			lhs := strings.TrimSpace(e[:idx])
			rhs := strings.TrimSpace(e[idx+len(op):])
			return compare(vars[lhs], op, rhs)
		}
	}
	// Операторгүй: ганц хувьсагчийн truthy шалгалт.
	return truthy(vars[strings.TrimSpace(e)])
}

func truthy(v interface{}) bool {
	switch t := v.(type) {
	case nil:
		return false
	case bool:
		return t
	case float64:
		return t != 0
	case string:
		return t != "" && t != "false" && t != "0"
	default:
		return v != nil
	}
}

// compare нь зүүн талын (хувьсагчийн) утга, оператор, баруун талын литералыг
// харьцуулна. >,>=,<,<= нь тоон; ==,= ба != нь тоон эсвэл текстээр.
func compare(lhs interface{}, op, rhsLiteral string) bool {
	switch op {
	case ">", ">=", "<", "<=":
		l, lok := toFloat(lhs)
		r, rok := parseFloatLiteral(rhsLiteral)
		if !lok || !rok {
			return false
		}
		switch op {
		case ">":
			return l > r
		case ">=":
			return l >= r
		case "<":
			return l < r
		case "<=":
			return l <= r
		}
	case "==", "=", "!=":
		eq := valuesEqual(lhs, rhsLiteral)
		if op == "!=" {
			return !eq
		}
		return eq
	}
	return false
}

// valuesEqual нь хувьсагчийн утга ба литералыг (тоо/bool/текст) тэнцүү эсэхийг
// шалгана.
func valuesEqual(lhs interface{}, rhsLiteral string) bool {
	// Тоо хэлбэрээр тулгаж үзэх.
	if l, lok := toFloat(lhs); lok {
		if r, rok := parseFloatLiteral(rhsLiteral); rok {
			return l == r
		}
	}
	// bool хэлбэр.
	if lb, ok := lhs.(bool); ok {
		return strconv.FormatBool(lb) == strings.ToLower(unquote(rhsLiteral))
	}
	// текст харьцуулалт.
	return valueToString(lhs) == unquote(rhsLiteral)
}

func toFloat(v interface{}) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case int:
		return float64(t), true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func parseFloatLiteral(s string) (float64, bool) {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f, err == nil
}

// unquote нь литералын хашилтыг (' эсвэл ") хасна.
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
