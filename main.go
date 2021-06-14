package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
)

// DB related stuff
var settings mysql.ConnectionURL
var sess db.Session
var err error

func setupDB(sqlFilename string, jsonFilename string) {
	// Set logging level to DEBUG
	// db.LC().SetLevel(db.LogLevelDebug)

	// Set up the database
	for sess, err = mysql.Open(settings); err != nil; {
		sess, err = mysql.Open(settings)
	}
	sqlBytes, err := ioutil.ReadFile(sqlFilename)
	check(err)
	setupScript := string(sqlBytes)
	commands := strings.Split(setupScript, ";")
	for _, v := range commands {
		s := strings.TrimSpace(v)
		if 0 < len(s) {
			sess.SQL().Exec(s)
		}
	}

	// Insert some data in the db
	jsonBytes, err := ioutil.ReadFile(jsonFilename)
	check(err)
	var books []Book
	err = json.Unmarshal(jsonBytes, &books)
	check(err)
	for _, book := range books {
		sess.Collection("books").Insert(book)
	}
}

// connection related stuff
var myRouter *mux.Router

func handleRequests() {
	myRouter = mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/genres", getGenres)
	myRouter.HandleFunc("/books", getBooks)
	myRouter.HandleFunc("/book", postBook).Methods("POST")          // Create
	myRouter.HandleFunc("/book/{id}", putBook).Methods("PUT")       // Update
	myRouter.HandleFunc("/book/{id}", deleteBook).Methods("DELETE") // Delete
	myRouter.HandleFunc("/book/{id}", getBook)                      // Read
	log.Fatal(http.ListenAndServe(":1151", myRouter))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit: homePage")
	fmt.Fprintf(w, "Welcome to the bookstore homepage!\n"+
		"\n"+
		"Available endpoints:\n"+
		"\tGET    /genres      - get the available genres\n"+
		"\tGET    /books       - get all books\n"+
		"\t\tThe following filters can can be sent in the query string, optionally:\n"+
		"\t\t\tname - strictly match a book's name\n"+
		"\t\t\tminPrice, maxPrice - a price range, inclusively\n"+
		"\t\t\tgenre - a genre's ID\n"+
		"\tPOST   /book        - save a book submitted in the request body, the book ID is generated\n"+
		"\tPUT    /book/{id}   - update a book identified by its ID\n"+
		"\tDELETE /book/{id}   - delete a book by its ID\n"+
		"\tGET    /book/{id}   - get a book by its ID\n")
}

func getGenres(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit: getGenres")
	if nil == genres {
		err := sess.Collection("genres").Find().All(&genres)
		check(err)
	}
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genres)
}

