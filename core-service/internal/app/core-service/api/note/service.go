package note

import (
	"context"
	"core-service/internal/domain"
	pb "core-service/pkg/core/api/core"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	pb.UnimplementedNoteServiceServer
	noteService NoteService
}

type NoteService interface {
	CreateNote(ctx context.Context, organizationID domain.ID, content string) (domain.Note, error)
	ListNotesByOrganization(ctx context.Context, organizationID domain.ID) ([]domain.Note, error)
	DeleteNote(ctx context.Context, id domain.ID) error
}

func NewService(noteService NoteService) *Service {
	return &Service{
		noteService: noteService,
	}
}

func noteToProto(note domain.Note) *pb.Note {
	return &pb.Note{
		Id:             note.ID.String(),
		OrganizationId: note.OrganizationID.String(),
		Content:        note.Content,
		CreatedAt:      timestamppb.New(note.CreatedAt),
	}
}

func (s *Service) CreateNote(ctx context.Context, req *pb.CreateNoteRequest) (*pb.CreateNoteResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.CreateNote")
	defer span.Finish()

	orgID, err := domain.ParseID(req.OrganizationId)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	note, err := s.noteService.CreateNote(ctx, orgID, req.Content)
	if err != nil {
		return nil, err
	}

	return &pb.CreateNoteResponse{
		Note: noteToProto(note),
	}, nil
}

func (s *Service) ListNotes(ctx context.Context, req *pb.ListNotesRequest) (*pb.ListNotesResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.ListNotes")
	defer span.Finish()

	orgID, err := domain.ParseID(req.OrganizationId)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	notes, err := s.noteService.ListNotesByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	pbNotes := make([]*pb.Note, 0, len(notes))
	for _, note := range notes {
		pbNotes = append(pbNotes, noteToProto(note))
	}

	return &pb.ListNotesResponse{
		Notes: pbNotes,
	}, nil
}

func (s *Service) DeleteNote(ctx context.Context, req *pb.DeleteNoteRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.DeleteNote")
	defer span.Finish()

	id, err := domain.ParseID(req.Id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	err = s.noteService.DeleteNote(ctx, id)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
