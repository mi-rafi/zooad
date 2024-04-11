package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	models "github.com/mi-raf/zooad/internal/models"
)

const (
	delete     = "DELETE FROM Animals WHERE id_anim = $1"
	insert     = "INSERT INTO Animals (name_an, age, gender, id_sp) VALUES($1, $2, $3, $4) RETURNING id_anim"
	searchIdSp = "SELECT id_sp FROM Species WHERE title = $1"
	search     = `SELECT id_anim, name_an, age, gender, title, descrip FROM 
	Animals JOIN Species ON Animals.id_sp = Species.id_sp
	WHERE id_anim = $1`
	searchGetAll = `SELECT id_anim, name_an, age, gender, title, descrip FROM 
	Animals JOIN Species ON Animals.id_sp = Species.id_sp
	ORDER BY id_anim
	LIMIT $1
	OFFSET $2`
	update = "UPDATE Animals SET name_an = $1, age = $2, gender = $3, id_sp = (SELECT id_sp FROM Species WHERE title = $4) WHERE id_anim = $5;"
)

type AnimalRepository interface {
	Delete(ctx context.Context, idAnim int64) error
	Add(ctx context.Context, individual *models.Animal) (int64, error)
	Get(ctx context.Context, idAnim int64) (*models.Animal, error)
	GetAll(ctx context.Context, offset, limit int) ([]models.Animal, error)
	Update(ctx context.Context, individual *models.Animal) error
}

// type PgAnimalRepository struct {
// 	pool *pgxpool.Pool
// }

// todo
type PgAnimalRepository struct {
	pool *pgxpool.Pool

	//mux         sync.RWMutex
	requestTime time.Time
}

// todo
// func (r *PgAnimalRepository) Init(ctx context.Context) error {
// 	conn := ""
// 	NewAnimalRepository(context.TODO(), conn)
// 	panic("t")
// }

// todo
func (r *PgAnimalRepository) Ping(ctx context.Context) error {
	if time.Since(r.requestTime).Minutes() > 1 {
		r.requestTime = time.Now()
		return r.refresh(ctx)
	}
	return nil
}

func (r *PgAnimalRepository) Close(ctx context.Context) error {
	r.pool.Close()
	return nil
}

// todo
func (r *PgAnimalRepository) refresh(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func NewAnimalRepository(ctx context.Context, p *pgxpool.Pool) (*PgAnimalRepository, error) {

	return &PgAnimalRepository{pool: p}, nil
}

func (r *PgAnimalRepository) Delete(ctx context.Context, idAnim int64) error {
	_, err := r.pool.Exec(ctx, delete, idAnim)
	return err
}

func (r *PgAnimalRepository) Add(ctx context.Context, individual *models.Animal) (int64, error) {

	tx, err := r.pool.Begin(ctx)

	if err != nil {
		return -1, err
	}

	defer func() {
		err = tx.Rollback(ctx)
		if !errors.Is(err, pgx.ErrTxClosed) {
			fmt.Printf("Undefinded error in tx: %s\n", err)
		}
	}()

	var id_sp int64
	err = tx.QueryRow(ctx, searchIdSp, individual.Title).Scan(&id_sp)
	if err != nil {
		return -1, err
	}

	var id_an int64
	err = tx.QueryRow(ctx, insert, individual.NameAn, individual.Age, individual.Gender, id_sp).Scan(&id_an)
	if err != nil {
		return -1, err
	}

	return id_an, tx.Commit(ctx)
}

// TODO This is not working
func (r *PgAnimalRepository) Get(ctx context.Context, idAnim int64) (*models.Animal, error) {

	animalFull := models.Animal{}

	err := r.pool.QueryRow(ctx, search, idAnim).Scan(&animalFull.IdAnim, &animalFull.NameAn, &animalFull.Age, &animalFull.Gender, &animalFull.Title, &animalFull.Descrip)

	return &animalFull, err
}

func (r *PgAnimalRepository) GetAll(ctx context.Context, offset, limit int) ([]models.Animal, error) {

	rows, err := r.pool.Query(ctx, searchGetAll, limit, offset)
	if err != nil {
		return nil, err
	}

	animalsFull := make([]models.Animal, 0)

	for rows.Next() {
		var an models.Animal
		err = rows.Scan(&an.IdAnim, &an.NameAn, &an.Age, &an.Gender, &an.Title, &an.Descrip)
		if err != nil {
			return nil, err
		}
		animalsFull = append(animalsFull, an)
	}
	// check rows.Err() after the last rows.Next() :
	if err := rows.Err(); err != nil {
		// on top of errors triggered by bad conditions on the 'rows.Scan()' call,
		// there could also be some bad things like a truncated response because
		// of some network error, etc ...
		fmt.Printf("*** iteration error: %s", err)
		return nil, err
	}

	return animalsFull, nil
}

func (r *PgAnimalRepository) Update(ctx context.Context, individual *models.Animal) error {
	var newTitle string
	r.pool.QueryRow(ctx, "SELECT title FROM Species WHERE title = $1", individual.Title).Scan(&newTitle)
	if newTitle != individual.Title {
		return errors.New("title is not exists")
	}
	_, err := r.pool.Exec(ctx, update, individual.NameAn, individual.Age, individual.Gender, individual.Title, individual.IdAnim)
	return err
}
