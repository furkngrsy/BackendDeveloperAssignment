package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

var db *sql.DB

func init() {
	var err error
	// Replace the connection string with your PostgreSQL connection details
	connStr := "user=postgres password = 123456 dbname=Task sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to the database")
}

// Create a new task
func createTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert task into the database
	err = db.QueryRow("INSERT INTO tasks(title, description, status) VALUES($1, $2, $3) RETURNING id",
		task.Title, task.Description, task.Status).Scan(&task.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// Get all tasks
func getAllTasks(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM tasks")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// Get a specific task by ID
func getTaskByID(w http.ResponseWriter, r *http.Request) {
	// Extract the task ID from the URL parameters
	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	var task Task
	err := db.QueryRow("SELECT * FROM tasks WHERE id = $1", taskID).
		Scan(&task.ID, &task.Title, &task.Description, &task.Status)
	if err == sql.ErrNoRows {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// Update an existing task by ID
func updateTaskByID(w http.ResponseWriter, r *http.Request) {
	// Extract the task ID from the URL parameters
	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	// Check if the task exists
	var existingTask Task
	err := db.QueryRow("SELECT * FROM tasks WHERE id = $1", taskID).
		Scan(&existingTask.ID, &existingTask.Title, &existingTask.Description, &existingTask.Status)
	if err == sql.ErrNoRows {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Decode the request body into a temporary task object
	var updatedTask Task
	err = json.NewDecoder(r.Body).Decode(&updatedTask)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update the existing task values with the new values
	existingTask.Title = updatedTask.Title
	existingTask.Description = updatedTask.Description
	existingTask.Status = updatedTask.Status

	// Perform the update
	_, err = db.Exec("UPDATE tasks SET title = $1, description = $2, status = $3 WHERE id = $4",
		existingTask.Title, existingTask.Description, existingTask.Status, taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingTask)
}

// Delete a task by ID
func deleteTaskByID(w http.ResponseWriter, r *http.Request) {
	// Extract the task ID from the URL parameters
	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	// Check if the task exists
	var existingTask Task
	err := db.QueryRow("SELECT * FROM tasks WHERE id = $1", taskID).
		Scan(&existingTask.ID, &existingTask.Title, &existingTask.Description, &existingTask.Status)
	if err == sql.ErrNoRows {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Perform the delete
	_, err = db.Exec("DELETE FROM tasks WHERE id = $1", taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, "Task deleted successfully ")
	json.NewEncoder(w).Encode(existingTask)
}

func main() {
	http.HandleFunc("/create-task", createTask)
	http.HandleFunc("/get-all-tasks", getAllTasks)
	http.HandleFunc("/get-task-by-id", getTaskByID)
	http.HandleFunc("/update-task-by-id", updateTaskByID)
	http.HandleFunc("/delete-task-by-id", deleteTaskByID)

	port := 8080
	log.Printf("Server is running on :%d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
