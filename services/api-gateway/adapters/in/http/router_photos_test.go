package http

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	identityv1 "github.com/jurwaCST/Jurwa/proto/gen/identity/v1"
	photomediav1 "github.com/jurwaCST/Jurwa/proto/gen/photo-media/v1"
)

// ---- photo routes (SCRUM-120) ----

const (
	mb = 1 << 20
	// photo-media's message for a photo over 5 MB. The gateway sends exactly
	// this when it refuses an upload itself, so the PWA handles one message.
	tooLargeMsg = "invalid photo: photo is larger than 5 MB"
)

// fakePhotos acts like photo-media: it refuses a photo over 5 MB and knows
// photo p1, uploaded by u-student.
type fakePhotos struct {
	uploaded *photomediav1.UploadPhotoRequest
	asked    string
	getErr   error
	getType  string
}

func (f *fakePhotos) UploadPhoto(_ context.Context, in *photomediav1.UploadPhotoRequest) (*photomediav1.UploadPhotoResponse, error) {
	f.uploaded = in
	if len(in.GetData()) > 5*mb {
		return nil, status.Error(codes.InvalidArgument, tooLargeMsg)
	}
	return &photomediav1.UploadPhotoResponse{PhotoId: "p1"}, nil
}

func (f *fakePhotos) GetPhoto(_ context.Context, photoID string) (*photomediav1.GetPhotoResponse, error) {
	f.asked = photoID
	if f.getErr != nil {
		return nil, f.getErr
	}
	ct := f.getType
	if ct == "" {
		ct = "image/jpeg"
	}
	data := []byte{0xFF, 0xD8, 0xFF, 0x01, 0x02}
	return &photomediav1.GetPhotoResponse{PhotoId: photoID, ContentType: ct, Data: data, SizeBytes: int64(len(data)), UploaderId: "u-student"}, nil
}

// photoForm builds a multipart body with the file in field, plus extra text fields.
func photoForm(t *testing.T, field, fileName, contentType string, data []byte, extra map[string]string) (body *bytes.Buffer, formType string) {
	t.Helper()
	body = &bytes.Buffer{}
	w := multipart.NewWriter(body)
	for k, v := range extra {
		require.NoError(t, w.WriteField(k, v))
	}
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", `form-data; name="`+field+`"; filename="`+fileName+`"`)
	h.Set("Content-Type", contentType)
	part, err := w.CreatePart(h)
	require.NoError(t, err)
	_, err = part.Write(data)
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return body, w.FormDataContentType()
}

func (e *env) upload(token string, body *bytes.Buffer, formType string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", "/photos", body)
	req.RemoteAddr = "10.0.0.1:5555"
	req.Header.Set("Content-Type", formType)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, req)
	return rec
}

func TestPhotoRoutesNeedAToken(t *testing.T) {
	e := newEnv()
	body, ft := photoForm(t, "photo", "tap.jpg", "image/jpeg", []byte{0xFF, 0xD8, 0xFF}, nil)
	assert.Equal(t, http.StatusUnauthorized, e.upload("", body, ft).Code)
	assert.Equal(t, http.StatusUnauthorized, e.do("GET", "/photos/p1", "", "").Code)
	assert.Nil(t, e.photos.uploaded)
	assert.Empty(t, e.photos.asked)
}

func TestUploadPhotoSendsTheFileAndTheUploaderFromTheJWT(t *testing.T) {
	e := newEnv()
	data := []byte{0xFF, 0xD8, 0xFF, 0x10, 0x20}
	// A forged uploaderId form field must be ignored.
	body, ft := photoForm(t, "photo", "tap.jpg", "image/jpeg", data, map[string]string{"uploaderId": "u-evil"})
	rec := e.upload("student-token", body, ft)

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	assert.JSONEq(t, `{"photoId":"p1"}`, rec.Body.String())
	got := e.photos.uploaded
	require.NotNil(t, got)
	assert.Equal(t, "u-student", got.GetUploaderId(), "the uploader comes from the JWT, never the form")
	assert.Equal(t, "image/jpeg", got.GetContentType(), "the part's Content-Type is passed on")
	assert.Equal(t, "tap.jpg", got.GetFileName())
	assert.Equal(t, data, got.GetData())
}

// The gateway does no type check: photo-media checks the real bytes.
func TestUploadPhotoPassesTheTypeThrough(t *testing.T) {
	e := newEnv()
	body, ft := photoForm(t, "photo", "doc.pdf", "application/pdf", []byte("%PDF-1.4"), nil)
	require.Equal(t, http.StatusCreated, e.upload("student-token", body, ft).Code)
	assert.Equal(t, "application/pdf", e.photos.uploaded.GetContentType())
}

