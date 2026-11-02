package grpc

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	categorytaxonomyv1 "github.com/jurwaCST/Jurwa/proto/gen/category-taxonomy/v1"
	commonv1 "github.com/jurwaCST/Jurwa/proto/gen/common/v1"
	"github.com/jurwaCST/Jurwa/services/category-taxonomy/application"
	"github.com/jurwaCST/Jurwa/services/category-taxonomy/domain"
)

// Handler implements categorytaxonomyv1.CategoryTaxonomyServiceServer.
type Handler struct {
	categorytaxonomyv1.UnimplementedCategoryTaxonomyServiceServer
	svc *application.Service
}

func NewHandler(svc *application.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListCategories(ctx context.Context, _ *categorytaxonomyv1.ListCategoriesRequest) (*categorytaxonomyv1.ListCategoriesResponse, error) {
	cats, err := h.svc.ListCategories(ctx)
	if err != nil {
		return nil, toStatus(err)
	}
	resp := &categorytaxonomyv1.ListCategoriesResponse{}
	for _, c := range cats {
		resp.Categories = append(resp.Categories, toProtoInfo(c))
	}
	return resp, nil
}

func (h *Handler) GetCategory(ctx context.Context, req *categorytaxonomyv1.GetCategoryRequest) (*categorytaxonomyv1.GetCategoryResponse, error) {
	info, err := h.svc.GetCategory(ctx, toDomainCategory(req.GetCategory()))
	if err != nil {
		return nil, toStatus(err)
	}
	return &categorytaxonomyv1.GetCategoryResponse{Category: toProtoInfo(info)}, nil
}

func (h *Handler) UpsertCategory(ctx context.Context, req *categorytaxonomyv1.UpsertCategoryRequest) (*categorytaxonomyv1.UpsertCategoryResponse, error) {
	info, err := h.svc.UpsertCategory(ctx, application.UpsertCommand{
		Category:    toDomainCategory(req.GetCategory()),
		Name:        req.GetDisplayName(),
		Description: req.GetDescription(),
		ActorRole:   domain.Role(req.GetActorRole().String()),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &categorytaxonomyv1.UpsertCategoryResponse{Category: toProtoInfo(info)}, nil
}

// toDomainCategory maps the proto enum to the domain code. CATEGORY_UNSPECIFIED
// becomes "" (InvalidArgument); a number outside the enum becomes a code that is
// not in the list (NotFound).
func toDomainCategory(c commonv1.Category) domain.Category {
	if c == commonv1.Category_CATEGORY_UNSPECIFIED {
		return ""
	}
	return domain.Category(c.String())
}

func toProtoInfo(i domain.Info) *categorytaxonomyv1.CategoryInfo {
	return &categorytaxonomyv1.CategoryInfo{
		Category:     commonv1.Category(commonv1.Category_value[string(i.Category)]),
		Name:         i.Name,
		Description:  i.Description,
		SortOrder:    int32(i.SortOrder),
		Active:       i.Active,
		HeadPosition: i.HeadPosition,
	}
}

// toStatus maps domain errors to gRPC status codes, as written in the contract.
func toStatus(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidCategory), errors.Is(err, domain.ErrInvalidName):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, domain.ErrRequiredCategory):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		slog.Error("category-taxonomy request failed", "error", err)
		return status.Error(codes.Internal, "internal error")
	}
}
