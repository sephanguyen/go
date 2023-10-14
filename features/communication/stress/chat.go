package stress

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/eibanam/communication/helper"
	"github.com/manabie-com/backend/internal/golibs/try"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"golang.org/x/sync/errgroup"
)

// n conversations
// k user per conversations
type StressChat struct {
	pusher    *push.Pusher
	semaphore semaphore

	msgDelivery prometheus.Histogram
	// deliveredMsg    prometheus.Counter
	// totalMsg        prometheus.Counter
	commonSuite     *common.Suite
	connections     *common.Connections
	helper          *helper.CommunicationHelper
	DefaultLocation string
	TotalConv       int
	MemberPerConv   int
	SpamInterval    time.Duration
}

// Enable push gateway in local to run this function
func (s *StressChat) initPusher(ctx context.Context) {
	// s.deliveredMsg = prometheus.NewCounter(prometheus.CounterOpts{
	// 	Name: "manabie_app_tom_stress_test_delivered_msg",
	// })
	// 	s.totalMsg = prometheus.NewCounter(prometheus.CounterOpts{
	// 		Name: "manabie_app_tom_stress_test_total_msg",
	// 	})
	s.msgDelivery = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "manabie_app_tom_stress_test_delivery_latency",
	})
	pusher := push.New("http://prometheus-pushgateway.monitoring.svc.cluster.local:9091", "tom_stress_test").Collector(s.msgDelivery)
	// Collector(s.totalMsg).
	// Collector(s.deliveredMsg)

	s.pusher = pusher
	go s.pushByInterval(ctx)
}

func (s *StressChat) pushByInterval(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			err := s.pusher.Push()
			if err != nil {
				fmt.Printf("error %s\n", err)
			}
			return
		case <-ticker.C:
			err := s.pusher.Add()
			if err != nil {
				fmt.Printf("error %s\n", err)
			}
		}
	}
}

func NewStressChat(suite *common.Suite, conn *common.Connections, helper *helper.CommunicationHelper,
	locID string,
	totalConv int,
	memberConv int,
) StressChat {
	return StressChat{
		semaphore:       make(chan struct{}, 3),
		commonSuite:     suite,
		connections:     conn,
		helper:          helper,
		DefaultLocation: locID,
		TotalConv:       totalConv,
		MemberPerConv:   memberConv,
	}
}

func (s *StressChat) Run(ctx context.Context) (context.Context, error) {
	s.initPusher(ctx)
	// s.TotalConv = totalConv
	// s.MemberPerConv = memPerConv
	s.SpamInterval = time.Second
	errgr := &errgroup.Group{}
	for i := 0; i < s.TotalConv; i++ {
		errgr.Go(func() error {
			err := s.SetupPerConversation(ctx)
			if err != nil {
				fmt.Printf("SetupPerConversation %s\n", err)
			}
			return err
		})
	}
	err := errgr.Wait()
	return ctx, err
}

func (s *StressChat) findChat(ctx context.Context, studentID string) (convID string, err error) {
	conv := pgtype.Text{}
	err = try.Do(func(attempt int) (bool, error) {
		err = s.connections.TomDB.QueryRow(ctx, "select conversation_id from conversation_students where student_id=$1", studentID).Scan(&conv)
		if err != nil {
			time.Sleep(3 * time.Second)
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return "", err
	}
	return conv.String, nil
}

type semaphore chan struct{}

func (s semaphore) acq() {
	s <- struct{}{}
}

func (s semaphore) rel() {
	<-s
}

func (s *StressChat) SetupPerConversation(ctx context.Context) error {
	s.semaphore.acq()
	stu, err := s.commonSuite.CreateStudent(ctx, []string{s.DefaultLocation}, nil)
	if err != nil {
		s.semaphore.rel()
		return err
	}

	tok := make([]string, 0, s.MemberPerConv)
	ids := make([]string, 0, s.MemberPerConv)
	stuTok, err := s.commonSuite.GenerateExchangeTokenCtx(ctx, stu.UserProfile.UserId, cpb.UserGroup_USER_GROUP_STUDENT.String())
	if err != nil {
		s.semaphore.rel()
		return err
	}
	s.semaphore.rel()
	convID, err := s.findChat(ctx, stu.UserProfile.UserId)
	if err != nil {
		return err
	}
	tok = append(tok, stuTok)
	ids = append(ids, stu.UserProfile.UserId)
	for i := 0; i < s.MemberPerConv-1; i++ {
		prof, token, err := s.commonSuite.CreateTeacher(ctx)
		if err != nil {
			return err
		}
		tok = append(tok, token)
		ids = append(ids, prof.StaffId)
		err = s.helper.JoinSupportChatGroup(ctx, token, convID)
		if err != nil {
			return err
		}
	}

	s.spamRoutine(ctx, convID, tok, ids)
	return nil
	// return s.spamPingRoutine(ctx, tok, ids)
}

func (s *StressChat) spamPingRoutine(ctx context.Context, toks []string, ids []string) error {
	for idx, tok := range toks {
		id := ids[idx]
		// ping until ctx is canceled
		err := s.helper.ConnectChatStreamWithHashAndPings(ctx, tok, id, 1)
		if err != nil {
			return err
		}
	}
	<-ctx.Done()
	return nil
}

// TODO: research how to track missed msg, because msg can be redelivered multiple times
func (s *StressChat) spamRoutine(ctx context.Context, convID string, toks []string, ids []string) {
	chatWg := &sync.WaitGroup{}
	// 1 token send msg, n -1 token receive msg
	chatWg.Add(len(toks))
	stuTok := toks[0]

	teacherTok := toks[1:]

	for idx := range teacherTok {
		tok := teacherTok[idx]
		id := ids[idx+1]
		stream, err := s.helper.ConnectChatStreamWithHash(ctx, tok, id)
		if err != nil {
			return
		}
		go func() {
			defer chatWg.Done()
			msgchan := s.helper.MsgChanFromStream(ctx, stream)
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-msgchan:
					tstr := strings.Split(msg.Content, " ")[1]
					sentAt, err := time.Parse(time.RFC3339, tstr)
					if err != nil {
						// never happen tho
						panic(fmt.Sprintf("time.Parse %s", err))
					}

					latency := time.Since(sentAt)
					s.msgDelivery.Observe(latency.Seconds())
				}
			}
		}()
	}
	go func() {
		defer chatWg.Done()
		ticker := time.NewTicker(s.SpamInterval)
		defer ticker.Stop()
		errCount := 0
		counter := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				counter++
				content := fmt.Sprintf(msgTemplate, counter, time.Now().Format(time.RFC3339))
				err := s.helper.SendTextMsgToConversation(ctx, stuTok, content, convID)
				if err != nil {
					errCount++
					if errCount > 10 {
						return
					}
				}
			}
		}
	}()
	chatWg.Wait()
}

var (
	msgTemplate = "%d %s"
)
