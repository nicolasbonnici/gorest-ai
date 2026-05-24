package ai

type LocaleProvider interface {
	DefaultLocale() string
	TargetLocales() []string
}
