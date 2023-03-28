package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "naha22"
	dbname   = "db-go-sql"
)

var (
	db  *sql.DB
	err error
)

type book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Desc   string `json:"desc"`
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected to database")

	r := gin.Default()

	r.GET("/books", getBook)
	r.POST("/books", createBook)
	r.PUT("/books/:id", updateBook)
	r.DELETE("/books/:id", deleteBook)

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func getBook(c *gin.Context) {
	var results []book

	rows, err := db.Query("SELECT * FROM book")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var Book book
		err := rows.Scan(&Book.ID, &Book.Title, &Book.Author, &Book.Desc)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, Book)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, results)
}

func createBook(c *gin.Context) {
	var Book book

	if err := c.ShouldBindJSON(&Book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sqlStatement := `
	INSERT INTO book ("title", "author", "desc")
	VALUES ($1, $2, $3)
	RETURNING ID`

	var id int
	err := db.QueryRow(sqlStatement, Book.Title, Book.Author, Book.Desc).Scan(&id)
	if err != nil {
		log.Fatal(err)
	}

	Book.ID = id

	c.JSON(http.StatusCreated, Book)
}

func updateBook(c *gin.Context) {
	id := c.Param("id")

	var Book book

	if err := c.ShouldBindJSON(&Book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sqlStatement := `
	UPDATE book
	SET "title" = $2, "author" = $3, "desc" = $4
	WHERE id = $1;`

	res, err := db.Exec(sqlStatement, id, Book.Title, Book.Author, Book.Desc)
	if err != nil {
		log.Fatal(err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}

	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No rows updated"})
		return
	}

}

func deleteBook(cok *gin.Context) {
	id := cok.Param("id")

	var Book book
	if err := cok.ShouldBindJSON(&Book); err != nil {
		cok.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return

	}

	sqlStatement := `
	DELETE FROM book
	WHERE id = $1`

	res, err := db.Exec(sqlStatement, id)
	if err != nil {
		log.Fatal(err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		cok.JSON(http.StatusNotFound, gin.H{
			"error": "Nothing deleted",
		})
		return
	}
	cok.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Id = %s berhasil dihapus cokk", id),
	})

}