func getBooks(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit: getBooks")
	var books []Book
	err := sess.Collection("books").Find().All(&books)
	check(err)
	var auxBooks []Book
	for _, book := range books {
		if book.Amount > 0 {
			auxBooks = append(auxBooks, book)
		}
	}
	books = auxBooks
	auxBooks = []Book{}
	// TODO implement filtering
	queryString := r.URL.Query()
	name := queryString.Get("name")
	if name != "" {
		for _, book := range books {
			if book.Name == name {
				auxBooks = append(auxBooks, book)
			}
		}
		books = auxBooks
		auxBooks = []Book{}
	}
	minPrice := queryString.Get("minPrice")
	mp, err := strconv.ParseFloat(minPrice, 64)
	if err == nil || minPrice != "" {
		for _, book := range books {
			if book.Price >= mp {
				auxBooks = append(auxBooks, book)
			}
		}
		books = auxBooks
		auxBooks = []Book{}
	}
	maxPrice := queryString.Get("maxPrice")
	mp, err = strconv.ParseFloat(maxPrice, 64)
	if err == nil || maxPrice != "" {
		for _, book := range books {
			if book.Price <= mp {
				auxBooks = append(auxBooks, book)
			}
		}
		books = auxBooks
		auxBooks = []Book{}
	}
	genre := queryString.Get("genre")
	genreID, err := strconv.Atoi(genre)
	if err == nil || genre != "" {
		for _, book := range books {
			if book.GenreID == uint(genreID) {
				auxBooks = append(auxBooks, book)
			}
		}
		books = auxBooks
	}
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func postBook(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit: postBook")
	reqBody, err := ioutil.ReadAll(r.Body)
	check(err)
	var book Book
	err = json.Unmarshal(reqBody, &book)
	check(err)
	if validateBook(book) {
		res := sess.Collection("books").Find("name", book.Name)
		count, err := res.Count()
		check(err)
		if count > 0 {
			err = res.One(&book)
			check(err)
			fmt.Fprintf(w, "Book already exist, with id %d.\n"+
				"\n"+
				"Tip: try updating its count!", book.ID)
			w.WriteHeader(http.StatusNotAcceptable)
		} else {
			sess.Collection("books").Insert(book)
			err = sess.Collection("books").Find("name", book.Name).One(&book)
			check(err)
			fmt.Fprintf(w, "Book saved with ID: %d", book.ID)
			w.WriteHeader(http.StatusCreated)
		}
	} else {
		fmt.Fprintf(w, "Invalid input!\n"+
			"\n"+
			"All fields are required:\n"+
			"Name - string, max length - 100 characters.\n"+
			"Price - float, >= 0\n"+
			"Genre - int, a valid genre ID\n"+
			"Amount - int, >= 0")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func validateBook(book Book) bool {
	if len(book.Name) == 0 || len(book.Name) > 100 {
		return false
	}
	if book.Price <= 0 {
		return false
	}
	genreValid := false
	if nil == genres {
		err := sess.Collection("genres").Find().All(&genres)
		check(err)
	}
	for _, genre := range genres {
		if book.GenreID == genre.ID {
			genreValid = true
		}
	}
	if !genreValid {
		return false
	}
	if book.Amount <= 0 {
		return false
	}
	return true
}

func putBook(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit: putBook")
	vars := mux.Vars(r)
	id := vars["id"]
	reqBody, err := ioutil.ReadAll(r.Body)
	check(err)
	var book Book
	err = json.Unmarshal(reqBody, &book)
	check(err)
	if validateBook(book) {
		res := sess.Collection("books").Find("id", id)
		count, err := res.Count()
		check(err)
		if count == 0 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "No book found with the id %v", id)
		} else {
			res.Update(book)
			err = sess.Collection("books").Find("id", id).One(&book)
			check(err)
			fmt.Fprintf(w, "Book with ID: %d updated", book.ID)
			w.WriteHeader(http.StatusAccepted)
		}
	} else {
		fmt.Fprintf(w, "Invalid input!\n"+
			"\n"+
			"All fields are required:\n"+
			"Name - string, max length - 100 characters.\n"+
			"Price - float, >= 0\n"+
			"Genre - int, a valid genre ID\n"+
			"Amount - int, >= 0")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit: deleteBook")
	id := mux.Vars(r)["id"]
	res := sess.Collection("books").Find("id", id)
	count, _ := res.Count()
	if count == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "No book found with the id %v", id)
	} else {
		var book Book
		res.One(&book)
		err := res.Delete()
		check(err)
		fmt.Fprintf(w, "Deleted the book \"%v\", with the id %v", book.Name, id)
		w.WriteHeader(http.StatusNoContent)
	}
}

func getBook(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: getBook")
	id := mux.Vars(r)["id"]
	res := sess.Collection("books").Find("id", id)
	count, err := res.Count()
	check(err)
	if count == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "No book found with the id %v", id)
	} else {
		var book Book
		res.One(&book)
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(book)
		w.WriteHeader(http.StatusFound)
	}
}

var genres []Genre = nil

type Genre struct {
	ID        uint   `db:"id" json:"id"`
	GenreName string `db:"genre" json:"genre"`
}

type Book struct {
	ID      uint    `db:"id,omitempty" json:"id,omitempty"`
	Name    string  `db:"name" json:"name"`
	Price   float64 `db:"price" json:"price"`
	GenreID uint    `db:"genre_id" json:"genre"`
	Amount  uint    `db:"amount" json:"amount"`
}

func check(e error) {
	if e != nil {
		fmt.Println("Error:", e.Error())
		log.Fatal(e)
		panic(e)
	}
}

func main() {
	settings = mysql.ConnectionURL{
		User:     `root`,
		Password: `jelszavam`,
		Database: `bookstore`,
		Host:     `db-bookstore`,
	}
	setupDB("setup.sql", "books.json")
	handleRequests()
	defer sess.Close()
}
