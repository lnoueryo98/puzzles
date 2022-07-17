package models

import (
	"backend/modules/crypto"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)



type Login struct {
	Organization	string    `json:"organization"`
	Email			string    `json:"email"`
	Password		string    `json:"password"`
}

// ログインのバリデーション
func TryToLogin(DB *gorm.DB, w http.ResponseWriter, r *http.Request) (*User, error) {

	var u *User
	l, err := getLoginJson(r)
	if err != nil {
		return u, err
	}

	// ログインフォーム空白確認
	err = l.CheckLoginFormBlank()
	if err != nil {
		return u, err
	}

	// メールアドレスのフォームチェック
	err = CheckEmailFormat(l.Email)
	if err != nil {
		return u, err
	}

	// 全ての項目を踏まえログイン情報が正しい確認
	u, err = l.FindUser(DB, u)
	if err != nil {
		return u, err
	}

	return u, nil
}

func (l *Login) CheckLoginFormBlank() error {

	if l.Organization == "" {
		message := "organization is blank"
		err := errors.New(message)
		return err
	}

	if l.Email == "" {
		message := "email address is blank"
		err := errors.New(message)
		return err
	}

	if l.Password == "" {
		message := "password is blank"
		err := errors.New(message)
		return err
	}
	return nil
}

func (l *Login) FindUser(DB *gorm.DB, u *User) (*User, error) {
	// Password check
	cryptoPassword := crypto.Encrypt(l.Password)
	result := DB.Preload("Organizations", "organization_id = ?", l.Organization).Preload(clause.Associations).First(&u, "email = ? and password = ?", l.Email, cryptoPassword)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		message := "email or password is wrong"
		err := errors.New(message)
		return u, err
	}

	if len(u.Organizations) == 0 {
		message := "organization is wrong"
		err := errors.New(message)
		return u, err
	}
	return u, nil
}

func getLoginJson(r *http.Request) (Login, error) {
	var login Login
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return login, err
	}

	if len(body) > 0 {
		err = json.Unmarshal(body, &login)
		if err != nil {
			return login, err
		}
		return login, nil
	}

	message := "request body is empty"
	err = errors.New(message)
	return login, err
}