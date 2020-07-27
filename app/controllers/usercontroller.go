package controllers

import (
	"encoding/json"
	"fmt"
	"gopattern/app/helpers"
	"gopattern/app/models"
	"io/ioutil"
	"net/http"
	"strings"
	"os"

	"github.com/gorilla/context"
)

// Register a new user
func (app *App) Register(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{"Status": "Success", "Message": "Register succesfully"}
	user := &models.User{}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		helpers.ERROR(w, http.StatusBadRequest, err)
		return
	}

	err = json.Unmarshal(body, &user)
	if err != nil {
		helpers.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// Validate the user
	err = user.Validate("register")
	if err != nil {
		helpers.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// Check the user
	checkUser, _ := user.GetUserByEmail(app.DB)
	if checkUser != nil {
		response["Status"] = "Error"
		response["Message"] = "User already registered"
		helpers.JSON(w, http.StatusUnauthorized, response)
		return
	}

	_, err = user.Register(app.DB)
	if err != nil {
		helpers.ERROR(w, http.StatusBadRequest, err)
		return
	}

	helpers.JSON(w, http.StatusCreated, response)
	return
}

// Login a user
func (app *App) Login(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{"Status": "Success", "Message": "Login Success"}
	user := &models.User{}
	role := &models.Role{}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		helpers.ERROR(w, http.StatusBadRequest, err)
		return
	}

	err = json.Unmarshal(body, &user)
	if err != nil {
		helpers.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// Validate user
	err = user.Validate("login")
	if err != nil {
		helpers.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// Check the user
	checkUser, _ := user.GetUserByEmail(app.DB)
	if checkUser != nil {
		// Check Password Hash
		err = user.CheckHashedPassword(checkUser.Password, user.Password)
		if err != nil {
			helpers.ERROR(w, http.StatusBadRequest, err)
			return
		}

		// Get Role for user
		makeIDtoString := fmt.Sprint(checkUser.RoleID)
		role, err := role.GetRoleByID(makeIDtoString, app.DB)
		if err != nil {
			helpers.ERROR(w, http.StatusBadRequest, err)
			return
		}

		token, err := helpers.EncodeAuthToken(checkUser.ID, checkUser.Name, checkUser.Email, role.Name)
		if err != nil {
			helpers.ERROR(w, http.StatusBadRequest, err)
			return
		}
		response["Token"] = token
		helpers.JSON(w, http.StatusOK, response)
		return
	}

	response["Status"] = "Error"
	response["Message"] = "User not found"
	helpers.JSON(w, http.StatusNotFound, response)
	return
}

// GetAllUsers getting all users
func (app *App) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{"Status": "Success", "Message": "Users List"}
	user := &models.UserJSON{}

	users, err := user.GetUsers(app.DB)
	if err != nil {
		helpers.ERROR(w, http.StatusBadRequest, err)
		return
	}

	response["Data"] = users
	helpers.JSON(w, http.StatusOK, response)

	return
}

// GetOneUser getting one user
func (app *App) GetOneUser(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{"Status": "Success", "Message": "User Detail"}
	user := &models.UserJSON{}

	userIDFromToken := fmt.Sprint(context.Get(r, "UserID"))
	userData, err := user.GetUser(userIDFromToken, app.DB)
	if err != nil {
		helpers.ERROR(w, http.StatusBadRequest, err)
		return
	}

	response["Data"] = userData
	helpers.JSON(w, http.StatusOK, response)
	return
}

// UploadImage user
func (app *App) UploadUserImage(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{"Status": "Success", "Message": "Image Uploaded"}
	responseFail := map[string]interface{}{"Status": "Error", "Message": "File is required"}
	responseValidation := map[string]interface{}{"Status": "Error", "Message": "The file must be png, jpeg, or jpg"}

	// Update the user
	user := &models.UserJSON{}
	userIDFromContext := fmt.Sprint(context.Get(r, "UserID"))
	if err := app.DB.Debug().Table("users").Preload("Role").Where("id = ?", userIDFromContext).First(&user).Error; err != nil {
		helpers.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// Parse input to type multipart/form-data
	// Set the maximum file size
	r.ParseMultipartForm(10 << 20)

	// Retreive file from posted form-data
	file, handler, err := r.FormFile("file")
	if err != nil {
		helpers.ERROR(w, http.StatusBadRequest, err)
		return
	}
	defer file.Close()

	if file == nil {
		helpers.JSON(w, http.StatusBadRequest, responseFail)
		return
	}

	// Get header type
	headerType := handler.Header["Content-Type"][0]
	headerTypesArray := []string{"image/png", "image/jpeg", "image/jpg"}
	headerTypes := map[string]string{}
	for _, header := range headerTypesArray {
		headerTypes[header] = header
	}

	// Check the type of the file
	if headerType != headerTypes[headerType] {
		helpers.JSON(w, http.StatusUnprocessableEntity, responseValidation)
		return
	}

	// Write temporary file in local
	getFileExtension := strings.Split(headerType, "/")[1]
	tempFile, err := ioutil.TempFile("public/images", "images-*."+getFileExtension)
	if err != nil {
		helpers.ERROR(w, http.StatusBadRequest, err)
		return
	}
	defer tempFile.Close()

	// Get The file bytes
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		helpers.ERROR(w, http.StatusBadRequest, err)
		return
	}
	// Write the file
	tempFile.Write(fileBytes)

	// Remove Previous image (if exists)
	if user.ImageURL != "" {
		getFileNameOnly := strings.Split(user.ImageURL, "/")[3]
		err := os.Remove("public/images/" + getFileNameOnly)
		if err != nil {
			helpers.ERROR(w, http.StatusBadRequest, err)
			return
		}
		user.ImageURL = r.Host + "/" + strings.ReplaceAll(tempFile.Name(), "\\", "/")
	} else {
		// Upload file as usual
		user.ImageURL = r.Host + "/" + strings.ReplaceAll(tempFile.Name(), "\\", "/")
	}

	// Save the user
	app.DB.Save(&user)

	response["Data"] = user
	helpers.JSON(w, http.StatusOK, response)
}

// DeleteImage user
func (app *App) DeleteUser(w http.ResponseWriter, r *http.Request) {

}

