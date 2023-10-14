package services

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListChapters(t *testing.T) {
	t.Parallel()
	chapterRepo := &mock_repositories.MockChapterRepo{}
	svc := &ChapterReaderService{
		ChapterRepo: chapterRepo,
	}

	time := &timestamppb.Timestamp{
		Seconds: timestamppb.Now().Seconds,
	}

	testCases := []TestCase{
		{
			name:         "empty chapters",
			expectedResp: &pb.ListChaptersResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				chapterRepo.On("ListChapters", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)

			},
			req: &pb.ListChaptersRequest{},
		},
		{
			name: "success in query first page",
			expectedResp: &pb.ListChaptersResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetCombined{
						OffsetCombined: &cpb.Paging_Combined{
							OffsetString: "id2",
						},
					},
				},
				Items: []*cpb.Chapter{
					{Info: &cpb.ContentBasicInfo{
						Id:        "id1",
						Name:      "sid",
						UpdatedAt: time,
						CreatedAt: time,
					}},
					{Info: &cpb.ContentBasicInfo{
						Id:        "id2",
						Name:      "sid",
						UpdatedAt: time,
						CreatedAt: time,
					}},
				},
			},
			setup: func(ctx context.Context) {
				items := []*entities.Chapter{
					{
						ID:   database.Text("id1"),
						Name: database.Text("sid"),
						UpdatedAt: pgtype.Timestamptz{
							Time: time.AsTime(),
						},
						CreatedAt: pgtype.Timestamptz{
							Time: time.AsTime(),
						},
					},
					{
						ID:   database.Text("id2"),
						Name: database.Text("sid"),
						UpdatedAt: pgtype.Timestamptz{
							Time: time.AsTime(),
						},
						CreatedAt: pgtype.Timestamptz{
							Time: time.AsTime(),
						},
					},
				}
				chapterRepo.On("ListChapters", mock.Anything, mock.Anything, &repositories.ListChaptersArgs{
					ChapterIDs: database.TextArray([]string{"id1", "id2"}),
					Limit:      10,
					Offset:     pgtype.Int4{Status: pgtype.Null},
					ChapterID:  pgtype.Text{Status: pgtype.Null},
				}).Once().Return(items, nil)
			},
			req: &pb.ListChaptersRequest{
				Filter: &cpb.CommonFilter{
					Ids: []string{"id1", "id2"},
				},
				Paging: &cpb.Paging{
					Limit: 10,
				},
			},
		},
		{
			name: "success in query next page",
			expectedResp: &pb.ListChaptersResponse{
				NextPage: &cpb.Paging{
					Limit: 5,
					Offset: &cpb.Paging_OffsetCombined{
						OffsetCombined: &cpb.Paging_Combined{
							OffsetString: "id2",
						},
					},
				},
				Items: []*cpb.Chapter{
					{Info: &cpb.ContentBasicInfo{
						Id:        "id1",
						Name:      "sid",
						UpdatedAt: time,
						CreatedAt: time,
					}},
					{Info: &cpb.ContentBasicInfo{
						Id:        "id2",
						Name:      "sid",
						UpdatedAt: time,
						CreatedAt: time,
					}},
				},
			},
			setup: func(ctx context.Context) {
				items := []*entities.Chapter{
					{
						ID:   database.Text("id1"),
						Name: database.Text("sid"),
						UpdatedAt: pgtype.Timestamptz{
							Time: time.AsTime(),
						},
						CreatedAt: pgtype.Timestamptz{
							Time: time.AsTime(),
						},
					},
					{
						ID:   database.Text("id2"),
						Name: database.Text("sid"),
						UpdatedAt: pgtype.Timestamptz{
							Time: time.AsTime(),
						},
						CreatedAt: pgtype.Timestamptz{
							Time: time.AsTime(),
						},
					},
				}
				chapterRepo.On("ListChapters", mock.Anything, mock.Anything, &repositories.ListChaptersArgs{
					ChapterIDs: database.TextArray([]string{"id1", "id2"}),
					Limit:      5,
					Offset:     pgtype.Int4{Int: 2, Status: pgtype.Present},
					ChapterID:  pgtype.Text{String: "id2", Status: pgtype.Present},
				}).Once().Return(items, nil)
			},
			req: &pb.ListChaptersRequest{
				Filter: &cpb.CommonFilter{
					Ids: []string{"id1", "id2"},
				},
				Paging: &cpb.Paging{
					Limit: 5,
					Offset: &cpb.Paging_OffsetCombined{
						OffsetCombined: &cpb.Paging_Combined{
							OffsetInteger: 2,
							OffsetString:  "id2",
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			request := testCase.req.(*pb.ListChaptersRequest)
			expectResp := testCase.expectedResp.(*pb.ListChaptersResponse)
			resp, err := svc.ListChapters(context.Background(), request)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, expectResp, resp)
		})
	}
}
