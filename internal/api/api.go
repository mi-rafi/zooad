package api

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	models "github.com/mi-raf/zooad/internal/models"
	"github.com/mi-raf/zooad/internal/service"
	zl "github.com/rs/zerolog/log"
)

const (
	MAX_LIMIT = 100
)

type (
	Config struct {
		Addr string
	}

	API struct {
		e    *echo.Echo
		s    *service.AnimalService
		addr string
	}

	Context struct {
		echo.Context
		Ctx context.Context
	}
)

func New(ctx context.Context, cfg *Config, s *service.AnimalService) (*API, error) {
	e := echo.New()
	a := &API{
		s:    s,
		e:    e,
		addr: cfg.Addr,
	}

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &Context{
				Context: c,
				Ctx:     ctx,
			}
			return next(cc)
		}
	})
	//TODO запроосы для рест
	e.Use(logger())
	e.GET("/health", healthCheck)
	e.GET("/animal/:id", a.getAnimal)
	e.GET("/animal", a.getAllAnimal)
	e.POST("/animal", a.addAnimal)
	e.PUT("/animal", a.updateAnimal)
	e.DELETE("/animal/:id", a.deleteAnimal)
	return a, nil
}

func logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()

			err := next(c)
			stop := time.Now()

			zl.Debug().
				Str("remote", req.RemoteAddr).
				Str("user_agent", req.UserAgent()).
				Str("method", req.Method).
				Str("path", c.Path()).
				Int("status", res.Status).
				Dur("duration", stop.Sub(start)).
				Str("duration_human", stop.Sub(start).String()).
				Msgf("called url %s", req.URL)
			return err
		}
	}
}

func healthCheck(e echo.Context) error {
	return e.JSON(http.StatusOK, struct {
		Message string
	}{Message: "OK"})
}

type (
	//todo
	mineAnimalfull struct {
		IdAnim  int64  `json:"id_anim"`
		NameAn  string `json:"name_animal"`
		Age     int    `json:"age"`
		Gender  string `json:"gender"`
		Title   string `json:"title"`
		Descrip string `json:"description"`
		Mood    string `json:"mood"`
	}

	mineRes struct {
		str string `json:"string for res"`
	}

	mineError struct {
		msg string `json:"msg"`
	}
)

func (a *API) getAnimal(e echo.Context) error {
	cc, err := getParentContext(e)
	if err != nil {
		return err
	}
	id, err := strconv.ParseInt(e.Param("id"), 10, 64)
	if err != nil {
		zl.Error().Msg("id is empty")
		return e.NoContent(echo.ErrBadRequest.Code)
	}

	animal, err := a.s.GetAnimal(cc.Ctx, id)
	if err != nil {
		zl.Error().Err(err).Int64("mine id animal", id).Msg("can't find animal")
		return err
	}
	res := &mineAnimalfull{animal.IdAnim, animal.NameAn, animal.Age, animal.Gender, animal.Title, animal.Descrip, string(animal.Mood)}
	return e.JSON(http.StatusOK, res)
}

func (a *API) getAllAnimal(e echo.Context) error {
	cc, err := getParentContext(e)
	if err != nil {
		return err
	}
	limit, err := strconv.Atoi(e.QueryParam("limit"))
	if err != nil || limit <= 0 {
		return e.JSON(echo.ErrBadRequest.Code, mineError{msg: "incorrect limit"})
	}

	offset, err := strconv.Atoi(e.QueryParam("offset"))
	if err != nil || offset < 0 {
		return e.JSON(echo.ErrBadRequest.Code, mineError{msg: "incorrect offset"})
	}
	limit = min(limit, MAX_LIMIT)

	animals, err := a.s.GetAllAnimal(cc.Ctx, offset, limit)
	if err != nil || offset < 0 {
		zl.Error().Err(err).Str("error", "it is no OK").Msg("can't find animal")
		return err
	}
	return e.JSON(http.StatusOK, animals)
}

func (a *API) addAnimal(e echo.Context) error {
	cc, err := getParentContext(e)
	if err != nil {
		return err
	}
	age, err := strconv.Atoi(e.QueryParam("age"))
	if err != nil {
		return e.JSON(echo.ErrBadRequest.Code, mineError{msg: "incorrect age of animal"})
	}

	newAnimal := models.Animal{
		NameAn:  e.QueryParam("name_animal"),
		Age:     age,
		Gender:  e.QueryParam("gender"),
		Title:   e.QueryParam("title"),
		Descrip: e.QueryParam("description")}
	err = a.s.AddAnimal(cc.Ctx, &newAnimal)
	if err != nil {
		return e.JSON(echo.ErrNotImplemented.Code, mineError{msg: "don't create animal"})
	}
	return e.JSON(http.StatusCreated, mineRes{str: "correct create animal"})
}

func (a *API) updateAnimal(e echo.Context) error {
	cc, err := getParentContext(e)
	if err != nil {
		return err
	}
	age, err := strconv.Atoi(e.QueryParam("age"))
	if err != nil {
		return e.JSON(echo.ErrBadRequest.Code, mineError{msg: "incorrect age of animal"})
	}

	newAnimal := models.Animal{
		NameAn:  e.QueryParam("name_animal"),
		Age:     age,
		Gender:  e.QueryParam("gender"),
		Title:   e.QueryParam("title"),
		Descrip: e.QueryParam("description")}
	err = a.s.Update(cc.Ctx, &newAnimal)
	if err != nil {
		return e.JSON(echo.ErrNotImplemented.Code, mineError{msg: "don't update animal"})
	}
	return e.JSON(http.StatusOK, mineRes{str: "correct update animal"})

}

func (a *API) deleteAnimal(e echo.Context) error {
	cc, err := getParentContext(e)
	if err != nil {
		return err
	}
	id, err := strconv.ParseInt(e.Param("id"), 10, 64)
	if err != nil {
		zl.Error().Msg("id is empty")
		return errors.New("empty id")
	}
	err = a.s.DeleteAnimal(cc.Ctx, id)
	if err != nil {
		zl.Error().Err(err).Int64("mine id anim", id).Msg("can't delete animal")
		return err
	}
	res := &mineRes{str: "you kill that animal!!!!"}
	return e.JSON(http.StatusOK, res)
}

func getParentContext(e echo.Context) (*Context, error) {
	cc, ok := e.(*Context)
	if !ok {
		zl.Error().Interface("type", reflect.TypeOf(e)).Msg("can't cast to custom context")
		return nil, echo.ErrInternalServerError
	}
	return cc, nil
}

func (a *API) Start() error {
	zl.Debug().Msgf("listening on %v", a.addr)
	err := a.e.Start(a.addr)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (a *API) Close() error {
	return a.e.Close()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
