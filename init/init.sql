CREATE TABLE IF NOT EXISTS Species (
    id_sp bigserial PRIMARY KEY,
    title varchar(40) NOT NULL UNIQUE CONSTRAINT non_empty_title CHECK(length(title)>0),
    descrip varchar(400) NOT NULL  CONSTRAINT non_empty_desc CHECK(length(descrip)>0)

);

CREATE TABLE IF NOT EXISTS Animals (
    id_anim bigserial PRIMARY KEY,
    name_an varchar(40) NOT NULL,
    age integer NOT NULL,
    gender varchar(1),
    id_sp integer REFERENCES species(id_sp)
);
