package constants

import cpb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2/common"

type LearningMaterialType string

const (
	LearningObjective LearningMaterialType = "LEARNING_MATERIAL_LEARNING_OBJECTIVE"
	ExamLO            LearningMaterialType = "LEARNING_MATERIAL_EXAM_LO"
	FlashCard         LearningMaterialType = "LEARNING_MATERIAL_FLASH_CARD"
	GeneralAssignment LearningMaterialType = "LEARNING_MATERIAL_GENERAL_ASSIGNMENT"
	TaskAssignment    LearningMaterialType = "LEARNING_MATERIAL_TASK_ASSIGNMENT"
)

func (lmt LearningMaterialType) GetProtobufType() cpb.LearningMaterialType {
	switch lmt {
	case LearningObjective:
		return cpb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE
	case FlashCard:
		return cpb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD
	case GeneralAssignment:
		return cpb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT
	case TaskAssignment:
		return cpb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT
	case ExamLO:
		return cpb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO

	default:
		return cpb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO
	}
}

func GetLearningMaterialFromProtobuf(lmt cpb.LearningMaterialType) LearningMaterialType {
	switch lmt {
	case cpb.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE:
		return LearningObjective
	case cpb.LearningMaterialType_LEARNING_MATERIAL_FLASH_CARD:
		return FlashCard
	case cpb.LearningMaterialType_LEARNING_MATERIAL_GENERAL_ASSIGNMENT:
		return GeneralAssignment
	case cpb.LearningMaterialType_LEARNING_MATERIAL_TASK_ASSIGNMENT:
		return TaskAssignment
	case cpb.LearningMaterialType_LEARNING_MATERIAL_EXAM_LO:
		return ExamLO
	default:
		return ""
	}
}
