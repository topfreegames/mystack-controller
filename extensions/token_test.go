package extensions_test

import (
	"fmt"
	"time"

	"golang.org/x/oauth2"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/topfreegames/mystack-controller/extensions"
)

var _ = Describe("Token", func() {
	AfterEach(func() {
		err = mock.ExpectationsWereMet()
		Expect(err).NotTo(HaveOccurred())
	})

	var (
		accessToken  = "my_access_token"
		refreshToken = "my_refresh_token"
		expiry       = time.Unix(0, 0)
		tokenType    = "my_token_type"
		email        = "user@example.com"
	)

	Describe("SaveToken", func() {
		It("should save valid token and email", func() {
			mock.
				ExpectExec(`^INSERT INTO tokens\(access_token, refresh_token, expiry, token_type, email\)
					VALUES\((.+)\)
					ON CONFLICT\(email\) DO UPDATE
						SET access_token = excluded.access_token,
								refresh_token = excluded.refresh_token,
								expiry = excluded.expiry;$`).
				WithArgs(accessToken, refreshToken, expiry, tokenType, email).
				WillReturnResult(sqlmock.NewResult(1, 1))

			token := &oauth2.Token{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				Expiry:       expiry,
				TokenType:    tokenType,
			}

			err := SaveToken(token, email, sqlxDB)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should update access token and expiry if refresh token is empty", func() {
			newAccessToken := "new_access_token"
			newRefreshToken := ""
			newExpiry := time.Unix(100, 0)

			mock.
				ExpectExec(`^UPDATE tokens 
				SET access_token = (.+),
						expiry = (.+)
				WHERE email = (.+)$`).
				WithArgs(newAccessToken, newExpiry, email).
				WillReturnResult(sqlmock.NewResult(1, 1))

			token := &oauth2.Token{
				AccessToken:  newAccessToken,
				RefreshToken: newRefreshToken,
				Expiry:       newExpiry,
				TokenType:    tokenType,
			}

			err := SaveToken(token, email, sqlxDB)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error if email not found", func() {
			mock.
				ExpectExec(`^INSERT INTO tokens\(access_token, refresh_token, expiry, token_type, email\)
					VALUES\((.+)\)
					ON CONFLICT\(email\) DO UPDATE
						SET access_token = excluded.access_token,
								refresh_token = excluded.refresh_token,
								expiry = excluded.expiry;$`).
				WithArgs(accessToken, refreshToken, expiry, tokenType, email).
				WillReturnError(fmt.Errorf("sql: no rows in result set"))

			token := &oauth2.Token{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				Expiry:       expiry,
				TokenType:    tokenType,
			}

			err := SaveToken(token, email, sqlxDB)
			Expect(err).To(HaveOccurred())
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.DatabaseError"))
		})
	})

	Describe("Token", func() {
		It("should return correct token", func() {
			rows := sqlmock.NewRows(
				[]string{"access_token", "refresh_token", "expiry", "token_type"},
			).AddRow(accessToken, refreshToken, expiry, tokenType)
			mock.
				ExpectQuery(`
					^SELECT access_token, refresh_token, expiry, token_type
					FROM tokens
					WHERE access_token = (.+)$`).
				WithArgs(accessToken).
				WillReturnRows(rows)

			token, err := Token(accessToken, sqlxDB)
			Expect(err).NotTo(HaveOccurred())
			Expect(token.AccessToken).To(Equal(accessToken))
			Expect(token.RefreshToken).To(Equal(refreshToken))
			Expect(token.Expiry).To(Equal(expiry))
			Expect(token.TokenType).To(Equal(tokenType))
		})

		It("should return error if token doesn't exist", func() {
			mock.
				ExpectQuery(`
					^SELECT access_token, refresh_token, expiry, token_type
					FROM tokens
					WHERE access_token = (.+)$`).
				WithArgs(accessToken).
				WillReturnError(fmt.Errorf("sql: no rows in result set"))

			_, err := Token(accessToken, sqlxDB)
			Expect(err).To(HaveOccurred())
			Expect(fmt.Sprintf("%T", err)).To(Equal("*errors.AccessError"))
		})
	})
})
