package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Post struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func main() {
	var err error

	db, err = sql.Open("mysql", "joshua468:Temi2080#@tcp(localhost:3306)/mydb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id INT AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			body TEXT NOT NULL
		);
	`); err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	r.GET("/posts", getPosts)
	r.GET("/posts/:id", getPost)
	r.POST("/posts", createPost)
	r.PUT("/posts/:id", updatePost)
	r.DELETE("/posts/:id", deletePost)

	r.Run(":8080")
}

func getPosts(c *gin.Context) {
	var posts []Post
	rows, err := db.Query("SELECT id, title, body FROM posts")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Body); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		posts = append(posts, post)
	}

	c.JSON(http.StatusOK, posts)
}

func getPost(c *gin.Context) {
	id := c.Param("id")
	var post Post
	row := db.QueryRow("SELECT id, title, body FROM posts WHERE id = ?", id)
	if err := row.Scan(&post.ID, &post.Title, &post.Body); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}
	c.JSON(http.StatusOK, post)
}

func createPost(c *gin.Context) {
	var post Post
	if err := c.BindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	_, err := db.Exec("INSERT INTO posts (title, body) VALUES (?, ?)", post.Title, post.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusCreated, post)
}

func updatePost(c *gin.Context) {
	id := c.Param("id")
	var existingPost Post
	row := db.QueryRow("SELECT id, title, body FROM posts WHERE id = ?", id)
	if err := row.Scan(&existingPost.ID, &existingPost.Title, &existingPost.Body); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var updatedPost Post
	if err := c.BindJSON(&updatedPost); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	_, err := db.Exec("UPDATE posts SET title = ?, body = ? WHERE id = ?", updatedPost.Title, updatedPost.Body, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	updatedPost.ID = existingPost.ID
	c.JSON(http.StatusOK, updatedPost)
}

func deletePost(c *gin.Context) {
	id := c.Param("id")

	result, err := db.Exec("DELETE FROM posts WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}
