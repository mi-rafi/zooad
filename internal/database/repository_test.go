package database_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	database "github.com/mi-raf/zooad/internal/database"
	models "github.com/mi-raf/zooad/internal/models"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RepositoryTestSuite struct {
	suite.Suite
	r           database.AnimalRepository
	pgContainer *postgres.PostgresContainer
	ctx         context.Context
}

func (suite *RepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error
	suite.pgContainer, err = postgres.RunContainer(suite.ctx,
		testcontainers.WithImage("postgres:15.3-alpine"),
		postgres.WithInitScripts(filepath.Join("..", "..", "testdata", "init-db.sql")),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	suite.NoError(err)
	connStr, err := suite.pgContainer.ConnectionString(suite.ctx, "sslmode=disable")

	suite.NoError(err)
	p, err := pgxpool.New(suite.ctx, connStr)
	suite.NoError(err)
	suite.r, err = database.NewAnimalRepository(suite.ctx, p)
	suite.NoError(err)

}

func (s *RepositoryTestSuite) TearDownSuite() {
	err := s.pgContainer.Terminate(s.ctx)
	s.NoError(err)
}

func (s *RepositoryTestSuite) TestCreateAnimals() {
	//given
	expAnimalFull := &models.Animal{
		IdAnim:  10500,
		NameAn:  "star",
		Age:     23,
		Gender:  "f",
		Title:   "cat",
		Descrip: "pp"}
	//when
	id, err := s.r.Add(s.ctx, expAnimalFull)
	//then
	s.NoError(err)
	s.NotEqual(-1, id)

	animalFull, err := s.r.Get(s.ctx, id)
	s.NoError(err)
	s.NotNil(animalFull)
	s.Equal(expAnimalFull.NameAn, animalFull.NameAn)
	s.Equal(expAnimalFull.Age, animalFull.Age)
	s.Equal(expAnimalFull.Gender, animalFull.Gender)
	s.Equal(expAnimalFull.Title, animalFull.Title)
	s.Equal("The party gave out a bowl of rice and a cat wife", animalFull.Descrip)
	s.NotEqual(expAnimalFull.IdAnim, animalFull.IdAnim)

}

func (s *RepositoryTestSuite) TestCreateAnimalsWithError() {
	//given
	animalFull := &models.Animal{}
	id, err := s.r.Add(s.ctx, animalFull)
	s.Error(err)
	var resultId int64
	resultId = -1
	s.Equal(resultId, id)

}

func (s *RepositoryTestSuite) TestGetAnimals() {
	//when
	animalFull, err := s.r.Get(s.ctx, 2)
	//then
	s.NoError(err)
	s.NotNil(animalFull)
	s.Equal("Wahaha", animalFull.NameAn)
	s.Equal(1, animalFull.Age)
	s.Equal("m", animalFull.Gender)
	s.Equal("cat", animalFull.Title)
	s.Equal("The party gave out a bowl of rice and a cat wife", animalFull.Descrip)
}

func (s *RepositoryTestSuite) TestGetAnimalsWithError() {
	//when
	animalFull, err := s.r.Get(s.ctx, 10)
	//then
	s.Error(err)
	animalFullEmpty := &models.Animal{}
	s.Equal(animalFullEmpty, animalFull)
}

func (s *RepositoryTestSuite) TestGetAllAnimals() {
	//when
	animals, err := s.r.GetAll(s.ctx, 2, 3)
	//then
	s.NoError(err)
	s.NotNil(animals)
	s.Equal(3, len(animals))
}

func (s *RepositoryTestSuite) TestGetAllAnimalsWithoutRows() {
	//when
	animals, err := s.r.GetAll(s.ctx, 123, 300)
	//then
	s.NoError(err)
	s.NotNil(animals)
	s.Equal(0, len(animals))
}

func (s *RepositoryTestSuite) TestDeleteAnimals() {
	//when
	err := s.r.Delete(s.ctx, 1)
	//then
	s.NoError(err)
	animalFull, _ := s.r.Get(s.ctx, 1)
	animalFullEmpty := &models.Animal{}
	s.Equal(animalFullEmpty, animalFull)
}

func (s *RepositoryTestSuite) TestDeleteWithoutAnimals() {
	//when
	beforeArr, _ := s.r.GetAll(s.ctx, 0, 600)
	err := s.r.Delete(s.ctx, 500)
	//then
	s.NoError(err)

	afterArr, _ := s.r.GetAll(s.ctx, 0, 600)
	s.Equal(len(beforeArr), len(afterArr))

}

func (s *RepositoryTestSuite) TestUpdateAnimals() {
	//given
	individ := models.Animal{
		IdAnim:  3,
		NameAn:  "star",
		Age:     60,
		Gender:  "m",
		Title:   "rat",
		Descrip: "pp"}
	//when
	s.r.Update(s.ctx, &individ)
	//then

	animalFull, err := s.r.Get(s.ctx, 3)
	s.NoError(err)
	s.NotNil(animalFull)
	s.Equal("star", animalFull.NameAn)
	s.Equal(60, animalFull.Age)
	s.Equal("m", animalFull.Gender)
	s.Equal("rat", animalFull.Title)
	s.Equal("You are a rat, and I am a rat", animalFull.Descrip)

}

func (s *RepositoryTestSuite) TestUpdateWithoutAnimals() {
	//given
	individ := models.Animal{
		IdAnim:  666,
		NameAn:  "star",
		Age:     60,
		Gender:  "m",
		Title:   "rat",
		Descrip: "pp"}
	//when
	err := s.r.Update(s.ctx, &individ)
	s.NoError(err)
	//then
	animalFull, _ := s.r.Get(s.ctx, 666)
	animalFullEmpty := &models.Animal{}
	s.Equal(animalFullEmpty, animalFull)

}

func (s *RepositoryTestSuite) TestUpdateAnimalsErrorTitle() {
	//given
	exAnimalFull, _ := s.r.Get(s.ctx, 2)
	individ := models.Animal{
		IdAnim:  2,
		NameAn:  "star",
		Age:     60,
		Gender:  "m",
		Title:   "fox",
		Descrip: "pp"}
	//when
	err := s.r.Update(s.ctx, &individ)
	s.Error(err)
	//then
	animalFull, _ := s.r.Get(s.ctx, 2)
	s.Equal(exAnimalFull, animalFull)

}

func TestCustomerRepoTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
