package main

import (
	"database/sql"
	"log"
	"net/http"
)

func init() {
	// データベースの初期化
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE posts (id INTEGER PRIMARY KEY, user_id INTEGER, title TEXT, body TEXT)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE comments (id INTEGER PRIMARY KEY, post_id INTEGER, body TEXT)`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			rows, err := db.Query(`SELECT id, name FROM users`)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
		} else if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			name := r.FormValue("name")
			_, err := db.Exec(`INSERT INTO users (name) VALUES (?)`, name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	})

	http.HandleFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			id := r.URL.Query().Get("id")
			if id == "" {
				http.Error(w, "id is required", http.StatusBadRequest)
				return
			}
			rows, err := db.Query(`SELECT id, name FROM users WHERE id = ?`, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
		} else if r.Method == http.MethodPut {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id := r.URL.Query().Get("id")
			name := r.FormValue("name")
			_, err := db.Exec(`UPDATE users SET name = ? WHERE id = ?`, name, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else if r.Method == http.MethodDelete {
			id := r.URL.Query().Get("id")
			_, err := db.Exec(`DELETE FROM users WHERE id = ?`, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	})

	http.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			rows, err := db.Query(`SELECT id, user_id, title, body FROM posts`)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
		} else if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			userID := r.FormValue("user_id")
			title := r.FormValue("title")
			body := r.FormValue("body")
			_, err := db.Exec(`INSERT INTO posts (user_id, title, body) VALUES (?, ?, ?)`, userID, title, body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	})

	http.HandleFunc("/posts/:id", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			id := r.URL.Query().Get("id")
			if id == "" {
				http.Error(w, "id is required", http.StatusBadRequest)
				return
			}
			rows, err := db.Query(`SELECT id, user_id, title, body FROM posts WHERE id = ?`, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
		} else if r.Method == http.MethodPut {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id := r.URL.Query().Get("id")
			userID := r.FormValue("user_id")
			title := r.FormValue("title")
			body := r.FormValue("body")
			_, err := db.Exec(`UPDATE posts SET user_id = ?, title = ?, body = ? WHERE id = ?`, userID, title, body, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else if r.Method == http.MethodDelete {
			id := r.URL.Query().Get("id")
			_, err := db.Exec(`DELETE FROM posts WHERE id = ?`, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	})

	http.HandleFunc("/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			rows, err := db.Query(`SELECT id, post_id, body FROM comments`)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
		} else if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			postID := r.FormValue("post_id")
			body := r.FormValue("body")
			_, err := db.Exec(`INSERT INTO comments (post_id, body) VALUES (?, ?)`, postID, body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	})

	http.HandleFunc("/comments/:id", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			id := r.URL.Query().Get("id")
			if id == "" {
				http.Error(w, "id is required", http.StatusBadRequest)
				return
			}
			rows, err := db.Query(`SELECT id, post_id, body FROM comments WHERE id = ?`, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
		} else if r.Method == http.MethodPut {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id := r.URL.Query().Get("id")
			body := r.FormValue("body")
			_, err := db.Exec(`UPDATE comments SET body = ? WHERE id = ?`, body, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else if r.Method == http.MethodDelete {
			id := r.URL.Query().Get("id")
			_, err := db.Exec(`DELETE FROM comments WHERE id = ?`, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	})

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
