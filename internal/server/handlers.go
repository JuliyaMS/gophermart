package server

import (
	"github.com/JuliyaMS/gophermart/internal/json"
	"github.com/JuliyaMS/gophermart/internal/storage"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type Handlers struct {
	dataStore      storage.Storager
	loggerHandlers *zap.SugaredLogger
}

func NewHandlers(logger *zap.SugaredLogger) *Handlers {

	storageDB := storage.NewConnectionDB(logger)

	err := storageDB.CheckConnection()
	if err != nil {
		logger.Error("Get error while create tables:", err.Error())
		return nil
	}

	err = storageDB.Init()
	if err != nil {
		logger.Error("Get error while create tables:", err.Error())
		return nil
	}

	return &Handlers{dataStore: storageDB, loggerHandlers: logger}
}

func (h *Handlers) createCookie(value, path string) http.Cookie {
	cookie := http.Cookie{
		Name:     "UserAuthentication",
		Value:    value,
		Path:     path,
		HttpOnly: true,
		Secure:   true,
	}
	return cookie
}

func (h *Handlers) registration(w http.ResponseWriter, r *http.Request) {

	h.loggerHandlers.Infow("Start handler: registration")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var auth json.AuthData

	h.loggerHandlers.Infow("Decode body")
	if err := json.Decode(&auth, r.Body); err != nil {
		h.loggerHandlers.Error("Get error while decode data: ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.loggerHandlers.Infow("Start function CheckUser")
	if err := h.dataStore.CheckUser(auth.Login); err == nil {
		h.loggerHandlers.Infow("Login is already exist")
		w.WriteHeader(http.StatusConflict)
		return
	}

	h.loggerHandlers.Infow("Start function AddUser")
	if err := h.dataStore.AddUser(auth.Login, auth.Password); err != nil {
		h.loggerHandlers.Error("Get error in function AddUser :", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.loggerHandlers.Infow("Set cookie")
	cookie := h.createCookie(auth.Login, "/api/user/register")
	http.SetCookie(w, &cookie)

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) login(w http.ResponseWriter, r *http.Request) {

	h.loggerHandlers.Infow("Start handler: login")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var auth json.AuthData

	h.loggerHandlers.Infow("Decode body")
	if err := json.Decode(&auth, r.Body); err != nil {
		h.loggerHandlers.Error("Get error while decode data: ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.loggerHandlers.Infow("Start function CheckUser")
	if err := h.dataStore.CheckUser(auth.Login); err != nil {
		h.loggerHandlers.Info("User: ", auth.Login, " doesn't exist")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	h.loggerHandlers.Infow("Start function CheckPassword")
	password, err := h.dataStore.CheckPassword(auth.Login)

	if err != nil {
		h.loggerHandlers.Error("Get error in function CheckPassword: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.loggerHandlers.Infow("Check password...")
	if password != auth.Password {
		h.loggerHandlers.Infow("Password is not correct!")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	h.loggerHandlers.Infow("Password is correct")

	h.loggerHandlers.Infow("Set cookie")
	cookie := h.createCookie(auth.Login, "/api/user/login")
	http.SetCookie(w, &cookie)

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) loadOrders(w http.ResponseWriter, r *http.Request) {

	h.loggerHandlers.Infow("Start handler: loadOrders")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	h.loggerHandlers.Infow("loadOrders: Get cookie...")
	cookie, err := r.Cookie("UserAuthentication")
	if err != nil {
		h.loggerHandlers.Infow("User not authenticated")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	h.loggerHandlers.Infow("Start function CheckUser")
	if err = h.dataStore.CheckUser(cookie.Value); err != nil {
		h.loggerHandlers.Info("User: ", cookie.Value, " doesn't exist")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	h.loggerHandlers.Infow("Get order`s number")
	var body []byte
	defer r.Body.Close()

	if body, err = io.ReadAll(r.Body); err != nil {
		h.loggerHandlers.Error("Get error while read data: ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	order := string(body)

	if !AlgorithmLuna(order) {
		h.loggerHandlers.Infow("Format of order`s number is not correct")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	h.loggerHandlers.Infow("Start function CheckOrder")
	login, err := h.dataStore.CheckOrder(order)
	if err == nil {
		if login == cookie.Value {
			h.loggerHandlers.Infow("This order already uploaded this user")
			w.WriteHeader(http.StatusOK)
			return
		} else {
			h.loggerHandlers.Infow("This order already uploaded another user")
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	h.loggerHandlers.Infow("Start function AddOrder")
	if err = h.dataStore.AddOrder(cookie.Value, order); err != nil {
		h.loggerHandlers.Error("Get error in function AddOrder :", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) getOrders(w http.ResponseWriter, r *http.Request) {

	h.loggerHandlers.Infow("Start handler: getOrders")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	h.loggerHandlers.Infow("getOrders: Get cookie...")
	cookie, err := r.Cookie("UserAuthentication")
	if err != nil {
		h.loggerHandlers.Infow("User not unauthorized")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	h.loggerHandlers.Infow("Start function CheckUser")
	if err = h.dataStore.CheckUser(cookie.Value); err != nil {
		h.loggerHandlers.Info("User: ", cookie.Value, " doesn't exist")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	h.loggerHandlers.Infow("Start function GetOrders")
	orders, err := h.dataStore.GetOrders(cookie.Value)
	if err != nil {
		h.loggerHandlers.Error("Get error while execute function: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		h.loggerHandlers.Infow("This user does not have orders")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.loggerHandlers.Infow("Encode data orders")
	err = json.Encode(orders, w)
	if err != nil {
		h.loggerHandlers.Error("Get error while encode data: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) getBalance(w http.ResponseWriter, r *http.Request) {

	h.loggerHandlers.Infow("Start handler: getBalance")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	h.loggerHandlers.Infow("getBalance: Get cookie...")
	cookie, err := r.Cookie("UserAuthentication")
	if err != nil {
		h.loggerHandlers.Infow("User not unauthorized")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	h.loggerHandlers.Infow("Start function CheckUser")
	if err = h.dataStore.CheckUser(cookie.Value); err != nil {
		h.loggerHandlers.Info("User: ", cookie.Value, " doesn't exist")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	h.loggerHandlers.Infow("Start function GetBalance")
	balance, err := h.dataStore.GetBalance(cookie.Value)
	if err != nil {
		h.loggerHandlers.Error("Get error while execute function: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.loggerHandlers.Infow("Encode data orders")
	err = json.Encode(balance, w)
	if err != nil {
		h.loggerHandlers.Error("Get error while encode data: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) balanceWithdraw(w http.ResponseWriter, r *http.Request) {
	h.loggerHandlers.Infow("Start handler: balanceWithdraw")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	h.loggerHandlers.Infow("balanceWithdraw: Get cookie...")
	cookie, err := r.Cookie("UserAuthentication")
	if err != nil {
		h.loggerHandlers.Infow("User not unauthorized")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	h.loggerHandlers.Infow("Start function CheckUser")
	if err = h.dataStore.CheckUser(cookie.Value); err != nil {
		h.loggerHandlers.Info("User: ", cookie.Value, " doesn't exist")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var wr json.Withdrawal

	h.loggerHandlers.Infow("Decode body")
	if err = json.Decode(&wr, r.Body); err != nil {
		h.loggerHandlers.Error("Get error while decode data: ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !AlgorithmLuna(wr.Order) {
		h.loggerHandlers.Infow("Format of order`s number is not correct")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	h.loggerHandlers.Infow("Start function GetBalance")
	balance, err := h.dataStore.GetBalance(cookie.Value)
	if err != nil {
		h.loggerHandlers.Error("Get error while execute function: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if balance.Current < wr.Sum {
		h.loggerHandlers.Infow("Insufficient funds")
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}

	h.loggerHandlers.Infow("Start function AddWithdraw")
	if err = h.dataStore.AddWithdraw(cookie.Value, wr.Order, wr.Sum); err != nil {
		h.loggerHandlers.Error("Get error while execute function: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) infoWithdraw(w http.ResponseWriter, r *http.Request) {
	h.loggerHandlers.Infow("Start handler: infoWithdraw")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	h.loggerHandlers.Infow("infoWithdraw: Get cookie...")
	cookie, err := r.Cookie("UserAuthentication")
	if err != nil {
		h.loggerHandlers.Infow("User not unauthorized")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	h.loggerHandlers.Infow("Start function CheckUser")
	if err = h.dataStore.CheckUser(cookie.Value); err != nil {
		h.loggerHandlers.Info("User: ", cookie.Value, " doesn't exist")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	h.loggerHandlers.Infow("Start function GetWithdraws")
	withdraws, err := h.dataStore.GetWithdraws(cookie.Value)
	if err != nil {
		h.loggerHandlers.Error("Get error while execute function: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(withdraws) == 0 {
		h.loggerHandlers.Infow("This user does not have withdraws")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.loggerHandlers.Infow("Encode data orders")
	err = json.Encode(withdraws, w)
	if err != nil {
		h.loggerHandlers.Error("Get error while encode data: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
