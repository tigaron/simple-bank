package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/tigaron/simple-bank/db/mock"
	db "github.com/tigaron/simple-bank/db/sqlc"
	"github.com/tigaron/simple-bank/util"
)

func TestCreateAccountAPI(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		arg           db.CreateAccountParams
		buildStubs    func(store *mockdb.MockStore, arg db.CreateAccountParams)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			arg: db.CreateAccountParams{
				Owner:    account.Owner,
				Balance:  0,
				Currency: account.Currency,
			},
			buildStubs: func(store *mockdb.MockStore, arg db.CreateAccountParams) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name: "InternalError",
			arg: db.CreateAccountParams{
				Owner:    account.Owner,
				Balance:  0,
				Currency: account.Currency,
			},
			buildStubs: func(store *mockdb.MockStore, arg db.CreateAccountParams) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidBody",
			arg: db.CreateAccountParams{
				Owner:    account.Owner,
				Balance:  0,
				Currency: "JPY",
			},
			buildStubs: func(store *mockdb.MockStore, arg db.CreateAccountParams) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// build stubs
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store, tc.arg)

			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := "/accounts"
			request, err := http.NewRequest(http.MethodPost, url, createBody(tc.arg))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(t, recorder)
		})
	}

}

func TestGetAccountAPI(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// build stubs
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccountAPI(t *testing.T) {
	accounts := []db.Account{
		randomAccount(),
		randomAccount(),
		randomAccount(),
		randomAccount(),
		randomAccount(),
	}

	testCases := []struct {
		name          string
		params        listAccountRequest
		buildStubs    func(store *mockdb.MockStore, arg db.ListAccountsParams)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			params: listAccountRequest{
				PageID:   1,
				PageSize: 5,
			},
			buildStubs: func(store *mockdb.MockStore, arg db.ListAccountsParams) {
				store.EXPECT().
					ListAccounts(gomock.Any(), arg).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, accounts)
			},
		},
		{
			name: "InternalError",
			params: listAccountRequest{
				PageID:   1,
				PageSize: 5,
			},
			buildStubs: func(store *mockdb.MockStore, arg db.ListAccountsParams) {
				store.EXPECT().
					ListAccounts(gomock.Any(), arg).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidPageSize",
			params: listAccountRequest{
				PageID:   1,
				PageSize: 50,
			},
			buildStubs: func(store *mockdb.MockStore, arg db.ListAccountsParams) {
				store.EXPECT().
					ListAccounts(gomock.Any(), arg).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "EmptyResult",
			params: listAccountRequest{
				PageID:   2,
				PageSize: 5,
			},
			buildStubs: func(store *mockdb.MockStore, arg db.ListAccountsParams) {
				store.EXPECT().
					ListAccounts(gomock.Any(), arg).
					Times(1).
					Return([]db.Account{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// build stubs
			store := mockdb.NewMockStore(ctrl)
			arg := db.ListAccountsParams{
				Limit:  tc.params.PageSize,
				Offset: (tc.params.PageID - 1) * tc.params.PageSize,
			}
			tc.buildStubs(store, arg)

			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts?page_id=%d&page_size=%d", tc.params.PageID, tc.params.PageSize)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)

			// check response
			tc.checkResponse(t, recorder)
		})
	}
}

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	var gotAccount db.Account

	data, err := io.ReadAll(body)
	require.NoError(t, err)

	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, accounts []db.Account) {
	var gotAccounts []db.Account

	data, err := io.ReadAll(body)
	require.NoError(t, err)

	err = json.Unmarshal(data, &gotAccounts)
	require.NoError(t, err)
	require.Equal(t, accounts, gotAccounts)
}

func createBody(body interface{}) io.Reader {
	json, _ := json.Marshal(body)
	return bytes.NewReader(json)
}
