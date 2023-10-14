package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/manabie-com/backend/developments/generate/configs"
	"github.com/manabie-com/backend/developments/generate/internal"
	"github.com/manabie-com/backend/developments/generate/internal/models"
	"github.com/manabie-com/backend/developments/generate/internal/utils"
)

var s configs.Config

func main() {
	_, genStr, err := utils.Generator.Run()
	if err != nil {
		panic(err)
	}
	switch genStr {
	case "gen-bdd":
		_, s.Service, _ = utils.Selector.Run()
		internal.RunGenBdd(&s)

	case "gen-bdd-test":
		runGenBddV2()

	case "gen-all":
		runGenAll()
	}
}

func runGenAll() {
	var err error
	_, s.Service, err = utils.ServiceName.Select.Run()
	if err != nil {
		panic(err)
	}
	if s.Service == "syllabus" {
		runGenAllSyllabus(s.Service)
		return
	}
	for _, step := range stepList {
		switch step.Type {
		case models.PROMPT:
			step.Val, err = step.Prompt.Run()
			if err != nil {
				panic(err)
			}
			if strings.Contains(step.Prompt.Label.(string), "Protobuf service name") {
				runSelectMethod(step.Val)
			}
		case models.SELECT:
			_, step.Val, err = step.Select.Run()
			if err != nil {
				panic(err)
			}
		}
	}

	s.ChildService = stepList[0].Val
	s.Entity = stepList[1].Val
	internal.RunGenAll(&s)
}

func runGenBddV2() {
	var err error
	var optionFolder string

	baseDir, _ := os.Getwd()

	_, s.Service, err = utils.ServiceName.Select.Run()
	if err != nil {
		panic(err)
	}
	_, optionFolder, err = utils.OptionWithSubFolder.Select.Run()
	if err != nil {
		panic(err)
	}

	if optionFolder == "existed folder" {
		existOption := utils.RunSelectDir(fmt.Sprintf("%s/features/%s", baseDir, s.Service))
		_, s.FolderEntity, err = existOption.Select.Run()
		if err != nil {
			panic(err)
		}
		for _, step := range stepGenBddV3List {
			switch step.Type {
			case models.PROMPT:
				step.Val, err = step.Prompt.Run()
				if err != nil {
					panic(err)
				}
			case models.SELECT:
				_, step.Val, err = step.Select.Run()
				if err != nil {
					panic(err)
				}
			}
		}

		s.Method = stepGenBddV2List[0].Val
		s.Entity = stepGenBddV2List[1].Val
		internal.RunGenBddV2(&s)
		return
	}

	existOption := utils.RunPromptDir(fmt.Sprintf("%s/features/%s", baseDir, s.Service))
	s.FolderEntity, err = existOption.Prompt.Run()
	if err != nil {
		panic(err)
	}
	for _, step := range stepGenBddV2List {
		switch step.Type {
		case models.PROMPT:
			step.Val, err = step.Prompt.Run()
			if err != nil {
				panic(err)
			}
		case models.SELECT:
			_, step.Val, err = step.Select.Run()
			if err != nil {
				panic(err)
			}
		}
	}

	s.Method = stepGenBddV2List[0].Val
	s.Entity = stepGenBddV2List[1].Val
	internal.RunGenBddV2(&s)
}

func runGenAllSyllabus(service string) {
	var err error
	for _, step := range stepsListSyllabus {
		switch step.Type {
		case models.PROMPT:
			step.Val, err = step.Prompt.Run()
			if err != nil {
				panic(err)
			}
			if strings.Contains(step.Prompt.Label.(string), "Protobuf service name") {
				runSelectMethod(step.Val)
			}
		case models.SELECT:
			_, step.Val, err = step.Select.Run()
			if err != nil {
				panic(err)
			}
		}
	}

	s.Service = service
	s.ChildService = stepsListSyllabus[0].Val
	s.FolderEntity = stepsListSyllabus[1].Val
	s.Entity = stepsListSyllabus[2].Val
	internal.RunGenAll(&s)
}

func runSelectMethod(childService string) {
	var err error
	if strings.Contains(childService, "_reader") {
		_, s.Method, err = optionMethod[0].Select.Run()
		if err != nil {
			panic(err)
		}
	}

	if strings.Contains(childService, "_modifier") {
		_, s.Method, err = optionMethod[1].Select.Run()
		if err != nil {
			panic(err)
		}
	}
}

var stepList = []*models.CliStep{
	&utils.ProtobufChildServiceName,
	&utils.EntityName,
}

var optionMethod = utils.OptionMethod

var stepGenBddV2List = []*models.CliStep{
	&utils.MethodName,
	&utils.EntityName,
}

var stepGenBddV3List = []*models.CliStep{
	&utils.MethodName,
	&utils.EntityName,
}

var stepsListSyllabus = []*models.CliStep{
	&utils.ProtobufChildServiceName,
	&utils.FolderEntityName,
	&utils.EntityName,
}
