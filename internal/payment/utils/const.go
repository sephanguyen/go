package utils

var NumberNames = [...]string{
	"first",
	"second",
	"third",
	"fourth",
	"fifth",
	"sixth",
	"seventh",
	"eighth",
	"ninth",
	"tenth",
	"eleventh",
	"twelveth",
	"thirdteenth",
	"fourteenth",
}

type LanguageCode string

const (
	LayoutISO        string = "2006-01-02"
	pkgHeaderKey     string = "pkg"
	tokenHeaderKey   string = "token"
	versionHeaderKey string = "version"

	EnCode LanguageCode = "en"
	JpCode LanguageCode = "jp"

	EngLOANotificationContentTemp string = "LOA Status of %s %s in %s is expiring."
	JpLOANotificationContentTemp  string = "%s の %s %s の休塾終了予定日が近づいています"
)
