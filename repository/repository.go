package repository

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type User struct {
	ID    int
	Name  string
	Email string
}

type Bookmark struct {
	ID        int
	UserID    int
	Title     string
	URL       string
	CreatedAt time.Time
}

// ErrEmailTaken signals that the email is already taken.
var ErrEmailTaken = errors.New("email already in use")

// CreateUser inserts a new user, returning ErrEmailTaken if the email is already taken.
func CreateUser(db *sql.DB, name, email string) (*User, error) {
	// Check uniqueness
	var exists bool

	// QueryRow executes a query expected to return at most one row.
	row := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email)

	// Scan copies the columns from the matched row into the values pointed to by its arguments.
	if err := row.Scan(&exists); err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrEmailTaken
	}

	// Exec executes a query without returning any rows.
	// The '?' are placeholders for the parameters that follow the query string.
	res, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", name, email)
	if err != nil {
		return nil, err
	}

	// LastInsertId returns the integer ID of the last row inserted.
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &User{ID: int(id), Name: name, Email: email}, nil
}

// CreateBookmark inserts a new bookmark for a given user.
func CreateBookmark(db *sql.DB, userID int, title, url string) (*Bookmark, error) {
	res, err := db.Exec(
		"INSERT INTO bookmarks (user_id, title, url) VALUES (?, ?, ?)", userID, title, url,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	bm := &Bookmark{
		ID:     int(id),
		UserID: userID,
		Title:  title,
		URL:    url,
	}

	// fetch created_at
	err = db.QueryRow("SELECT created_at FROM bookmarks WHERE id = ?", id).Scan(&bm.CreatedAt)
	if err != nil {
		// if we can't get the timestamp, it's better to return the error
		// than a partially populated object.
		return nil, err
	}
	return bm, nil
}

// Helper func to add a bookmark.
func AddBookmark(db *sql.DB, reader *bufio.Reader, userID int) {
	fmt.Print("Title: ")
	title, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading title:", err)
		return
	}
	title = strings.TrimSpace(title)

	fmt.Print("URL: ")
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)

	bm, err := CreateBookmark(db, userID, title, url)
	if err != nil {
		fmt.Println("could not save bookmark: ", err)
		return
	}
	fmt.Printf("Saved: %s\n", bm.Title)
}

// FetchBookmarks retrieves all bookmarks for a user
func FetchBookmarks(db *sql.DB, userID int) ([]*Bookmark, error) {
	rows, err := db.Query("SELECT id, title, url, created_at FROM bookmarks WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Bookmark
	for rows.Next() {
		var bm Bookmark
		bm.UserID = userID
		if err := rows.Scan(&bm.ID, &bm.Title, &bm.URL, &bm.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &bm)
	}
	return list, nil
}

// Helper func to list bookmarks.
func ListBookmarks(db *sql.DB, userID int) {
	bms, err := FetchBookmarks(db, userID)
	if err != nil {
		fmt.Println("could not retrieve bookmarks:", err)
		return
	}

	if len(bms) == 0 {
		fmt.Println("Empty. You haven't added any bookmarks yet.")
		return
	}

	fmt.Println("\n--- Your bookmarks ---")
	for _, bm := range bms {
		fmt.Printf("\nTitle: %s\nURL: %s\nCreated At: %s\n", bm.Title, bm.URL, bm.CreatedAt.Format(time.RFC3339))
	}
}

// GetUserByEmail retrieves a user by their email address.
// It returns sql.ErrNoRows if no user is found.
func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	user := &User{Email: email}

	// since emails are unique, we only want one user, QueryRow is appropriate.
	// db.Query for multiple rows.
	err := db.QueryRow("SELECT id, name FROM users WHERE email = ?", email).Scan(&user.ID, &user.Name)
	if err != nil {
		return nil, err // sql.ErrNoRows is returned if no user is found
	}
	return user, nil
}

// DeleteBookmark is a helper func to prompt for a title and delete the bookmark.
func DeleteBookmark(db *sql.DB, reader *bufio.Reader, userID int) {
	fmt.Print("Enter the title of the bookmark to delete: ")
	title, _ := reader.ReadString('\n')
	title = strings.TrimSpace(title)

	rowsAffected, err := deleteBookmarkByTitle(db, userID, title)
	if err != nil {
		fmt.Printf("Could not delete bookmark: %v\n", err)
		return
	}
	if rowsAffected == 0 {
		fmt.Println("No bookmark found with this title.")
	} else {
		fmt.Printf("Bookmark '%s' deleted successfully.\n", title)
	}

}

// deleteBookmarkByTitle deletes a bookmark for a user given its title.
func deleteBookmarkByTitle(db *sql.DB, userID int, title string) (int64, error) {
	res, err := db.Exec("DELETE FROM bookmarks WHERE user_id = ? AND title = ?", userID, title)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// DeleteAccount is a helper func to handle account deletion.
func DeleteAccount(db *sql.DB, reader *bufio.Reader, user *User) bool {
	fmt.Printf("Are you sure you want to delete your account, %s? This action cannot be undone. (y/n): ", user.Name)
	confirmation, _ := reader.ReadString('\n')
	confirmation = strings.TrimSpace(confirmation)
	confirmation = strings.ToLower(confirmation)
	if confirmation != "y" && confirmation != "yes" {
		fmt.Println("Account deletion cancelled.")
		return false
	} else if confirmation == "y" || confirmation == "yes" {
		fmt.Print("Type your email to proceed deletion process: ")
		confirmEmail, _ := reader.ReadString('\n')
		confirmEmail = strings.TrimSpace(confirmEmail)
		if confirmEmail != user.Email {
			fmt.Println("Email does not match. Aborting deletion.")
			return false
		}

		err := deleteUserAndBookmarks(db, user.ID)
		if err != nil {
			fmt.Printf("Failed to delete account: %v\n", err)
			return false
		}

		fmt.Println("Your account and all associated bookmarks have been deleted successfully.")
		return true
	}
	// Default case: if none of the above conditions are met, return false
	return false
}

// deleteUserAndBookmarks deletes a user and all their bookmarks in a transaction.
func deleteUserAndBookmarks(db *sql.DB, userID int) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not start the transaction: %w", err)
	}

	// defer a rollback in case anything fails.
	// it will be ignored if the transaction is committed.
	defer tx.Rollback()

	// Delete bookmarks first
	_, err = tx.Exec("DELETE FROM bookmarks WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("could not delete bookmarks: %w", err)
	}

	// delete the user
	_, err = tx.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		return fmt.Errorf("could not delete user: %w", err)
	}

	// Commit the transaction
	return tx.Commit()
}
