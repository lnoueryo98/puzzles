package models

import (
	"backend/modules/crypto"
	"backend/modules/mail"
	"backend/modules/session"
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/gorm"
)


type MailRequest struct {
	// リクエストで送られてきたメールアドレス
	Email			string 	`json:"email"`
	// 加入する組織のID
	OrganizationID	string	`json:"organization_id"`
	// 加入するユーザーの権限
	AuthorityID		int		`json:"authority_id"`
}

// ユーザーを組織に登録する関数。
// 組織に登録するユーザーの有無と、組織に既に登録されているかを確認し、
// 招待のメールを送る
func InviteUser(DB *gorm.DB, r *http.Request) error {

	err := DB.Transaction(func(tx *gorm.DB) error {

		cookie, err := r.Cookie("_cookie");if err != nil {
			return err
		}

		s, err := session.CheckSession(cookie.Value, project)
		if err != nil {
			return err
		}

		// メール送信用の情報をjsonから取得
		requestMail, _ := GetMailJson(r)

		// ユーザーが既に登録されているか確認。いない場合はユーザー作成
		var u User
		err = tx.FirstOrCreate(&u, User{Email: requestMail.Email}).Error; if err != nil {
			return err
		}

		// verification用のパスワード作成
		verification, _ := crypto.MakeRandomStr(30)
		var oa = OrganizationAuthority{
			UserID: u.ID,
			OrganizationID: requestMail.OrganizationID,
			AuthorityID: requestMail.AuthorityID,
			Active: false,
			Verification: verification,
		}

		// ユーザーに既に招待されているか確認
		condition := OrganizationAuthority{
			OrganizationID: oa.OrganizationID,
			UserID: oa.UserID,
		}
		var newOA OrganizationAuthority
		result := tx.Where(condition).First(&newOA); if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// まだ招待されていない場合
			result = tx.Create(&oa); if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return result.Error
			}
		} else {
			// 招待されており、既に組織に登録されている場合
			if newOA.Active {
				err = errors.New("既に組織に登録されているようです。")
				return err
			}

			// 招待されているが、ユーザーがまだ組織に登録されている場合
			newOA.Verification = oa.Verification
			result = tx.Save(&newOA); if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return result.Error
			}
		}

		// 招待メールの送信
		m := email
		m.Sub = s.Name + "さんから招待が送られました"
		m.To = requestMail.Email
		url := mainMessage + origin + "/register/verification?code=" + verification
		m.Message = url
		err = mail.SendEmail(m); if err !=nil {
			errorlog.Print(err)
			err = errors.New("存在しないメールアドレスです。")
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// JSONを構造体MailRequestに変換
func GetMailJson(r *http.Request) (MailRequest, error) {
	var mailRequest MailRequest
	json.NewDecoder(r.Body).Decode(&mailRequest)
	return mailRequest, nil
}

const (
	// メールのメッセージ
	mainMessage = "下記のURLより参加できます。\n"
)