package clients

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jurwaCST/Jurwa/functions/digest-dispatcher/dispatcher"
	analyticsv1 "github.com/jurwaCST/Jurwa/proto/gen/analytics/v1"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	departmentdirectoryv1 "github.com/jurwaCST/Jurwa/proto/gen/department-directory/v1"
	notificationv1 "github.com/jurwaCST/Jurwa/proto/gen/notification/v1"
)

// Fakes for the generated gRPC clients. The embedded interface covers the
// methods the dispatcher never calls.

type fakeAnalytics struct {
	analyticsv1.AnalyticsServiceClient
	resp *analyticsv1.GetCategoryBreakdownResponse
	err  error
}

func (f fakeAnalytics) GetCategoryBreakdown(context.Context, *analyticsv1.GetCategoryBreakdownRequest, ...grpc.CallOption) (*analyticsv1.GetCategoryBreakdownResponse, error) {
	return f.resp, f.err
}

type fakeDirectory struct {
	departmentdirectoryv1.DepartmentDirectoryServiceClient
	got  string
	resp *departmentdirectoryv1.GetStaffResponse
	err  error
}

func (f *fakeDirectory) GetStaff(_ context.Context, in *departmentdirectoryv1.GetStaffRequest, _ ...grpc.CallOption) (*departmentdirectoryv1.GetStaffResponse, error) {
	f.got = in.GetPositionCode()
	return f.resp, f.err
}

type fakeNotification struct {
	notificationv1.NotificationServiceClient
	got *notificationv1.SendNotificationRequest
	err error
}

func (f *fakeNotification) SendNotification(_ context.Context, in *notificationv1.SendNotificationRequest, _ ...grpc.CallOption) (*notificationv1.SendNotificationResponse, error) {
	f.got = in
	return &notificationv1.SendNotificationResponse{}, f.err
}

func TestAnalyticsCategoryCounts(t *testing.T) {
	fake := fakeAnalytics{resp: &analyticsv1.GetCategoryBreakdownResponse{Categories: []*analyticsv1.CategoryCounts{
		{Category: commonv1.Category_HOSTEL, Counts: &analyticsv1.StatusCounts{Total: 5, Open: 2, Assigned: 1, InProgress: 1, Resolved: 1, Escalated: 1}},
		{Category: commonv1.Category_OTHERS, Counts: &analyticsv1.StatusCounts{}},
	}}}
	got, err := NewAnalytics(fake).CategoryCounts(context.Background())
	require.NoError(t, err)
	assert.Equal(t, dispatcher.Counts{Total: 5, Open: 2, Assigned: 1, InProgress: 1, Resolved: 1, Escalated: 1}, got[dispatcher.Hostel])
	assert.Equal(t, dispatcher.Counts{}, got[dispatcher.Others])

	_, err = NewAnalytics(fakeAnalytics{err: status.Error(codes.Unavailable, "down")}).CategoryCounts(context.Background())
	assert.Equal(t, codes.Unavailable, status.Code(err), "the gRPC code is kept")
}

func TestDirectoryHolderOf(t *testing.T) {
	fake := &fakeDirectory{resp: &departmentdirectoryv1.GetStaffResponse{Position: &departmentdirectoryv1.Position{HolderUserId: "user-dsa"}}}
	got, err := NewDirectory(fake).HolderOf(context.Background(), "DSA")
	require.NoError(t, err)
	assert.Equal(t, "user-dsa", got)
	assert.Equal(t, "DSA", fake.got)

	_, err = NewDirectory(&fakeDirectory{err: status.Error(codes.NotFound, "no DSA")}).HolderOf(context.Background(), "DSA")
	assert.Error(t, err)
}

func TestNotificationSend(t *testing.T) {
	fake := &fakeNotification{}
	err := NewNotification(fake).Send(context.Background(), "user-dsa", "Daily digest: Hostel", "Hostel: 2 unresolved")
	require.NoError(t, err)
	assert.Equal(t, "user-dsa", fake.got.GetRecipientUserId())
	assert.Equal(t, notificationv1.Channel_IN_APP, fake.got.GetChannel())
	assert.Equal(t, "digest.daily", fake.got.GetEventType())
	assert.Equal(t, "Daily digest: Hostel", fake.got.GetSubject())
	assert.Equal(t, "Hostel: 2 unresolved", fake.got.GetBody())
	assert.Nil(t, fake.got.ReportId, "a digest is not about one report")

	err = NewNotification(&fakeNotification{err: status.Error(codes.InvalidArgument, "bad")}).Send(context.Background(), "u", "s", "b")
	assert.Error(t, err)
}
