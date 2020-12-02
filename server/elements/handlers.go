package elements

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type integerKey int

var issuer = "ForcePoint"

const (
	productKey integerKey = iota
	Jti
)
const signature = "8IDHO65XH3GIO7RC4A6NLXND19B2A2Z19BN85YIOFWOCMX95G3N97YKD7CZN" +
	"558RSICXWYS5VENPB1GAID8QNICSIC57WQSCIXPJXNA4131JRX92LL77D74TNK9V0PB322Z5C3V3KF28E" +
	"FO5PT9OPNQY9MSLQASTD73JEDW7F617G8VWN7SQI1CH0Z0EQW4LTZXZ9ZBK9EQM98JB43TX09ZBW2VLQD" +
	"M9GWZYFHNYYQNJKHXRLND19F58IZJQM5F91HZVNQOHT1MZDHCOCXCMINN3M0X3W6JVEPEQ1O1BJ0WCASY" +
	"J0WDDYOK6XPZHOCQ04WRJ18L08MZC8NZZLZR4IAWV5IAPFK14XPFXBT7MLMZ8DLYYLJQ9BJJC0OWZIP4L" +
	"IAM6JCRAPX48C8D5WSWHRB0N77P8JNC9YDWTHVYWO4Q7V1W78G2NPZJ2J2FTV03GB1G8W8R53KHEBNT37" +
	"A0LVBPIRDC3027RA2CNARHIAAG2K0DE3W4TOXS1NHTCY6AJ"

type ErrorTemplate struct {
	ErrorCode int    `json:"error_code"`
	Message   string `json:"message"`
}

var ProductConnector Connector

