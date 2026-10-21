package proxy

import (
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// Country infers the ISO 3166-1 alpha-2 region of a node from its name, or
// "" when the name does not tell. Detection order:
//
//  1. a flag emoji in the name;
//  2. a region name in English, Simplified or Traditional Chinese or
//     Japanese, longest first, so "Hong Kong" wins over "HK";
//  3. the first two-letter ASCII token that is a region code, so "us-la-02"
//     is the United States rather than Laos.
func Country(name string) string {
	if code := embeddedCode(name); code != "" {
		return code
	}
	tables := keywords()
	for _, entry := range tables.long {
		if entry.matches(name) {
			return entry.code
		}
	}
	for _, token := range strings.FieldsFunc(name, func(r rune) bool { return r > unicode.MaxASCII || !unicode.IsLetter(r) }) {
		if len(token) == 2 {
			if code, ok := tables.iso[strings.ToUpper(token)]; ok {
				return code
			}
		}
	}
	return ""
}

// Flag returns the flag emoji of a region code.
func Flag(code string) string {
	if len(code) != 2 {
		return ""
	}
	var b strings.Builder
	for _, r := range strings.ToUpper(code) {
		if r < 'A' || r > 'Z' {
			return ""
		}
		b.WriteRune(r - 'A' + 0x1F1E6)
	}
	return b.String()
}

// DisplayName drops flag emoji from a name, for surfaces that draw the
// flag separately. The original name stays the node's identity.
func DisplayName(name string) string {
	var b strings.Builder
	removed := false
	runes := []rune(name)
	for i := 0; i < len(runes); i++ {
		if i+1 < len(runes) && isIndicator(runes[i]) && isIndicator(runes[i+1]) {
			removed = true
			i++
			continue
		}
		b.WriteRune(runes[i])
	}
	if !removed {
		return name
	}
	stripped := strings.Join(strings.Fields(b.String()), " ")
	stripped = strings.Trim(stripped, " -_·•/|｜")
	if stripped == "" {
		return name
	}
	return stripped
}

func isIndicator(r rune) bool { return r >= 0x1F1E6 && r <= 0x1F1FF }

func embeddedCode(name string) string {
	runes := []rune(name)
	for i := 0; i+1 < len(runes); i++ {
		if isIndicator(runes[i]) && isIndicator(runes[i+1]) {
			return string([]rune{runes[i] - 0x1F1E6 + 'A', runes[i+1] - 0x1F1E6 + 'A'})
		}
	}
	return ""
}

type keyword struct {
	text  string
	code  string
	ascii *regexp.Regexp
}

func (k keyword) matches(name string) bool {
	if k.ascii != nil {
		return k.ascii.MatchString(name)
	}
	return strings.Contains(strings.ToLower(name), strings.ToLower(k.text))
}

type keywordTables struct {
	long []keyword
	iso  map[string]string
}

// aliases are names subscriptions use that the CLDR does not give, such
// as the short Chinese names of Hong Kong and Taiwan.
var aliases = map[string]string{
	"USA": "US", "UK": "GB", "Britain": "GB", "England": "GB",
	"香港": "HK", "台湾": "TW", "台灣": "TW", "臺灣": "TW",
	"狮城": "SG", "獅城": "SG", "美国": "US", "日本": "JP", "韩国": "KR",
	"Hong Kong": "HK", "Macau": "MO", "Macao": "MO", "Taiwan": "TW", "澳门": "MO",
}

var keywords = sync.OnceValue(func() keywordTables {
	tables := keywordTables{iso: map[string]string{}}
	seen := map[string]bool{}
	add := func(text, code string) {
		text = strings.TrimSpace(text)
		isASCII := true
		for _, r := range text {
			if r > unicode.MaxASCII {
				isASCII = false
				break
			}
		}
		if text == "" || (isASCII && len(text) == 2) {
			return
		}
		key := strings.ToLower(text) + "|" + code
		if seen[key] {
			return
		}
		seen[key] = true
		entry := keyword{text: text, code: code}
		if isASCII {
			entry.ascii = regexp.MustCompile(`(?i)(^|[^A-Za-z0-9])` + regexp.QuoteMeta(text) + `($|[^A-Za-z0-9])`)
		}
		tables.long = append(tables.long, entry)
	}
	namers := []display.Namer{
		display.English.Regions(),
		display.Regions(language.SimplifiedChinese),
		display.Regions(language.TraditionalChinese),
		display.Regions(language.Japanese),
	}
	for a := 'A'; a <= 'Z'; a++ {
		for b := 'A'; b <= 'Z'; b++ {
			code := string([]rune{a, b})
			region, err := language.ParseRegion(code)
			if err != nil || !region.IsCountry() || region.String() != code {
				continue
			}
			tables.iso[code] = code
			for _, namer := range namers {
				if name := namer.Name(region); name != "" && !strings.EqualFold(name, code) {
					add(name, code)
				}
			}
		}
	}
	for text, code := range aliases {
		add(text, code)
		if len(text) == 2 {
			tables.iso[strings.ToUpper(text)] = code
		}
	}
	sort.SliceStable(tables.long, func(i, j int) bool {
		return len([]rune(tables.long[i].text)) > len([]rune(tables.long[j].text))
	})
	return tables
})
