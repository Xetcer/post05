package post05

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

/*
Поскольку пакет взаимодействует с PostgreSQL, мы импортируем пакет github.com/lib/pq и используем _
перед путем к пакету. Как обсуждалось ранее, это происходит потому, что импортированный пакет
регистрирует себя в качестве обработчика базы данных для пакета sql, но не применяется непосредственно
в коде. Он используется только через пакет sql.
*/

var (
	Hostname = ""
	Port     = 2345
	Username = ""
	Password = ""
	Database = ""
)

// Переменные с заглавной буквы, чтобы могли быть доступны из вне, т.е. они Открытые

// Userdata is for holding full user data
// Userdata table + Username
type Userdata struct {
	ID          int
	Username    string
	Name        string
	Surname     string
	Description string
}

func openConnection() (*sql.DB, error) {
	// строка подключения
	conn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", Hostname, Port, Username, Password, Database)
	// открыть базу данных
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Функция возвращает ID пользователя username
// -1, если пользователь не существует
func exist(username string) int {
	username = strings.ToLower(username)

	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()
	userID := -1
	statement := fmt.Sprintf(`SELECT "id" FROM "users" where username = '%s'`, username)
	rows, err := db.Query(statement)
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			fmt.Println("Scan", err)
			return -1
		}
		userID = id
	}
	defer rows.Close()
	return userID
}

// AddUser добавляет нового пользователя в базу данных
// Возвращаем новый ID пользователя
// -1, если произошла ошибка

func AddUser(d Userdata) int {
	d.Username = strings.ToLower(d.Username)
	db, err := openConnection()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	defer db.Close()

	userID := exist(d.Username)
	if userID != -1 {
		fmt.Println("User already exist", Username)
		return -1
	}
	insertStatement := `insert into "users" ("username") values ($1)`
	_, err = db.Exec(insertStatement, d.Username)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	userID = exist(d.Username)
	if userID == -1 {
		return userID
	}

	insertStatement = `insert into "userdata" ("userid", "name", "surname", "description") values ($1, $2, $3, $4)`
	_, err = db.Exec(insertStatement, userID, d.Name, d.Surname, d.Description)
	if err != nil {
		fmt.Println("db.Exec", err)
		return -1
	}
	return userID
}

// deleteUser удаляет существующего пользователя
func DeleteUser(id int) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	// Существует ли идентификатор ?
	statement := fmt.Sprintf(`SELECT "username" FROM "users" where id = %id`, id)
	rows, err := db.Query(statement)

	var username string
	for rows.Next() {
		err = rows.Scan(&username)
		if err != nil {
			return err
		}
	}
	defer rows.Close()

	if exist(username) != id {
		return fmt.Errorf("User with ID %d does not exist", id)
	}

	// Удаление из таблицы userdata
	deleteStatement := `delete from "userdata" where userid=$1`
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}

	// удалить из users
	deleteStatement = `delete from "user" where userid=$1`
	_, err = db.Exec(deleteStatement, id)
	if err != nil {
		return err
	}
	return nil
}

func ListUsers() ([]Userdata, error) {
	Data := []Userdata{}
	db, err := openConnection()
	if err != nil {
		return Data, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT "id","username","name","surname","description" FROM "users", "userdata" WHERE users.id=userdata.userid`)
	if err != nil {
		return Data, err
	}

	for rows.Next() {
		var id int
		var username string
		var name string
		var surname string
		var description string
		err = rows.Scan(&id, &username, &name, &surname, &description)
		temp := Userdata{ID: id, Username: username, Name: name, Surname: surname, Description: description}
		Data = append(Data, temp)
		if err != nil {
			return Data, err
		}
	}
	defer rows.Close()
	return Data, nil
}

// UpdateUser предназначена для обновления данных существующего пользователя
func UpdateUser(d Userdata) error {
	db, err := openConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	userID := exist(d.Username)
	if userID == -1 {
		return errors.New("User does not exist")
	}
	d.ID = userID
	updateStatement := `update "userdata" set "name"=$1, "surename"=$2, "description"=$3 where "userid"=$4`
	_, err = db.Exec(updateStatement, d.Name, d.Username, d.Description, d.ID)
	if err != nil {
		return err
	}
	return nil
}