func init() {
	ProductConnector = &ForcepointConnector{}
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("******************")
	//fmt.Println(r.Method, r.RequestURI)
	//b, _ := ioutil.ReadAll(r.Body)
	//fmt.Println(string(b))
	//fmt.Println("******************")
	//add data to response
	user := NewUser()
	args := r.URL.Query()
	count, startIndex := GetCountAndStartIndex(args)
	if filterQuery, ok := args["filter"]; ok {
		regx, err := regexp.Compile(`(\w+) eq "([^"]*)"`)
		if err != nil {
			loggerWithField(r).Fatal("Failed in defining the regex Compile")
		} else {
			if regx.Match([]byte(filterQuery[0])) {
				result := regx.FindStringSubmatch(filterQuery[0])[1:]
				//searchKey = result[0]
				searchKeyValue := result[1]
				user.Id = searchKeyValue
			}
		}

	} else {
		user.Id = ""
	}
	users, err := user.GetUsers(ProductConnector)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(users) == 0 {
		message := fmt.Sprintf("User %s not found", user.Id)
		w.Header().Set("Content-Type", "application/json")
		userNotFoundError := GenerateScimError(http.StatusNotFound, message)
		err := json.NewEncoder(w).Encode(userNotFoundError)
		if err != nil {
			loggerWithField(r).Fatal(err.Error())
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	scimSource := ScimSourceUsers(users, startIndex, count)
	err = json.NewEncoder(w).Encode(scimSource)
	if err != nil {
		loggerWithField(r).Fatal(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	loggerWithField(r).Info("get Users request")
}

func GetUserById(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("******************")
	//fmt.Println(r.Method, r.RequestURI)
	//b, _ := ioutil.ReadAll(r.Body)
	//fmt.Println(string(b))
	//fmt.Println("******************")
	vars := mux.Vars(r)
	userObj := NewUser()
	userObj.Id = vars["id"]
	users, err := userObj.GetUsers(ProductConnector)
	if err != nil {
		loggerWithField(r).Fatal(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if len(users) == 0 {
		message := fmt.Sprintf("User %s not found", vars["id"])
		userNotFoundError := GenerateScimError(http.StatusNotFound, message)
		err := json.NewEncoder(w).Encode(userNotFoundError)
		if err != nil {
			loggerWithField(r).Fatal(err.Error())
		}
		return
	}
	err = json.NewEncoder(w).Encode(users[0])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
	loggerWithField(r).Info("get UsersById request")
}

func PostUser(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("**************************************************")
	//b, _ := ioutil.ReadAll(r.Body)
	//fmt.Println(string(b))
	//fmt.Println("**************************************************")
	//rs := bytes.NewReader(b)

	defer r.Body.Close()
	var user UserObj
	w.Header().Set("Content-Type", "application/json")
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		loggerWithField(r).Fatal(err.Error())
	}
	createdUser, statusCode, err := user.PostUser(ProductConnector)
	if statusCode == http.StatusBadRequest {
		w.WriteHeader(statusCode)
		return
	}
	if statusCode == http.StatusUnprocessableEntity {
		message := fmt.Sprintf("User %s already exists", user.UserName)
		userNotFoundError := GenerateScimError(http.StatusConflict, message)
		err := json.NewEncoder(w).Encode(userNotFoundError)
		if err != nil {
			loggerWithField(r).Fatal(err.Error())
		}
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	createdUser.Schemas = createdUser.Schemas[:1]
	fmt.Printf("Created user: %+v\n", createdUser.UserName)
	err = json.NewEncoder(w).Encode(&createdUser)
	if err != nil {
		loggerWithField(r).Fatal(err.Error())
	}
	//fmt.Printf("%#v\n", createdUser)
	//b, _ := ioutil.ReadAll(r.Body)
	//fmt.Println(string(b))
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("**************************************************")
	//b, _ := ioutil.ReadAll(r.Body)
	//fmt.Println(string(b))
	//fmt.Println("**************************************************")
	//rs := bytes.NewReader(b)
	var user UserObj
	vars := mux.Vars(r)
	userId := vars["id"]
	var updateOp UpdateOperation
	err := json.NewDecoder(r.Body).Decode(&updateOp)
	if err != nil {
		loggerWithField(r).Fatal(err.Error())
	}
	updateOp.Id = userId
	updatedUser, err := user.UpdateUser(ProductConnector, updateOp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(updatedUser)
	if err != nil {
		loggerWithField(r).Fatal(err.Error())
	}
	loggerWithField(r).Infof("updated user: %s", userId)
}

func Root(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Forcepoint SMC SCIM API"))
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	var user UserObj
	vars := mux.Vars(r)
	user.UserName = vars["id"]
	err := user.DeleteUser(ProductConnector)
	if err != nil {
		loggerWithField(r).Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	loggerWithField(r).Infof("Delete user: %s", user.UserName)

}

func GetGroups(w http.ResponseWriter, r *http.Request) {
	group := NewGroup()
	args := r.URL.Query()
	count, startIndex := GetCountAndStartIndex(args)
	if filterQuery, ok := args["filter"]; ok {
		regx, err := regexp.Compile(`(\w+) eq "([^"]*)"`)
		if err != nil {
			loggerWithField(r).Fatal("Failed in defining regex Compile")
		} else {
			if regx.Match([]byte(filterQuery[0])) {
				result := regx.FindStringSubmatch(filterQuery[0])[1:]
				displayName := result[1]
				group.DisplayName = displayName
			}
		}
	} else {
		group.DisplayName = ""
	}
	groups, err := group.GetGroups(ProductConnector)
	w.Header().Set("Content-Type", "application/json")
	if len(groups) == 0 {
		message := fmt.Sprintf("User %s not found", group.DisplayName)
		userNotFoundError := GenerateScimError(http.StatusNotFound, message)
		err := json.NewEncoder(w).Encode(userNotFoundError)
		if err != nil {
			loggerWithField(r).Fatal(err.Error())
		}
		return
	}
	scimSource := ScimSourceGroups(groups, startIndex, count)
	err = json.NewEncoder(w).Encode(scimSource)
	if err != nil {
		loggerWithField(r).Fatal(err.Error())
	}

}

func GetGroupById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	group := NewGroup()
	group.DisplayName = vars["id"]
	groups, err := group.GetGroups(ProductConnector)
	if err != nil {
		loggerWithField(r).Fatal(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	if len(groups) == 0 {
		message := fmt.Sprintf("User %s not found", group.DisplayName)
		notFoundError := GenerateScimError(http.StatusNotFound, message)
		err := json.NewEncoder(w).Encode(notFoundError)
		if err != nil {
			loggerWithField(r).Fatal(err.Error())
		}
		return
	}
	err = json.NewEncoder(w).Encode(groups)
	if err != nil {
		loggerWithField(r).Fatal(err.Error())
	}
}

func CreateGroup(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	group := NewGroup()
	w.Header().Set("Content-Type", "application/json")
	err := json.NewDecoder(r.Body).Decode(&group)
	if err != nil {
		loggerWithField(r).Fatal(err.Error())
	}
	group.Schemas = []string{"urn:ietf:params:scim:schemas:core:2.0:Group"}
	createdGroup, err := group.CreateGroup(ProductConnector)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(createdGroup)
	if err != nil {
		loggerWithField(r).Fatal(err.Error())
	}
	fmt.Printf("%#v\n", createdGroup)
	//b, _ := ioutil.ReadAll(r.Body)
	//fmt.Println(string(b))

}

func UpdateGroup(w http.ResponseWriter, r *http.Request) {
	group := NewGroup()
	vars := mux.Vars(r)
	groupId := vars["id"]
	group.ID = groupId
	var updateOp UpdateOperation
	err := json.NewDecoder(r.Body).Decode(&updateOp)
	if err != nil {
		loggerWithField(r).Fatal(err.Error())
	}
	updateOp.Id = groupId
	updatedGroup, err := group.UpdateGroup(ProductConnector, updateOp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(updatedGroup)
	if err != nil {
		loggerWithField(r).Fatal(err.Error())
	}
	fmt.Printf("%+v\n", updatedGroup)
}

func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	group := NewGroup()
	vars := mux.Vars(r)
	group.ID = vars["id"]
	err := group.DeleteGroup(ProductConnector)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func LogToScreen(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{
			"RequestMethod": r.Method, "RequestURL": r.RequestURI, "RemoteAddress": r.RemoteAddr,
		}).Info("HTTP Request")
		f(w, r)
	}
}

func GetToken(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusGatewayTimeout)
			errInfo := ErrorTemplate{
				ErrorCode: http.StatusGatewayTimeout,
				Message:   fmt.Sprintf("Error: %s, Please ensure the product IP address and key are currect", err),
			}
			loggerWithField(r).Fatal(errInfo.Message)
			err := json.NewEncoder(w).Encode(errInfo)
			if err != nil {
				loggerWithField(r).Warning(err.Error())
			}
			return
		}
	}()
	tokenRequest := TokenRequest{
		ProductName: r.FormValue("productName"),
		UserName:    r.FormValue("userName"),
		Password:    r.FormValue("password"),
	}
	errTemplate, err := validateTokenRequest(&tokenRequest)
	if err != nil {
		loggerWithField(r).Error(errTemplate.Message)
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(errTemplate)
		if err != nil {
			loggerWithField(r).Fatal(err.Error())
		}
		return
	}
	result, err := ProductConnector.TokenPermission(&tokenRequest)
	if err != nil {
		loggerWithField(r).Error(err.Error())
		if !result {
			errInfo := ErrorTemplate{
				ErrorCode: http.StatusNotAcceptable,
				Message:   err.Error(),
			}
			loggerWithField(r).Warning(errInfo.Message)
			err := json.NewEncoder(w).Encode(errInfo)
			if err != nil {
				loggerWithField(r).Fatal(err.Error())
			}
		}
		return
	}

	// create a token. default expiry date is 12 month
	now := time.Now()
	//TODO: read the expiry  month from config or cli
	expiryDate := now.AddDate(0, 12, 0)

	jtiRaw := make([]byte, 128)
	_, err = rand.Read(jtiRaw)
	jti := replaceSlashesAndPlus(base64.StdEncoding.EncodeToString(jtiRaw))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errInfo := ErrorTemplate{
			ErrorCode: http.StatusInternalServerError,
			Message:   err.Error(),
		}
		loggerWithField(r).Error(errInfo.Message)
		err := json.NewEncoder(w).Encode(errInfo)
		if err != nil {
			loggerWithField(r).Fatal(err.Error())
		}
		return
	}
	claims := jwt.MapClaims{
		"iss": issuer,
		"sub": tokenRequest.ProductName,
		"exp": expiryDate.Unix(),
		"iat": now.Unix(),
		"jti": jti,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString([]byte(signature))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		errInfo := ErrorTemplate{
			ErrorCode: http.StatusUnauthorized,
			Message:   err.Error(),
		}
		loggerWithField(r).Error(errInfo.Message)
		err := json.NewEncoder(w).Encode(errInfo)
		if err != nil {
			loggerWithField(r).Fatal(err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, tokenString)
}

func replaceSlashesAndPlus(str string) string {
	str = strings.Replace(str, "/", "0", -1)
	str = strings.Replace(str, "+", "0", -1)
	return str
}

func ValidateJsonWebToken(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			errInfo := ErrorTemplate{
				ErrorCode: http.StatusUnauthorized,
				Message:   "Unauthorized: No Authorization Header",
			}
			loggerWithField(r).Error(errInfo.Message)
			err := json.NewEncoder(w).Encode(errInfo)
			if err != nil {
				loggerWithField(r).Fatal(err.Error())
			}
			return
		}
		bearerPosition := strings.Index(authorizationHeader, "Bearer")
		if bearerPosition < 0 {
			w.WriteHeader(http.StatusUnauthorized)
			errInfo := ErrorTemplate{
				ErrorCode: http.StatusUnauthorized,
				Message:   "Unauthorized: No Bearer on Authorization Header",
			}
			loggerWithField(r).Error(errInfo.Message)
			err := json.NewEncoder(w).Encode(errInfo)
			if err != nil {
				loggerWithField(r).Fatal(err.Error())
			}
			return
		}

		jsonWebTokenString := strings.TrimSpace(strings.TrimPrefix(authorizationHeader, "Bearer"))

		token, err := jwt.Parse(jsonWebTokenString, func(token *jwt.Token) (interface{}, error) {

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(signature), nil
		})
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			errInfo := ErrorTemplate{
				ErrorCode: http.StatusUnauthorized,
				Message:   "Unauthorized: Wrong JWT alg.",
			}
			loggerWithField(r).Error(errInfo.Message)
			err := json.NewEncoder(w).Encode(errInfo)
			if err != nil {
				loggerWithField(r).Fatal(err.Error())
			}
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			errInfo := ErrorTemplate{
				ErrorCode: http.StatusUnauthorized,
				Message:   "Unauthorized: Wrong JWT claims.",
			}
			loggerWithField(r).Error(errInfo.Message)
			err := json.NewEncoder(w).Encode(errInfo)
			if err != nil {
				loggerWithField(r).Fatal(err.Error())
			}
			return
		}

		ok = claims.VerifyExpiresAt(time.Now().Unix(), true)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			errInfo := ErrorTemplate{
				ErrorCode: http.StatusUnauthorized,
				Message:   "Unauthorized: JWT has expired.",
			}
			loggerWithField(r).Error(errInfo.Message)
			err := json.NewEncoder(w).Encode(errInfo)
			if err != nil {
				loggerWithField(r).Fatal(err.Error())
			}
			return
		}
		ok = claims.VerifyIssuer(issuer, true)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			errInfo := ErrorTemplate{
				ErrorCode: http.StatusUnauthorized,
				//TODO: for all response do not explain why token is not valid. only display error which will not provide
				//information about token. log the errors
				Message: "Unauthorized: JWT has wrong issuers.",
			}
			loggerWithField(r).Error(errInfo.Message)
			err := json.NewEncoder(w).Encode(errInfo)
			if err != nil {
				loggerWithField(r).Fatal(err.Error())
			}
			return
		}
		productName := claims["sub"].(string) //Type assertion neeeded because it is of type interface{}
		if productName == "" {
			w.WriteHeader(http.StatusUnauthorized)
			errInfo := ErrorTemplate{
				ErrorCode: http.StatusUnauthorized,
				Message:   "Unauthorized: JWT does not have a valid sub.",
			}
			loggerWithField(r).Error(errInfo.Message)
			err := json.NewEncoder(w).Encode(errInfo)
			if err != nil {
				loggerWithField(r).Fatal(err.Error())
			}
			return

		}
		jsonTokenI := claims["jti"].(string)
		if jsonTokenI == "" {
			w.WriteHeader(http.StatusUnauthorized)
			errInfo := ErrorTemplate{
				ErrorCode: http.StatusUnauthorized,
				Message:   "Unauthorized: JWT does not have a valid jti",
			}
			loggerWithField(r).Error(errInfo.Message)
			err := json.NewEncoder(w).Encode(errInfo)
			if err != nil {
				loggerWithField(r).Fatal(err.Error())
			}
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, Jti, jsonTokenI)
		ctx = context.WithValue(ctx, productKey, productName)
		//move to the next middleware in the middleware chain
		r = r.WithContext(ctx)
		f.ServeHTTP(w, r)
	}
}

func validateTokenRequest(tokenR *TokenRequest) (ErrorTemplate, error) {
	if strings.TrimSpace(tokenR.UserName) == "" {
		return ErrorTemplate{
			ErrorCode: http.StatusBadRequest,
			Message:   "Missing userName field",
		}, errors.New("missing userName field")
	}
	if strings.TrimSpace(tokenR.ProductName) == "" {
		return ErrorTemplate{
			ErrorCode: http.StatusBadRequest,
			Message:   "Missing productName field",
		}, errors.New("missing productName field")
	}
	if strings.TrimSpace(tokenR.Password) == "" {
		return ErrorTemplate{
			ErrorCode: http.StatusBadRequest,
			Message:   "Missing password field",
		}, errors.New("missing password field")
	}
	return ErrorTemplate{}, nil
}