func TestUploadPhotoOnlyStudents(t *testing.T) {
	for _, token := range []string{"head-token", "staff-token", "admin-token"} {
		e := newEnv()
		body, ft := photoForm(t, "photo", "tap.jpg", "image/jpeg", []byte{0xFF, 0xD8, 0xFF}, nil)
		assert.Equal(t, http.StatusForbidden, e.upload(token, body, ft).Code, token)
		assert.Nil(t, e.photos.uploaded, token)
	}
}

// Three size cases agreed with Choki and Nyendrak: within 5 MB, 5 to 6 MB
// (photo-media refuses it), over 6 MB (the gateway refuses it). Both refusals
// are a 400 with the same message.
func TestUploadPhotoSizes(t *testing.T) {
	tests := []struct {
		name         string
		size         int
		wantCode     int
		reachesMedia bool
	}{
		{name: "exactly 5 MB", size: 5 * mb, wantCode: http.StatusCreated, reachesMedia: true},
		{name: "5 MB + 1 byte: photo-media refuses", size: 5*mb + 1, wantCode: http.StatusBadRequest, reachesMedia: true},
		{name: "5.5 MB: photo-media refuses", size: 5*mb + mb/2, wantCode: http.StatusBadRequest, reachesMedia: true},
		{name: "6 MB + 1 byte: the gateway refuses", size: 6*mb + 1, wantCode: http.StatusBadRequest, reachesMedia: false},
		{name: "10 MB: the gateway refuses", size: 10 * mb, wantCode: http.StatusBadRequest, reachesMedia: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv()
			body, ft := photoForm(t, "photo", "big.jpg", "image/jpeg", bytes.Repeat([]byte{0xFF}, tt.size), nil)
			rec := e.upload("student-token", body, ft)
			require.Equal(t, tt.wantCode, rec.Code, rec.Body.String())
			assert.Equal(t, tt.reachesMedia, e.photos.uploaded != nil)
			if tt.wantCode == http.StatusBadRequest {
				assert.JSONEq(t, `{"error":"`+tooLargeMsg+`"}`, rec.Body.String())
			}
		})
	}
}

func TestUploadPhotoBadForm(t *testing.T) {
	e := newEnv()
	rec := e.do("POST", "/photos", "student-token", `{"photo":"abc"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code, "JSON instead of multipart")
	assert.Contains(t, decode(t, rec)["error"], "invalid photo")

	e = newEnv()
	body, ft := photoForm(t, "file", "tap.jpg", "image/jpeg", []byte{0xFF, 0xD8, 0xFF}, nil)
	rec = e.upload("student-token", body, ft)
	assert.Equal(t, http.StatusBadRequest, rec.Code, "no field called photo")
	assert.Contains(t, decode(t, rec)["error"], "photo")
	assert.Nil(t, e.photos.uploaded, "a bad form never reaches photo-media")
}

func TestGetPhotoReturnsTheImage(t *testing.T) {
	e := newEnv()
	rec := e.do("GET", "/photos/p1", "student-token", "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "p1", e.photos.asked)
	assert.Equal(t, "image/jpeg", rec.Header().Get("Content-Type"))
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "private, no-store", rec.Header().Get("Cache-Control"))
	assert.Equal(t, []byte{0xFF, 0xD8, 0xFF, 0x01, 0x02}, rec.Body.Bytes(), "the raw image, ready for an <img> blob URL")
}

func TestGetPhotoOwnership(t *testing.T) {
	for _, token := range []string{"student-token", "head-token", "staff-token", "admin-token"} {
		e := newEnv()
		assert.Equal(t, http.StatusOK, e.do("GET", "/photos/p1", token, "").Code, token)
	}

	users["other-student"] = &identityv1.User{Id: "u-other", Roles: []commonv1.Role{commonv1.Role_STUDENT}}
	defer delete(users, "other-student")
	e := newEnv()
	rec := e.do("GET", "/photos/p1", "other-student", "")
	assert.Equal(t, http.StatusForbidden, rec.Code, "another student's photo")
	assert.NotEqual(t, "image/jpeg", rec.Header().Get("Content-Type"), "no image bytes are sent")
}

func TestGetPhotoErrors(t *testing.T) {
	e := newEnv()
	e.photos.getErr = status.Error(codes.NotFound, "photo not found")
	assert.Equal(t, http.StatusNotFound, e.do("GET", "/photos/p9", "head-token", "").Code)

	e = newEnv()
	e.photos.getErr = status.Error(codes.InvalidArgument, "invalid photo: id must be a UUID")
	assert.Equal(t, http.StatusBadRequest, e.do("GET", "/photos/xyz", "head-token", "").Code)
}

// Only JPG and PNG are sent with their own type; anything else is sent as a
// download type, so a browser never runs it.
func TestGetPhotoUnexpectedTypeIsNotServedAsIs(t *testing.T) {
	e := newEnv()
	e.photos.getType = "text/html"
	rec := e.do("GET", "/photos/p1", "head-token", "")
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/octet-stream", rec.Header().Get("Content-Type"))
}
