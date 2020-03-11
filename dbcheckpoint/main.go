package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/lib/pq"
)

var db *sql.DB
var tpl *template.Template

func init() {
	var err error
	db, err = sql.Open("postgres", "postgres://bbond:password@localhost/mydb?sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("You connected to your database.")

	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

type Std struct {
	ID       int
	Name     string
	Username string
	Age      int
	Level    string
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/students", studentsIndex)
	http.HandleFunc("/students/show", studentsShow)
	http.HandleFunc("/students/create", studentsCreateForm)
	http.HandleFunc("/students/create/process", studentsCreateProcess)
	http.HandleFunc("/students/update", studentsUpdateForm)
	http.HandleFunc("/students/update/process", studentsUpdateProcess)
	http.HandleFunc("/students/delete/process", studentsDeleteProcess)
	http.ListenAndServe(":8080", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/students", http.StatusSeeOther)
}

func studentsIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT * FROM student")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()

	stds := make([]Std, 0)
	for rows.Next() {
		st := Std{}
		err := rows.Scan(&st.ID, &st.Name, &st.Username, &st.Age, &st.Level) // order matters
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		stds = append(stds, st)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	tpl.ExecuteTemplate(w, "students.gohtml", stds)
}

func studentsShow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	row := db.QueryRow("SELECT * FROM student WHERE id = $1", id)

	st := Std{}
	err := row.Scan(&st.ID, &st.Name, &st.Username, &st.Age, &st.Level)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	tpl.ExecuteTemplate(w, "show.gohtml", st)
}

func studentsCreateForm(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "create.gohtml", nil)
}

func studentsCreateProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	// get form values
	st := Std{}
	d := r.FormValue("id")
	st.Name = r.FormValue("name")
	st.Username = r.FormValue("title")
	a := r.FormValue("age")
	st.Level = r.FormValue("level")

	// validate form values
	if d == "" || st.Name == "" || st.Username == "" || a == "" || st.Level == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	// insert values
	_, err := db.Exec("INSERT INTO student (id, name, username, age, level) VALUES ($1, $2, $3, $4, $5)", st.ID, st.Name, st.Username, st.Age, st.Level)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	// confirm insertion
	tpl.ExecuteTemplate(w, "created.gohtml", st)
}

func studentsUpdateForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	row := db.QueryRow("SELECT * FROM student WHERE id = $1", id)

	st := Std{}
	err := row.Scan(&st.ID, &st.Name, &st.Username, &st.Age, &st.Level)
	switch {
	case err == sql.ErrNoRows:
		http.NotFound(w, r)
		return
	case err != nil:
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	tpl.ExecuteTemplate(w, "update.gohtml", st)
}

func studentsUpdateProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	// get form values
	st := Std{}
	d := r.FormValue("id")
	st.Name = r.FormValue("name")
	st.Username = r.FormValue("title")
	a := r.FormValue("age")
	st.Level = r.FormValue("level")

	// validate form values
	if d == "" || st.Name == "" || st.Username == "" || a == "" || st.Level == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	// insert values
	_, err := db.Exec("UPDATE student SET id = $1, name=$2, username=$3, age=$4, level=$5 WHERE id=$1;", st.ID, st.Name, st.Username, st.Age, st.Level)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	// confirm insertion
	tpl.ExecuteTemplate(w, "updated.gohtml", st)
}

func studentsDeleteProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM student WHERE id=$1;", id)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/students", http.StatusSeeOther)
}
