package utils

import (
	"fmt"
	"strings"

	"github.com/manabie-com/backend/developments/generate/internal/models"

	"github.com/manifoldco/promptui"
)

var Generator = promptui.Select{
	Label: "Choosing auto generator",
	Items: []string{"gen-bdd", "gen-bdd-test", "gen-all"},
	Templates: &promptui.SelectTemplates{
		Active:   fmt.Sprintf("%s {{ . | underline | green }}", promptui.IconSelect),
		Label:    fmt.Sprintf("%s {{. | blue }}: ", promptui.IconInitial),
		Selected: fmt.Sprintf("%s {{ . | white }}", promptui.IconGood+promptui.Styler(promptui.FGGreen)(" Choosing auto generator: ")),
	},
	HideSelected: true,
}
var Selector = promptui.Select{
	Label: "Service name",
	Items: []string{"eureka", "bob", "syllabus", "yasuo", "fatima", "shamir"},
	Templates: &promptui.SelectTemplates{
		Active:   fmt.Sprintf("%s {{ . | underline | green }}", promptui.IconSelect),
		Label:    fmt.Sprintf("%s {{. | blue }}: ", promptui.IconInitial),
		Selected: fmt.Sprintf("%s {{ . | white }}", promptui.IconGood+promptui.Styler(promptui.FGGreen)(" Service name: ")),
	},
}

var ServiceName = models.CliStep{
	Name: "service step",
	Type: models.SELECT,
	Select: promptui.Select{
		Label: "Service name",
		Items: []string{"eureka", "syllabus", "bob", "yasuo", "fatima", "shamir"},
		Templates: &promptui.SelectTemplates{
			Active:   fmt.Sprintf("%s {{ . | underline | green }}", promptui.IconSelect),
			Label:    fmt.Sprintf("%s {{ . | blue }}: ", promptui.IconInitial),
			Selected: fmt.Sprintf("%s {{ . | white }}", promptui.IconGood+promptui.Styler(promptui.FGGreen)(" Service name: ")),
		},
	},
}

var ChildServiceName = models.CliStep{
	Name: "Child service step",
	Type: models.PROMPT,
	Prompt: promptui.Prompt{
		Label: "Child service name(Ex: topic_modifier, assignment_reader)",
		Templates: &promptui.PromptTemplates{
			Success: fmt.Sprintf("%s {{ . | green }}%s ", promptui.IconGood, promptui.Styler(promptui.FGGreen)(":")),
			Valid:   fmt.Sprintf("{{ . | blue }}%s ", promptui.Styler(promptui.FGBlue)(":")),
			Invalid: fmt.Sprintf("{{ . | blue }}%s ", promptui.Styler(promptui.FGBlue)(":")),
		},
		Validate: func(str string) error {
			if !strings.Contains(str, "_modifier") && !strings.Contains(str, "_reader") {
				return fmt.Errorf("must have _modifier or _ reader")
			}
			return nil
		},
	},
}

var ProtobufChildServiceName = models.CliStep{
	Name: "step 2",
	Type: models.PROMPT,
	Prompt: promptui.Prompt{
		Label: "Protobuf service name(Ex: topic_modifier, assignment_reader)",
		Templates: &promptui.PromptTemplates{
			Success: fmt.Sprintf("%s {{ . | green }}%s ", promptui.IconGood, promptui.Styler(promptui.FGGreen)(":")),
			Valid:   fmt.Sprintf("{{ . | blue }}%s ", promptui.Styler(promptui.FGBlue)(":")),
			Invalid: fmt.Sprintf("{{ . | blue }}%s ", promptui.Styler(promptui.FGBlue)(":")),
		},
		Validate: func(str string) error {
			if !strings.Contains(str, "_modifier") && !strings.Contains(str, "_reader") {
				return fmt.Errorf("must have _modifier or _ reader")
			}
			return nil
		},
	},
}

var OptionMethod = []*models.CliStep{
	{
		Name: "step 1",
		Type: models.SELECT,
		Select: promptui.Select{
			Label: "Method name",
			Items: []string{"list", "retrieve"},
			Templates: &promptui.SelectTemplates{
				Active:   fmt.Sprintf("%s {{ . | underline | green }}", promptui.IconSelect),
				Label:    fmt.Sprintf("%s {{ . | blue }}: ", promptui.IconInitial),
				Selected: fmt.Sprintf("%s {{ . | white }}", promptui.IconGood+promptui.Styler(promptui.FGGreen)(" Method name: ")),
			},
		},
	},
	{
		Name: "step 2",
		Type: models.SELECT,
		Select: promptui.Select{
			Label: "Method name",
			Items: []string{"upsert", "delete", "update", "insert"},
			Templates: &promptui.SelectTemplates{
				Active:   fmt.Sprintf("%s {{ . | underline | green }}", promptui.IconSelect),
				Label:    fmt.Sprintf("%s {{ . | blue }}: ", promptui.IconInitial),
				Selected: fmt.Sprintf("%s {{ . | white }}", promptui.IconGood+promptui.Styler(promptui.FGGreen)(" Method name: ")),
			},
		},
	},
}

var MethodName = models.CliStep{
	Name: "Method step",
	Type: models.SELECT,
	Select: promptui.Select{
		Label: "Method name",
		Items: []string{"list", "retrieve", "upsert", "delete", "update", "insert"},
		Templates: &promptui.SelectTemplates{
			Active:   fmt.Sprintf("%s {{ . | underline | green }}", promptui.IconSelect),
			Label:    fmt.Sprintf("%s {{ . | blue }}: ", promptui.IconInitial),
			Selected: fmt.Sprintf("%s {{ . | white }}", promptui.IconGood+promptui.Styler(promptui.FGGreen)(" Method name: ")),
		},
	},
}

var OptionWithSubFolder = models.CliStep{
	Name: "Subfolder step",
	Type: models.SELECT,
	Select: promptui.Select{
		Label: "Subfolder",
		Items: []string{"existed folder", "add new folder"},
		Templates: &promptui.SelectTemplates{
			Active:   fmt.Sprintf("%s {{ . | underline | green }}", promptui.IconSelect),
			Label:    fmt.Sprintf("%s {{ . | blue }}: ", promptui.IconInitial),
			Selected: fmt.Sprintf("%s {{ . | white }}", promptui.IconGood+promptui.Styler(promptui.FGGreen)(" Subfolder: ")),
		},
	},
}

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
			return nil
		},
	},
}

var EntityName = models.CliStep{
	Name: "Feature",
	Type: models.PROMPT,
	Prompt: promptui.Prompt{
		Label: "Feature name",
		Templates: &promptui.PromptTemplates{
			Success: fmt.Sprintf("%s {{ . | green }}%s ", promptui.IconGood, promptui.Styler(promptui.FGGreen)(":")),
			Valid:   fmt.Sprintf("%s {{ . | blue }}%s ", promptui.IconGood, promptui.Styler(promptui.FGBlue)(":")),
		},
		Validate: func(str string) error {
			if str == "" {
				return fmt.Errorf("can't be empty")
			}
			return nil
		},
	},
}
