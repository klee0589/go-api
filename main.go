package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func init() {
	// Open SQLite database
	var err error
	db, err = sql.Open("sqlite3", "./fitness.db")
	if err != nil {
		panic(err)
	}

	// Create workouts table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS workouts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			duration INTEGER
		)
	`)
	if err != nil {
		panic(err)
	}
}

type Workout struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Duration int    `json:"duration"`
}

func main() {
	router := gin.Default()

	// Get all workouts
	router.GET("/api/workouts", getWorkouts)

	// Get a specific workout by ID
	router.GET("/api/workouts/:id", getWorkoutByID)

	// Create a new workout
	router.POST("/api/workouts", createWorkout)

	router.Run(":8080")
}

func getWorkouts(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, duration FROM workouts")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	defer rows.Close()

	var workouts []Workout
	for rows.Next() {
		var w Workout
		if err := rows.Scan(&w.ID, &w.Name, &w.Duration); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		workouts = append(workouts, w)
	}

	c.JSON(http.StatusOK, workouts)
}

func getWorkoutByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workout ID"})
		return
	}

	var w Workout
	err = db.QueryRow("SELECT id, name, duration FROM workouts WHERE id = ?", id).Scan(&w.ID, &w.Name, &w.Duration)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workout not found"})
		return
	}

	c.JSON(http.StatusOK, w)
}

func createWorkout(c *gin.Context) {
	var newWorkout Workout
	if err := c.ShouldBindJSON(&newWorkout); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Insert the new workout into SQLite
	result, err := db.Exec("INSERT INTO workouts (name, duration) VALUES (?, ?)", newWorkout.Name, newWorkout.Duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// Set the ID of the new workout
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	newWorkout.ID = int(lastInsertID)

	c.JSON(http.StatusCreated, newWorkout)
}
