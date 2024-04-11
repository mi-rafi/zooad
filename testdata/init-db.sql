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

INSERT INTO Species (title, descrip) VALUES('cat', 'The party gave out a bowl of rice and a cat wife');
INSERT INTO Species (title, descrip) VALUES('dog', 'I ll buy you a dog');
INSERT INTO Species (title, descrip) VALUES('rat', 'You are a rat, and I am a rat');

INSERT INTO Animals (name_an, age, gender, id_sp) VALUES('Klepa', 15, 'f', (SELECT id_sp FROM species 
WHERE title = 'cat'));
INSERT INTO Animals (name_an, age, gender, id_sp) VALUES('Wahaha', 1, 'm',  (SELECT id_sp FROM species 
WHERE title = 'cat'));
INSERT INTO Animals (name_an, age, gender, id_sp) VALUES('Zu', 24, 'f',  (SELECT id_sp FROM species 
WHERE title = 'cat'));
INSERT INTO Animals (name_an, age, gender, id_sp) VALUES('Zina', 2, 'f',  (SELECT id_sp FROM species 
WHERE title = 'cat'));
INSERT INTO Animals (name_an, age, gender, id_sp) VALUES('Tom', 32, 'm',  (SELECT id_sp FROM species 
WHERE title = 'cat'));

