package utils

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/manabie-com/backend/developments/generate/internal/models"

	"github.com/manifoldco/promptui"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// some snippet formatting regexps
var (
	snippetExprCleanup = regexp.MustCompile(`([\/\[\]\(\)\\^\\$\.\|\?\*\+\'])`)
	snippetExprQuoted  = regexp.MustCompile(`(\W|^)\"(?:[^\"]*)\"(\W|$)`)
	snippetMethodName  = regexp.MustCompile(`[^a-zA-Z\_\ ]`)
	snippetNumbers     = regexp.MustCompile(`(\d+)`)
)

func GetArgs(expr string) (ret string) {
	var (
		args      []string
		pos       int
		breakLoop bool
	)

	for !breakLoop {
		part := expr[pos:]
		ipos := strings.Index(part, "(\\d+)")
		spos := strings.Index(part, "\"([^\"]*)\"")

		switch {
		case spos == -1 && ipos == -1:
			breakLoop = true
		case spos == -1:
			pos += ipos + len("(\\d+)")
			args = append(args, reflect.Int.String())
		case ipos == -1:
			pos += spos + len("\"([^\"]*)\"")
			args = append(args, reflect.String.String())
		case ipos < spos:
			pos += ipos + len("(\\d+)")
			args = append(args, reflect.Int.String())
		case spos < ipos:
			pos += spos + len("\"([^\"]*)\"")
			args = append(args, reflect.String.String())
		}
	}

	var last string

	for i, arg := range args {
		if last == "" || last == arg {
			ret += fmt.Sprintf("arg%d, ", i+1)
		} else {
			ret = strings.TrimRight(ret, ", ") + fmt.Sprintf(" %s, arg%d, ", last, i+1)
		}

		last = arg
	}

	return strings.TrimSpace(strings.TrimRight(ret, ", ") + " " + last)
}

func GetExprAndFuncName(step string) (string, string) {
	expr := snippetExprCleanup.ReplaceAllString(step, "\\$1")
	expr = snippetNumbers.ReplaceAllString(expr, "(\\d+)")
	expr = snippetExprQuoted.ReplaceAllString(expr, "$1\"([^\"]*)\"$2")
	expr = "^" + strings.TrimSpace(expr) + "$"

	name := snippetNumbers.ReplaceAllString(step, " ")
	name = snippetExprQuoted.ReplaceAllString(name, " ")
	name = strings.TrimSpace(snippetMethodName.ReplaceAllString(name, ""))
	nameSplitArr := strings.Split(name, " ")
	words := make([]string, 0, len(nameSplitArr))
	for i, w := range nameSplitArr {
		switch {
		case i != 0:
			caser := cases.Title(language.English)
			w = caser.String(w)

		case len(w) > 0:
			w = string(unicode.ToLower(rune(w[0]))) + w[1:]
		}
		words = append(words, w)
	}
	name = strings.Join(words, "")

	return expr, name
}

func GetDir(filePath string) []string {
	files, err := ioutil.ReadDir(filePath)
	folder := make([]string, 0, len(files))

	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if f.IsDir() && f.Name() != "common" && f.Name() != "utils" {
			check, _ := IsEmpty(filePath + "/" + f.Name())
			if !check {
				folder = append(folder, f.Name())
			}
		}
	}
	return folder
}

func RunSelectDir(filePath string) models.CliStep {
	dirs := GetDir(filePath)
	var FolderEntityName = models.CliStep{
		Name: "Subfolder step",
		Type: models.SELECT,
		Select: promptui.Select{
			Label: "Subfolder",
			Items: dirs,
			Templates: &promptui.SelectTemplates{
				Active:   fmt.Sprintf("%s {{ . | underline | green }}", promptui.IconSelect),
				Label:    fmt.Sprintf("%s {{ . | blue }}: ", promptui.IconInitial),
				Selected: fmt.Sprintf("%s {{ . | white }}", promptui.IconGood+promptui.Styler(promptui.FGGreen)(" Subfolder: ")),
			},
		},
	}
	return FolderEntityName
}

func RunPromptDir(filePath string) models.CliStep {
	dirs := GetDir(filePath)
	var FolderEntityName = models.CliStep{
		Name: "Folder entity step",
		Type: models.PROMPT,
		Prompt: promptui.Prompt{
			Label: "Folder name: ",
			Templates: &promptui.PromptTemplates{
				Success: fmt.Sprintf("%s {{ . | green }}%s ", promptui.IconGood, promptui.Styler(promptui.FGGreen)(":")),
				Valid:   fmt.Sprintf("{{ . | blue }}%s ", promptui.Styler(promptui.FGBlue)(":")),
				Invalid: fmt.Sprintf("{{ . | blue }}%s ", promptui.Styler(promptui.FGBlue)(":")),
			},
			Validate: func(str string) error {
				if str == "" {
					return fmt.Errorf("can't be empty")
				}
				for _, dir := range dirs {
					if str == dir {
						return fmt.Errorf("folder is exist")
					}
				}
				return nil
			},
		},
	}
	return FolderEntityName
}
