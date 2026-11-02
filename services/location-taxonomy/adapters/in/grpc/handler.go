package grpc

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	locationtaxonomyv1 "github.com/jurwaCST/Jurwa/proto/gen/location-taxonomy/v1"
	"github.com/jurwaCST/Jurwa/services/location-taxonomy/domain"
)

// UseCases is what the handler needs from the application layer.
type UseCases interface {
	ListLocations(ctx context.Context, category domain.Category) ([]domain.Entry, error)
	ValidateLocation(ctx context.Context, category domain.Category, loc domain.Location) (bool, string, error)
}

// Handler implements locationtaxonomyv1.LocationTaxonomyServiceServer.
type Handler struct {
	locationtaxonomyv1.UnimplementedLocationTaxonomyServiceServer
	app UseCases
}

// NewHandler returns a gRPC handler backed by the use cases.
func NewHandler(app UseCases) *Handler {
	return &Handler{app: app}
}

// ListLocations returns the active locations of one category. A bad category returns INVALID_ARGUMENT.
func (h *Handler) ListLocations(ctx context.Context, req *locationtaxonomyv1.ListLocationsRequest) (*locationtaxonomyv1.ListLocationsResponse, error) {
	entries, err := h.app.ListLocations(ctx, categoryFromProto(req.GetCategory()))
	if err != nil {
		return nil, toStatus(err)
	}
	resp := &locationtaxonomyv1.ListLocationsResponse{Locations: make([]*locationtaxonomyv1.LocationEntry, 0, len(entries))}
	for _, e := range entries {
		resp.Locations = append(resp.Locations, entryToProto(e))
	}
	return resp, nil
}

// ValidateLocation says whether a location exists for a category. An unknown location is a
// normal answer (valid false with a reason); only a bad request returns INVALID_ARGUMENT.
func (h *Handler) ValidateLocation(ctx context.Context, req *locationtaxonomyv1.ValidateLocationRequest) (*locationtaxonomyv1.ValidateLocationResponse, error) {
	valid, reason, err := h.app.ValidateLocation(ctx, categoryFromProto(req.GetCategory()), locationFromProto(req.GetLocation()))
	if err != nil {
		return nil, toStatus(err)
	}
	return &locationtaxonomyv1.ValidateLocationResponse{Valid: valid, Reason: reason}, nil
}

// toStatus maps domain errors to gRPC status codes and hides internal details.
func toStatus(err error) error {
	if errors.Is(err, domain.ErrInvalidInput) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Error(codes.Internal, "internal error")
}

func categoryFromProto(c commonv1.Category) domain.Category {
	switch c {
	case commonv1.Category_CATEGORY_UNSPECIFIED:
		return ""
	case commonv1.Category_HOSTEL:
		return domain.CategoryHostel
	case commonv1.Category_ACADEMIC_BLOCK:
		return domain.CategoryAcademicBlock
	case commonv1.Category_OTHERS:
		return domain.CategoryOthers
	default:
		// An enum number this version does not know; the domain rejects it as invalid.
		return domain.Category(fmt.Sprintf("UNKNOWN_%d", c))
	}
}

func categoryToProto(c domain.Category) commonv1.Category {
	switch c {
	case domain.CategoryHostel:
		return commonv1.Category_HOSTEL
	case domain.CategoryAcademicBlock:
		return commonv1.Category_ACADEMIC_BLOCK
	case domain.CategoryOthers:
		return commonv1.Category_OTHERS
	default:
		return commonv1.Category_CATEGORY_UNSPECIFIED
	}
}

func locationFromProto(l *commonv1.Location) domain.Location {
	return domain.Location{Building: l.GetBuilding(), Wing: l.GetWing(), Area: l.GetArea()}
}

func entryToProto(e domain.Entry) *locationtaxonomyv1.LocationEntry {
	return &locationtaxonomyv1.LocationEntry{
		Id:       e.ID,
		Category: categoryToProto(e.Category),
		Location: &commonv1.Location{Building: e.Location.Building, Wing: e.Location.Wing, Area: e.Location.Area},
	}
}
