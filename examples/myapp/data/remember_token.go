package data

import (
	"time"

	up "github.com/upper/db/v4"
)

type RememberToken struct {
	ID            int       `json:"id,omitempty" db:"id,omitempty"`
	UserID        int       `json:"user_id" db:"user_id"`
	RememberToken string    `json:"remember_token" db:"remember_token"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

func (r *RememberToken) TableName() string {
	return "remember_tokens"
}

func (r *RememberToken) InsertToken(userID int, token string) (err error) {

	collection := upper.Collection(r.TableName())
	rememberToken := &RememberToken{
		UserID:        userID,
		RememberToken: token,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = collection.Insert(rememberToken)
	return
}

func (r *RememberToken) Delete(remember_token string) (err error) {
	collection := upper.Collection(r.TableName())
	res := collection.Find(up.Cond{"remember_token": remember_token})
	return res.Delete()
}
