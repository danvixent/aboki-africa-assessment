package tests

import (
	"context"
	"encoding/json"
	"fmt"
	app "github.com/danvixent/aboki-africa-assessment"
	"github.com/danvixent/aboki-africa-assessment/handler"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestRegisterUser(t *testing.T) {
	err := deleteAllFromTable("user_points")
	if !assert.NoError(t, err) {
		return
	}

	err = deleteAllFromTable("users")
	if !assert.NoError(t, err) {
		return
	}

	user1, err := seedOneUser("Daniel", "dan@gmail.com")
	if !assert.NoError(t, err) {
		return
	}

	_, err = seedPointBalanceForUser(user1.ID, 0)
	if !assert.NoError(t, err) {
		return
	}

	tests := []struct {
		requestBody *handler.UserRequest
		wantCode    int
		checkData   bool
	}{
		{
			requestBody: &handler.UserRequest{
				Name:         "Daniel",
				Email:        "daniel@gmail.com",
				ReferralCode: &user1.ReferralCode,
			},
			checkData: true,
			wantCode:  http.StatusOK,
		},
		{
			requestBody: &handler.UserRequest{
				Name:         "Dave",
				Email:        "daniel1@gmail.com",
				ReferralCode: &user1.ReferralCode,
			},
			wantCode: http.StatusOK,
		},
		{
			requestBody: &handler.UserRequest{
				Name:         "West",
				Email:        "daniel2@gmail.com",
				ReferralCode: &user1.ReferralCode,
			},
			checkData: true,
			wantCode:  http.StatusOK,
		},
		{
			requestBody: &handler.UserRequest{
				Name:         "",
				Email:        "daniel1@gmail.com",
				ReferralCode: &user1.ReferralCode,
			},
			checkData: false,
			wantCode:  http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		resp, err := registerUser(test.requestBody)
		if !assert.NoError(t, err) {
			return
		}

		if assert.Equal(t, test.wantCode, resp.StatusCode) {
			return
		}

		if test.checkData {
			body := &app.User{}
			err = getResponseBody(resp.Body, body)
			if !assert.NoError(t, err) {
				return
			}

			assert.NotEmpty(t, body.ID)
			assert.Equal(t, test.requestBody.Name, body.Name)
			assert.Equal(t, test.requestBody.Email, body.Email)
			assert.NotEmpty(t, body.CreatedAt)
			assert.NotEmpty(t, body.UpdatedAt)
			assert.Nil(t, body.DeletedAt)
		}
	}

	pp, err := testHandler.userPointRepository.GetUserPointsBalance(context.Background(), user1.ID)
	if !assert.NoError(t, err) {
		return
	}

	assert.EqualValues(t, 50, pp)
}

func TestTransaction(t *testing.T) {
	err := deleteAllFromTable("user_points")
	if !assert.NoError(t, err) {
		return
	}

	err = deleteAllFromTable("transactions")
	if !assert.NoError(t, err) {
		return
	}

	err = deleteAllFromTable("user_referrals")
	if !assert.NoError(t, err) {
		return
	}

	err = deleteAllFromTable("users")
	if !assert.NoError(t, err) {
		return
	}

	user1, err := seedOneUser("Daniel", "dan@gmail.com")
	if !assert.NoError(t, err) {
		return
	}

	_, err = seedPointBalanceForUser(user1.ID, 4000)
	if !assert.NoError(t, err) {
		return
	}

	tests := []struct {
		user      *app.User
		userPoint *app.UserPoints
		wantCode  int
	}{
		{
			user: &app.User{
				Name:         "Daniel",
				Email:        "da@gmail.com",
				ReferralCode: handler.GenReferralCode(6),
			},
			userPoint: &app.UserPoints{
				Points: 40000,
			},
			wantCode: http.StatusOK,
		},
		{
			user: &app.User{
				Name:         "Adewale",
				Email:        "danny@gmail.com",
				ReferralCode: handler.GenReferralCode(6),
			},
			userPoint: &app.UserPoints{
				Points: 40000,
			},
			wantCode: http.StatusOK,
		},
		{
			user: &app.User{
				Name:         "Fiona",
				Email:        "danboy@gmail.com",
				ReferralCode: handler.GenReferralCode(6),
			},
			userPoint: &app.UserPoints{
				Points: 40000,
			},
			wantCode: http.StatusOK,
		},
	}

	ctx := context.Background()
	for _, test := range tests {
		err := testHandler.userRepository.CreateUser(ctx, test.user)
		if !assert.NoError(t, err) {
			return
		}

		test.userPoint.UserID = test.user.ID
		err = testHandler.userPointRepository.CreateUserPoint(ctx, test.userPoint)
		if !assert.NoError(t, err) {
			return
		}

		resp, err := transaction(&handler.TransferPointsRequest{
			UserID:          test.user.ID,
			RecipientUserID: user1.ID,
			Points:          210,
		})

		if !assert.NoError(t, err) {
			return
		}

		if assert.Equal(t, test.wantCode, resp.StatusCode) {
			return
		}
	}

	pp, err := testHandler.userPointRepository.GetUserPointsBalance(context.Background(), user1.ID)
	if !assert.NoError(t, err) {
		return
	}

	assert.EqualValues(t, 50, pp)
}

func getResponseBody(respBody io.ReadCloser, data interface{}) error {
	buf, err := ioutil.ReadAll(respBody)
	if err != nil {
		return err
	}

	fmt.Printf("response body: %+v", string(buf))
	if err = json.Unmarshal(buf, data); err != nil {
		return err
	}

	return nil
}

func registerUser(req *handler.UserRequest) (*http.Response, error) {
	return http.Post(url+"/register", "application/json", serialize(req))
}

func transaction(req *handler.TransferPointsRequest) (*http.Response, error) {
	return http.Post(url+"/transaction", "application/json", serialize(req))
}

func seedOneUser(name string, email string) (*app.User, error) {
	user := &app.User{
		Name:         name,
		Email:        email,
		ReferralCode: handler.GenReferralCode(6),
	}

	err := testHandler.userRepository.CreateUser(context.Background(), user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func seedPointBalanceForUser(userID string, points int64) (*app.UserPoints, error) {
	p := &app.UserPoints{
		UserID: userID,
		Points: points,
	}
	err := testHandler.userPointRepository.CreateUserPoint(context.Background(), p)
	if err != nil {
		return nil, err
	}
	return p, nil
}
