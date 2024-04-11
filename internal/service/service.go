package service

import (
	"context"
	"math/rand/v2"

	"github.com/mi-raf/zooad/internal/database"
	mod "github.com/mi-raf/zooad/internal/models"
)

type (
	Service interface {
		Init(ctx context.Context) error
		Ping(ctx context.Context) error
		Close() error
	}

	AnimalService struct {
		r    database.AnimalRepository
		mood MoodService
	}
)

func NewAnimalService(r database.AnimalRepository, ms MoodService) *AnimalService {
	return &AnimalService{r: r, mood: ms}
}

func (s *AnimalService) AddAnimal(ctx context.Context, individual *mod.Animal) error {
	_, err := s.r.Add(ctx, individual)
	if err != nil {
		return err
	}
	return nil
}

func (s *AnimalService) DeleteAnimal(ctx context.Context, idAnim int64) error {
	return s.r.Delete(ctx, idAnim)
}

func (s *AnimalService) GetAnimal(ctx context.Context, idAnim int64) (*mod.AnimalFull, error) {
	animal, err := s.r.Get(ctx, idAnim)
	if err != nil {
		return nil, err
	}

	return &mod.AnimalFull{Animal: *animal, Mood: s.mood.GetMood()}, nil
}

func (s *AnimalService) GetAllAnimal(ctx context.Context, offset, limit int) ([]mod.Animal, error) {
	return s.r.GetAll(ctx, offset, limit)
}

func (s *AnimalService) Update(ctx context.Context, individ *mod.Animal) error {
	return s.r.Update(ctx, individ)

}

var (
	moodAngry = []mod.Mood{"happy", "angry", "sad", "cheerful", "I love Sencha", "I need more *4 svad`bi*"}
)

type (
	MoodService interface {
		GetMood() mod.Mood
	}

	MoodServiceImpl struct {
	}
)

func NewMoodService() *MoodServiceImpl {
	return &MoodServiceImpl{}
}

func (m *MoodServiceImpl) Init(ctx context.Context) error {
	return nil
}

func (m *MoodServiceImpl) Ping(ctx context.Context) error {
	return nil
}

func (m *MoodServiceImpl) Close() error {
	return nil
}

func (m *MoodServiceImpl) GetMood() mod.Mood {
	rand := rand.IntN(5)
	return moodAngry[rand]

}
