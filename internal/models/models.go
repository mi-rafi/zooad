package internal

type (
	AnimalSmall struct {
		IdAnim int64
		NameAn string
		Age    int
		Gender string
		IdSp   int64
	}
	Specie struct {
		IdSp    int64
		Title   string
		Descrip string
	}

	Animal struct {
		IdAnim  int64
		NameAn  string
		Age     int
		Gender  string
		Title   string
		Descrip string
	}

	AnimalFull struct {
		Animal
		Mood Mood
	}

	Mood string
)
