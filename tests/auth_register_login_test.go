package tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	ssov1 "github.com/kerrek8/protos_sso1/gen/go/sso"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sso/tests/suite"
	"testing"
	"time"
)

const (
	appID          = 1
	appSecret      = "test-secret"
	passDefaultLen = 10
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)
	email := gofakeit.Email()
	password := randomPassword()
	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, respReg.UserId)

	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appID,
	})
	require.NoError(t, err)

	token := respLogin.Token
	require.NotEmpty(t, token)
	loginTime := time.Now()
	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, respReg.UserId, int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	const deltaSeconds = 1
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)
}

func TestRegisterLogin_DuplicatedRegistration(t *testing.T) {
	ctx, st := suite.New(t)
	email := gofakeit.Email()
	password := randomPassword()
	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	require.NotEmpty(t, respReg.UserId)

	respReg, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.Error(t, err)
	assert.Empty(t, respReg.UserId)
	assert.ErrorContains(t, err, "user already exists")
}

func TestRegister_FailCases(t *testing.T) {
	ctx, st := suite.New(t)
	tests := []struct {
		name          string
		email         string
		password      string
		expectedError string
	}{
		{name: "Register with Empty password",
			email:         gofakeit.Email(),
			password:      "",
			expectedError: "email and password are required",
		},
		{
			name:          "Register with Empty email",
			email:         "",
			password:      randomPassword(),
			expectedError: "email and password are required",
		},
		{
			name:          "Register with Empty email and password",
			email:         "",
			password:      "",
			expectedError: "email and password are required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedError)
			assert.Empty(t, respReg.UserId)

		})
	}

}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := suite.New(t)
	tests := []struct {
		name          string
		email         string
		password      string
		appID         int
		expectedError string
	}{
		{
			name:          "Login with Empty password",
			email:         gofakeit.Email(),
			password:      "",
			appID:         appID,
			expectedError: "email and password are required",
		},
		{
			name:          "Login with Empty email",
			email:         "",
			password:      randomPassword(),
			appID:         appID,
			expectedError: "email and password are required",
		},
		{
			name:          "Login with Empty email and password",
			email:         "",
			password:      "",
			appID:         appID,
			expectedError: "email and password are required",
		},
		{
			name:          "Login with Empty appID",
			email:         gofakeit.Email(),
			password:      randomPassword(),
			appID:         0,
			expectedError: "app_id is required",
		},
		{
			name:          "Login with Invalid password",
			email:         gofakeit.Email(),
			password:      randomPassword(),
			appID:         appID,
			expectedError: "invalid email or password",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    gofakeit.Email(),
				Password: randomPassword(),
			})
			require.NoError(t, err)

			_, err = st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
				AppId:    int32(tt.appID),
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func randomPassword() string {
	return gofakeit.Password(true, true, true, true, true, passDefaultLen)
}
