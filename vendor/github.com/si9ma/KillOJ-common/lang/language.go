package lang

import (
	"os"
	"strings"

	"github.com/si9ma/KillOJ-common/constants"
	"golang.org/x/text/language"
)

var supportLangs = []language.Tag{
	language.English, // english
	language.Chinese, // chinese
}

var matcher = language.NewMatcher(supportLangs)

func GetLangFromEnv() language.Tag {
	var envLangs []language.Tag

	tagStr := constants.DefaultLang
	if val := os.Getenv(constants.EnvLang); val != "" {
		tagStr = val
	}

	tags := strings.Split(tagStr, ",")
	for _, tag := range tags {
		lang := language.Make(strings.TrimSpace(tag))
		envLangs = append(envLangs, lang)
	}

	tag, _, _ := matcher.Match(envLangs...)
	return tag
}
