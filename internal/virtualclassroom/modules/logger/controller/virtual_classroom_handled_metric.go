package controller

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/metrics"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/infrastructure/repo"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type MetricLabel string

const (
	MetricLabelUpdateRoomState MetricLabel = "update_room_state"
	MetricLabelGetRoomState    MetricLabel = "get_room_state"
)

type VirtualClassroomHandledMetric struct {
	HandleRoomState  *prometheus.HistogramVec
	RealLiveTime     *prometheus.HistogramVec
	TotalAttendees   *prometheus.HistogramVec
	TotalActiveRooms *prometheus.GaugeVec
}

func RegisterVirtualClassroomHandledMetric(collector metrics.MetricCollector) *VirtualClassroomHandledMetric {
	VCrMetric := &VirtualClassroomHandledMetric{
		HandleRoomState: collector.RegisterHistogram(metrics.MetricOpt{
			Name:       "backend_virtual_classroom_handled_total",
			Help:       "Total number of calling completed of every room, include update and get room state action.",
			LabelNames: []string{"action"},
		}, []float64{10, 20, 50, 100, 200, 500, 1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000}), // TODO: tracking metric and adjust these params after
		RealLiveTime: collector.RegisterHistogram(metrics.MetricOpt{
			Name: "backend_virtual_classroom_total_real_live_time_minutes",
			Help: "Total real live time of every room",
		}, prometheus.LinearBuckets(10, 10, 11)), // TODO: tracking metric and adjust these params after
		TotalAttendees: collector.RegisterHistogram(metrics.MetricOpt{
			Name: "backend_virtual_classroom_attendees_total",
			Help: "Total number of attendees who really joined room",
		}, prometheus.LinearBuckets(5, 5, 10)), // TODO: tracking metric and adjust these params after
		TotalActiveRooms: collector.RegisterGauge(metrics.MetricOpt{
			Name: "backend_virtual_classroom_active_rooms_total",
			Help: "Total active rooms",
		}),
	}

	return VCrMetric
}

func (v *VirtualClassroomHandledMetric) ObserveWhenEndRoom(ctx context.Context, log *repo.VirtualClassRoomLogDTO) {
	defer func() {
		if err := recover(); err != nil {
			logger := ctxzap.Extract(ctx)
			logger.Error(
				"VirtualClassroomHandledMetric.ObserveWhenEndRoom: panic: ",
				zap.Error(err.(error)),
			)
		}
	}()
	v.HandleRoomState.WithLabelValues(string(MetricLabelUpdateRoomState)).Observe(float64(log.TotalTimesUpdatingRoomState.Int))
	v.HandleRoomState.WithLabelValues(string(MetricLabelGetRoomState)).Observe(float64(log.TotalTimesGettingRoomState.Int))
	v.TotalAttendees.WithLabelValues().Observe(float64(len(log.AttendeeIDs.Elements)))
	v.RealLiveTime.WithLabelValues().Observe(log.UpdatedAt.Time.Sub(log.CreatedAt.Time).Minutes())
	v.TotalActiveRooms.WithLabelValues().Dec()
}

func (v *VirtualClassroomHandledMetric) ObserveWhenStartRoom(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			logger := ctxzap.Extract(ctx)
			logger.Error(
				"VirtualClassroomHandledMetric.ObserveWhenStartRoom: panic: ",
				zap.Error(err.(error)),
			)
		}
	}()
	v.TotalActiveRooms.WithLabelValues().Inc()
}
