package domain

import (
	"testing"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/domain"
	"github.com/stretchr/testify/require"
)

func TestAllocatedStudent_IsWeeklySchedule(t *testing.T) {
	t.Parallel()

	t.Run("Weekly Schedule", func(t *testing.T) {
		as := &AllocatedStudent{
			ProductTypeSchedule: string(domain.Frequency),
		}
		isWeekly := as.IsWeeklySchedule()
		require.True(t, isWeekly)
	})

	t.Run("Not Weekly Schedule", func(t *testing.T) {
		as := &AllocatedStudent{
			ProductTypeSchedule: string(domain.OneTime),
		}
		isWeekly := as.IsWeeklySchedule()
		require.False(t, isWeekly)
	})
}

func TestAllocatedStudent_AllocationStatus(t *testing.T) {
	t.Parallel()
	t.Run("Partially Assigned", func(t *testing.T) {
		as := &AllocatedStudent{
			ProductTypeSchedule: string(domain.Frequency),
			AssignedSlot:        6,
			PurchasedSlot:       10,
		}
		status := as.AllocationStatus()
		require.Equal(t, string(PartiallyAssigned), status)
	})
	t.Run("None Assigned", func(t *testing.T) {
		as := &AllocatedStudent{
			ProductTypeSchedule: string(domain.Frequency),
			AssignedSlot:        0,
			PurchasedSlot:       10,
		}
		status := as.AllocationStatus()
		require.Equal(t, string(NoneAssigned), status)
	})
	t.Run("Over Assigned", func(t *testing.T) {
		as := &AllocatedStudent{
			ProductTypeSchedule: string(domain.Frequency),
			AssignedSlot:        11,
			PurchasedSlot:       10,
		}
		status := as.AllocationStatus()
		require.Equal(t, string(OverAssigned), status)
	})
	t.Run("Fully Assigned", func(t *testing.T) {
		as := &AllocatedStudent{
			ProductTypeSchedule: string(domain.Frequency),
			AssignedSlot:        10,
			PurchasedSlot:       10,
		}
		status := as.AllocationStatus()
		require.Equal(t, string(FullyAssigned), status)
	})
}
